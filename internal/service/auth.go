package service

import (
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	stdmail "net/mail"
	"strings"
	"time"

	"github.com/azmiagr/cakra-hackathon/entity"
	"github.com/azmiagr/cakra-hackathon/internal/repository"
	"github.com/azmiagr/cakra-hackathon/model"
	"github.com/azmiagr/cakra-hackathon/pkg/bcrypt"
	"github.com/azmiagr/cakra-hackathon/pkg/config"
	constants "github.com/azmiagr/cakra-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/cakra-hackathon/pkg/errors"
	"github.com/azmiagr/cakra-hackathon/pkg/jwt"
	"github.com/azmiagr/cakra-hackathon/pkg/mail"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IAuthService interface {
	Register(req model.RegisterRequest) (*model.RegistrationResult, error)
	VerifyRegistrationOTP(sessionToken string, req model.VerifyRegistrationOTPRequest) (*model.RegistrationResult, error)
	ResendRegistrationOTP(sessionToken string) error
	SetRegistrationPassword(sessionToken string, req model.SetRegistrationPasswordRequest) (*model.CompleteRegistrationResult, error)
	RequestPasswordReset(req model.ForgotPasswordRequest) (*model.PasswordResetResult, error)
	VerifyPasswordResetOTP(sessionToken string, req model.VerifyPasswordResetOTPRequest) (*model.PasswordResetResult, error)
	ResendPasswordResetOTP(sessionToken string) error
	SetPasswordReset(sessionToken string, req model.SetPasswordResetRequest) error
}

type AuthService struct {
	db                *gorm.DB
	userRepo          repository.IUserRepository
	otpRepo           repository.IOtpRepository
	sessionRepo       repository.IRegistrationSessionRepository
	passwordResetRepo repository.IPasswordResetRepository
	roleRepo          repository.IRoleRepository
	bcrypt            bcrypt.Interface
	jwt               jwt.Interface
	mailer            mail.Interface
	registrationConf  config.RegistrationConfig
}

func NewAuthService(
	db *gorm.DB,
	userRepo repository.IUserRepository,
	otpRepo repository.IOtpRepository,
	sessionRepo repository.IRegistrationSessionRepository,
	passwordResetRepo repository.IPasswordResetRepository,
	roleRepo repository.IRoleRepository,
	bcryptAuth bcrypt.Interface,
	jwtAuth jwt.Interface,
	mailer mail.Interface,
	registrationConf config.RegistrationConfig,
) IAuthService {
	return &AuthService{
		db:                db,
		userRepo:          userRepo,
		otpRepo:           otpRepo,
		sessionRepo:       sessionRepo,
		passwordResetRepo: passwordResetRepo,
		roleRepo:          roleRepo,
		bcrypt:            bcryptAuth,
		jwt:               jwtAuth,
		mailer:            mailer,
		registrationConf:  registrationConf,
	}
}

func (s *AuthService) Register(req model.RegisterRequest) (*model.RegistrationResult, error) {
	fullName, email, err := normalizeRegistrationInput(req)
	if err != nil {
		return nil, err
	}

	rawSessionToken, sessionHash, err := generateSessionToken()
	if err != nil {
		return nil, appErrors.InternalServer("gagal memulai registrasi")
	}

	rawOTP, err := generateOTP()
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat OTP")
	}

	otpHash, err := s.bcrypt.GenerateFromPassword(rawOTP)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat OTP")
	}

	now := time.Now().UTC()
	otpExpiresAt := now.Add(s.registrationConf.OTPTTL)

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, appErrors.InternalServer("gagal memulai registrasi")
	}
	defer tx.Rollback()

	role, err := s.roleRepo.GetRoleByName(tx, constants.RoleUser)
	if err != nil {
		return nil, appErrors.InternalServer("role pengguna tidak tersedia")
	}

	user, err := s.userRepo.GetUser(tx, model.GetUserParam{Email: email})
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		user = &entity.User{
			UserID:   uuid.New(),
			RoleID:   role.RoleID,
			FullName: fullName,
			Email:    email,
			Status:   "inactive",
		}
		err = s.userRepo.CreateUser(tx, user)
		if err != nil {
			return nil, appErrors.InternalServer("gagal membuat akun")
		}
	case err != nil:
		return nil, appErrors.InternalServer("gagal memeriksa email")
	case user.Status == "active":
		return nil, appErrors.Conflict("email sudah terdaftar")
	default:
		user.FullName = fullName
		user.RoleID = role.RoleID
		err = s.userRepo.UpdateUser(tx, user)
		if err != nil {
			return nil, appErrors.InternalServer("gagal memperbarui registrasi")
		}
	}

	session, sessionErr := s.sessionRepo.GetByUserIDForUpdate(tx, user.UserID)
	if sessionErr != nil && !errors.Is(sessionErr, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memeriksa session registrasi")
	}

	existingOTP, otpErr := s.otpRepo.GetOtpByUserIDForUpdate(tx, user.UserID)
	if otpErr == nil && existingOTP.LastSentAt != nil && now.Sub(*existingOTP.LastSentAt) < s.registrationConf.ResendCooldown {
		return nil, appErrors.TooManyRequests("OTP baru dapat dikirim setelah cooldown selesai")
	}
	if otpErr != nil && !errors.Is(otpErr, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memeriksa OTP")
	}

	if errors.Is(otpErr, gorm.ErrRecordNotFound) {
		existingOTP = &entity.OtpCode{
			OtpID:        uuid.New(),
			UserID:       user.UserID,
			CodeHash:     otpHash,
			ExpiresAt:    otpExpiresAt,
			AttemptCount: 0,
			LastSentAt:   &now,
		}
		err = s.otpRepo.CreateOtp(tx, existingOTP)
		if err != nil {
			return nil, appErrors.InternalServer("gagal menyimpan OTP")
		}
	} else {
		existingOTP.CodeHash = otpHash
		existingOTP.ExpiresAt = otpExpiresAt
		existingOTP.AttemptCount = 0
		existingOTP.LastSentAt = &now
		err = s.otpRepo.UpdateOtp(tx, existingOTP)
		if err != nil {
			return nil, appErrors.InternalServer("gagal memperbarui OTP")
		}
	}

	if errors.Is(sessionErr, gorm.ErrRecordNotFound) {
		session = &entity.RegistrationSession{
			RegistrationSessionID: uuid.New(),
			UserID:                user.UserID,
			TokenHash:             sessionHash,
			Stage:                 constants.RegistrationSessionOTP,
			ExpiresAt:             now.Add(s.registrationConf.SessionTTL),
		}
		err = s.sessionRepo.CreateRegistrationSession(tx, session)
	} else {
		session.TokenHash = sessionHash
		session.Stage = constants.RegistrationSessionOTP
		session.ExpiresAt = now.Add(s.registrationConf.SessionTTL)
		err = s.sessionRepo.UpdateRegistrationSession(tx, session)
	}
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat session registrasi")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan registrasi")
	}

	result := &model.RegistrationResult{
		SessionToken: rawSessionToken,
		OTPExpiresAt: otpExpiresAt,
	}

	err = s.mailer.SendRegistrationOTP(user.Email, user.FullName, rawOTP, int(s.registrationConf.OTPTTL.Minutes()))
	if err != nil {
		s.clearOTPDeliveryTimestamp(user.UserID)
		return result, appErrors.ServiceUnavailable("gagal mengirim OTP, silakan kirim ulang")
	}

	return result, nil
}

func (s *AuthService) VerifyRegistrationOTP(sessionToken string, req model.VerifyRegistrationOTPRequest) (*model.RegistrationResult, error) {
	if len(req.OTP) != 6 || !isSixDigitOTP(req.OTP) {
		return nil, appErrors.BadRequest("OTP harus terdiri dari 6 digit")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, appErrors.InternalServer("gagal memverifikasi OTP")
	}
	defer tx.Rollback()

	session, err := s.getSessionForStage(tx, sessionToken, constants.RegistrationSessionOTP)
	if err != nil {
		return nil, err
	}

	otp, err := s.otpRepo.GetOtpByUserIDForUpdate(tx, session.UserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.BadRequest("OTP tidak tersedia, silakan kirim ulang")
	}
	if err != nil {
		return nil, appErrors.InternalServer("gagal memverifikasi OTP")
	}
	if time.Now().UTC().After(otp.ExpiresAt) {
		return nil, appErrors.BadRequest("OTP sudah kedaluwarsa, silakan kirim ulang")
	}

	err = s.bcrypt.CompareAndHashPassword(otp.CodeHash, req.OTP)
	if err != nil {
		otp.AttemptCount++
		if otp.AttemptCount >= s.registrationConf.MaxOTPAttempts {
			session.Stage = constants.RegistrationSessionComplete
			updateErr := s.sessionRepo.UpdateRegistrationSession(tx, session)
			if updateErr != nil {
				return nil, appErrors.InternalServer("gagal memperbarui session registrasi")
			}
		}
		updateErr := s.otpRepo.UpdateOtp(tx, otp)
		if updateErr != nil {
			return nil, appErrors.InternalServer("gagal memperbarui OTP")
		}

		commitErr := tx.Commit().Error
		if commitErr != nil {
			return nil, appErrors.InternalServer("gagal memverifikasi OTP")
		}
		if otp.AttemptCount >= s.registrationConf.MaxOTPAttempts {
			return nil, appErrors.TooManyRequests("batas percobaan OTP telah tercapai")
		}

		return nil, appErrors.BadRequest("OTP tidak valid")
	}

	rawSessionToken, sessionHash, err := generateSessionToken()
	if err != nil {
		return nil, appErrors.InternalServer("gagal melanjutkan registrasi")
	}

	session.TokenHash = sessionHash
	session.Stage = constants.RegistrationSessionPassword
	session.ExpiresAt = time.Now().UTC().Add(s.registrationConf.PasswordTTL)
	err = s.sessionRepo.UpdateRegistrationSession(tx, session)
	if err != nil {
		return nil, appErrors.InternalServer("gagal memperbarui session registrasi")
	}

	err = s.otpRepo.DeleteOtpByUserID(tx, session.UserID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menghapus OTP")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal memverifikasi OTP")
	}

	return &model.RegistrationResult{
		SessionToken: rawSessionToken,
	}, nil
}

func (s *AuthService) ResendRegistrationOTP(sessionToken string) error {
	rawOTP, err := generateOTP()
	if err != nil {
		return appErrors.InternalServer("gagal membuat OTP")
	}

	otpHash, err := s.bcrypt.GenerateFromPassword(rawOTP)
	if err != nil {
		return appErrors.InternalServer("gagal membuat OTP")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return appErrors.InternalServer("gagal mengirim ulang OTP")
	}
	defer tx.Rollback()

	session, err := s.getSessionForStage(tx, sessionToken, constants.RegistrationSessionOTP)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetUser(tx, model.GetUserParam{UserID: session.UserID})
	if err != nil {
		return appErrors.InternalServer("gagal memeriksa akun")
	}

	otp, err := s.otpRepo.GetOtpByUserIDForUpdate(tx, session.UserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.BadRequest("OTP tidak tersedia, mulai registrasi kembali")
	}
	if err != nil {
		return appErrors.InternalServer("gagal memeriksa OTP")
	}

	now := time.Now().UTC()
	if otp.LastSentAt != nil && now.Sub(*otp.LastSentAt) < s.registrationConf.ResendCooldown {
		return appErrors.TooManyRequests("OTP baru dapat dikirim setelah cooldown selesai")
	}

	otp.CodeHash = otpHash
	otp.ExpiresAt = now.Add(s.registrationConf.OTPTTL)
	otp.AttemptCount = 0
	otp.LastSentAt = &now
	err = s.otpRepo.UpdateOtp(tx, otp)
	if err != nil {
		return appErrors.InternalServer("gagal memperbarui OTP")
	}

	err = tx.Commit().Error
	if err != nil {
		return appErrors.InternalServer("gagal mengirim ulang OTP")
	}

	err = s.mailer.SendRegistrationOTP(user.Email, user.FullName, rawOTP, int(s.registrationConf.OTPTTL.Minutes()))
	if err != nil {
		s.clearOTPDeliveryTimestamp(session.UserID)
		return appErrors.ServiceUnavailable("gagal mengirim OTP, silakan coba lagi")
	}

	return nil
}

func (s *AuthService) SetRegistrationPassword(sessionToken string, req model.SetRegistrationPasswordRequest) (*model.CompleteRegistrationResult, error) {
	if req.Password != req.ConfirmPassword {
		return nil, appErrors.BadRequest("konfirmasi password tidak sesuai")
	}
	if len(req.Password) < 8 || len(req.Password) > 72 {
		return nil, appErrors.BadRequest("password harus terdiri dari 8 sampai 72 karakter")
	}

	passwordHash, err := s.bcrypt.GenerateFromPassword(req.Password)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan password")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, appErrors.InternalServer("gagal mengaktifkan akun")
	}
	defer tx.Rollback()

	session, err := s.getSessionForStage(tx, sessionToken, constants.RegistrationSessionPassword)
	if err != nil {
		return nil, err
	}

	err = s.userRepo.ActivateUser(tx, session.UserID, passwordHash)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengaktifkan akun")
	}

	session.Stage = constants.RegistrationSessionComplete
	err = s.sessionRepo.UpdateRegistrationSession(tx, session)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyelesaikan registrasi")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengaktifkan akun")
	}

	accessToken, err := s.jwt.CreateJWTToken(session.UserID, constants.RoleUser, 0)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat access token")
	}
	return &model.CompleteRegistrationResult{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}, nil
}

func (s *AuthService) RequestPasswordReset(req model.ForgotPasswordRequest) (*model.PasswordResetResult, error) {
	email, err := normalizeEmail(req.Email)
	if err != nil {
		return nil, err
	}

	rawSessionToken, sessionHash, err := generateSessionToken()
	if err != nil {
		return nil, appErrors.InternalServer("gagal memulai reset password")
	}
	result := &model.PasswordResetResult{
		SessionToken: rawSessionToken,
	}

	user, err := s.userRepo.GetUser(s.db, model.GetUserParam{Email: email})
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && user.Status != "active") {
		return result, nil
	}
	if err != nil {
		return nil, appErrors.InternalServer("gagal memproses reset password")
	}

	rawOTP, err := generateOTP()
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat OTP")
	}

	otpHash, err := s.bcrypt.GenerateFromPassword(rawOTP)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat OTP")
	}

	now := time.Now().UTC()
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, appErrors.InternalServer("gagal memulai reset password")
	}
	defer tx.Rollback()

	session, sessionErr := s.passwordResetRepo.GetByUserIDForUpdate(tx, user.UserID)
	if sessionErr != nil && !errors.Is(sessionErr, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memproses reset password")
	}
	if sessionErr == nil && session.Stage == constants.PasswordResetSessionOTP && session.LastSentAt != nil && now.Sub(*session.LastSentAt) < s.registrationConf.ResendCooldown {
		session.TokenHash = sessionHash
		session.ExpiresAt = now.Add(s.registrationConf.SessionTTL)
		err := s.passwordResetRepo.Update(tx, session)
		if err != nil {
			return nil, appErrors.InternalServer("gagal memproses reset password")
		}
		err = tx.Commit().Error
		if err != nil {
			return nil, appErrors.InternalServer("gagal memproses reset password")
		}
		return result, nil
	}

	otpExpiresAt := now.Add(s.registrationConf.OTPTTL)
	if errors.Is(sessionErr, gorm.ErrRecordNotFound) {
		session = &entity.PasswordReset{
			PasswordResetID: uuid.New(),
			UserID:          user.UserID,
			TokenHash:       sessionHash,
			Stage:           constants.PasswordResetSessionOTP,
			OTPCodeHash:     &otpHash,
			OTPExpiresAt:    &otpExpiresAt,
			LastSentAt:      &now,
			ExpiresAt:       now.Add(s.registrationConf.SessionTTL),
		}
		err = s.passwordResetRepo.Create(tx, session)
	} else {
		session.TokenHash = sessionHash
		session.Stage = constants.PasswordResetSessionOTP
		session.OTPCodeHash = &otpHash
		session.OTPExpiresAt = &otpExpiresAt
		session.AttemptCount = 0
		session.LastSentAt = &now
		session.ExpiresAt = now.Add(s.registrationConf.SessionTTL)
		err = s.passwordResetRepo.Update(tx, session)
	}
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan reset password")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan reset password")
	}

	err = s.mailer.SendPasswordResetOTP(user.Email, user.FullName, rawOTP, int(s.registrationConf.OTPTTL.Minutes()))
	if err != nil {
		s.clearPasswordResetDeliveryTimestamp(user.UserID)
	}
	return result, nil
}

func (s *AuthService) VerifyPasswordResetOTP(sessionToken string, req model.VerifyPasswordResetOTPRequest) (*model.PasswordResetResult, error) {
	if len(req.OTP) != 6 || !isSixDigitOTP(req.OTP) {
		return nil, appErrors.BadRequest("OTP harus terdiri dari 6 digit")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, appErrors.InternalServer("gagal memverifikasi OTP")
	}
	defer tx.Rollback()

	session, err := s.getPasswordResetSessionForStage(tx, sessionToken, constants.PasswordResetSessionOTP)
	if err != nil {
		return nil, err
	}
	if session.OTPCodeHash == nil || session.OTPExpiresAt == nil || time.Now().UTC().After(*session.OTPExpiresAt) {
		return nil, appErrors.BadRequest("OTP tidak valid atau sudah kedaluwarsa")
	}

	err = s.bcrypt.CompareAndHashPassword(*session.OTPCodeHash, req.OTP)
	if err != nil {
		session.AttemptCount++
		if session.AttemptCount >= s.registrationConf.MaxOTPAttempts {
			session.Stage = constants.PasswordResetSessionComplete
		}

		updateErr := s.passwordResetRepo.Update(tx, session)
		if updateErr != nil {
			return nil, appErrors.InternalServer("gagal memverifikasi OTP")
		}

		commitErr := tx.Commit().Error
		if commitErr != nil {
			return nil, appErrors.InternalServer("gagal memverifikasi OTP")
		}
		if session.AttemptCount >= s.registrationConf.MaxOTPAttempts {
			return nil, appErrors.TooManyRequests("batas percobaan OTP telah tercapai")
		}

		return nil, appErrors.BadRequest("OTP tidak valid")
	}

	rawSessionToken, sessionHash, err := generateSessionToken()
	if err != nil {
		return nil, appErrors.InternalServer("gagal melanjutkan reset password")
	}

	session.TokenHash = sessionHash
	session.Stage = constants.PasswordResetSessionPassword
	session.OTPCodeHash = nil
	session.OTPExpiresAt = nil
	session.ExpiresAt = time.Now().UTC().Add(s.registrationConf.PasswordTTL)
	err = s.passwordResetRepo.Update(tx, session)
	if err != nil {
		return nil, appErrors.InternalServer("gagal memverifikasi OTP")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal memverifikasi OTP")
	}

	return &model.PasswordResetResult{SessionToken: rawSessionToken}, nil
}

func (s *AuthService) ResendPasswordResetOTP(sessionToken string) error {
	if strings.TrimSpace(sessionToken) == "" {
		return nil
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return appErrors.InternalServer("gagal mengirim ulang OTP")
	}
	defer tx.Rollback()

	session, err := s.passwordResetRepo.GetByTokenHashForUpdate(tx, hashSessionToken(sessionToken))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return appErrors.InternalServer("gagal mengirim ulang OTP")
	}
	if session.Stage != constants.PasswordResetSessionOTP || time.Now().UTC().After(session.ExpiresAt) {
		return nil
	}

	now := time.Now().UTC()
	if session.LastSentAt != nil && now.Sub(*session.LastSentAt) < s.registrationConf.ResendCooldown {
		return nil
	}

	user, err := s.userRepo.GetUser(tx, model.GetUserParam{UserID: session.UserID})
	if err != nil || user.Status != "active" {
		return nil
	}

	rawOTP, err := generateOTP()
	if err != nil {
		return appErrors.InternalServer("gagal membuat OTP")
	}

	otpHash, err := s.bcrypt.GenerateFromPassword(rawOTP)
	if err != nil {
		return appErrors.InternalServer("gagal membuat OTP")
	}

	otpExpiresAt := now.Add(s.registrationConf.OTPTTL)
	session.OTPCodeHash = &otpHash
	session.OTPExpiresAt = &otpExpiresAt
	session.AttemptCount = 0
	session.LastSentAt = &now
	err = s.passwordResetRepo.Update(tx, session)
	if err != nil {
		return appErrors.InternalServer("gagal mengirim ulang OTP")
	}

	err = tx.Commit().Error
	if err != nil {
		return appErrors.InternalServer("gagal mengirim ulang OTP")
	}

	err = s.mailer.SendPasswordResetOTP(user.Email, user.FullName, rawOTP, int(s.registrationConf.OTPTTL.Minutes()))
	if err != nil {
		s.clearPasswordResetDeliveryTimestamp(user.UserID)
	}
	return nil
}

func (s *AuthService) SetPasswordReset(sessionToken string, req model.SetPasswordResetRequest) error {
	if req.Password != req.ConfirmPassword {
		return appErrors.BadRequest("konfirmasi password tidak sesuai")
	}
	if len(req.Password) < 8 || len(req.Password) > 72 {
		return appErrors.BadRequest("password harus terdiri dari 8 sampai 72 karakter")
	}

	passwordHash, err := s.bcrypt.GenerateFromPassword(req.Password)
	if err != nil {
		return appErrors.InternalServer("gagal menyimpan password")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return appErrors.InternalServer("gagal mengubah password")
	}
	defer tx.Rollback()

	session, err := s.getPasswordResetSessionForStage(tx, sessionToken, constants.PasswordResetSessionPassword)
	if err != nil {
		return err
	}

	err = s.userRepo.UpdatePasswordAndIncrementSessionVersion(tx, session.UserID, passwordHash)
	if err != nil {
		return appErrors.InternalServer("gagal mengubah password")
	}

	session.Stage = constants.PasswordResetSessionComplete
	err = s.passwordResetRepo.Update(tx, session)
	if err != nil {
		return appErrors.InternalServer("gagal menyelesaikan reset password")
	}

	err = tx.Commit().Error
	if err != nil {
		return appErrors.InternalServer("gagal mengubah password")
	}

	return nil
}

func (s *AuthService) getSessionForStage(tx *gorm.DB, sessionToken, stage string) (*entity.RegistrationSession, error) {
	if strings.TrimSpace(sessionToken) == "" {
		return nil, appErrors.Unauthorized("session registrasi tidak tersedia")
	}

	session, err := s.sessionRepo.GetByTokenHashForUpdate(tx, hashSessionToken(sessionToken))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.Unauthorized("session registrasi tidak valid")
	}
	if err != nil {
		return nil, appErrors.InternalServer("gagal memeriksa session registrasi")
	}
	if session.Stage != stage || time.Now().UTC().After(session.ExpiresAt) {
		return nil, appErrors.Unauthorized("session registrasi sudah tidak berlaku")
	}
	return session, nil
}

func (s *AuthService) getPasswordResetSessionForStage(tx *gorm.DB, sessionToken, stage string) (*entity.PasswordReset, error) {
	if strings.TrimSpace(sessionToken) == "" {
		return nil, appErrors.BadRequest("OTP tidak valid")
	}

	session, err := s.passwordResetRepo.GetByTokenHashForUpdate(tx, hashSessionToken(sessionToken))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.BadRequest("OTP tidak valid")
	}
	if err != nil {
		return nil, appErrors.InternalServer("gagal memproses reset password")
	}
	if session.Stage != stage || time.Now().UTC().After(session.ExpiresAt) {
		return nil, appErrors.BadRequest("session reset password sudah tidak berlaku")
	}

	return session, nil
}

func (s *AuthService) clearOTPDeliveryTimestamp(userID uuid.UUID) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return
	}
	defer tx.Rollback()

	err := s.otpRepo.ClearOtpSentAt(tx, userID)
	if err != nil {
		return
	}
	_ = tx.Commit().Error
}

func (s *AuthService) clearPasswordResetDeliveryTimestamp(userID uuid.UUID) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return
	}
	defer tx.Rollback()

	session, err := s.passwordResetRepo.GetByUserIDForUpdate(tx, userID)
	if err != nil {
		return
	}
	session.LastSentAt = nil
	if err := s.passwordResetRepo.Update(tx, session); err != nil {
		return
	}
	_ = tx.Commit().Error
}

func normalizeRegistrationInput(req model.RegisterRequest) (string, string, error) {
	fullName := strings.TrimSpace(req.FullName)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if len(fullName) < 3 || len(fullName) > 150 {
		return "", "", appErrors.BadRequest("nama lengkap harus terdiri dari 3 sampai 150 karakter")
	}

	parsedEmail, err := stdmail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email || len(email) > 100 {
		return "", "", appErrors.BadRequest("format email tidak valid")
	}

	return fullName, email, nil
}

func normalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	parsedEmail, err := stdmail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email || len(email) > 100 {
		return "", appErrors.BadRequest("format email tidak valid")
	}
	return email, nil
}

func generateOTP() (string, error) {
	limit := big.NewInt(900000)
	number, err := cryptoRand.Int(cryptoRand.Reader, limit)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", number.Int64()+100000), nil
}

func generateSessionToken() (string, string, error) {
	bytes := make([]byte, 32)
	_, err := cryptoRand.Read(bytes)
	if err != nil {
		return "", "", err
	}
	rawToken := hex.EncodeToString(bytes)

	return rawToken, hashSessionToken(rawToken), nil
}

func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func isSixDigitOTP(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
