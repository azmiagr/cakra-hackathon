package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/azmiagr/cakra-hackathon/entity"
	"github.com/azmiagr/cakra-hackathon/internal/repository"
	"github.com/azmiagr/cakra-hackathon/model"
	constants "github.com/azmiagr/cakra-hackathon/pkg/constant"
	"github.com/azmiagr/cakra-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/cakra-hackathon/pkg/errors"
	"github.com/azmiagr/cakra-hackathon/pkg/supabase"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

const maxXLSXSize = 2 * 1024 * 1024

type IAnalysisService interface {
	Upload(userID uuid.UUID, file *multipart.FileHeader) (*model.AnalysisUploadResponse, error)
	CreateSession(userID, uploadID uuid.UUID, req model.CreateAnalysisSessionRequest) (*model.AnalysisSessionResponse, error)
	GetSession(userID, sessionID uuid.UUID) (*model.AnalysisSessionResponse, error)
	GetHistory(userID uuid.UUID, query model.AnalysisHistoryQuery) (*model.AnalysisHistoryResponse, error)
	CompleteFromAI(sessionID uuid.UUID, req model.AIResultRequest) error
}

type AnalysisService struct {
	db           *gorm.DB
	analysisRepo repository.IAnalysisRepository
	creditRepo   repository.ICreditAccountRepository
	storage      supabase.Interface
}

func NewAnalysisService(analysisRepo repository.IAnalysisRepository, creditRepo repository.ICreditAccountRepository, storage supabase.Interface) IAnalysisService {
	return &AnalysisService{
		db:           mariadb.Connection,
		analysisRepo: analysisRepo,
		creditRepo:   creditRepo,
		storage:      storage,
	}
}

func (s *AnalysisService) Upload(userID uuid.UUID, file *multipart.FileHeader) (*model.AnalysisUploadResponse, error) {
	if file == nil {
		return nil, appErrors.BadRequest("file XLSX wajib diunggah")
	}
	if file.Size <= 0 || file.Size > maxXLSXSize {
		return nil, appErrors.BadRequest("ukuran file XLSX maksimal 2 MB")
	}
	if strings.ToLower(filepath.Ext(file.Filename)) != ".xlsx" {
		return nil, appErrors.BadRequest("file harus berformat XLSX")
	}

	src, err := file.Open()
	if err != nil {
		return nil, appErrors.BadRequest("file XLSX tidak dapat dibaca")
	}
	defer src.Close()

	data, err := io.ReadAll(io.LimitReader(src, maxXLSXSize+1))
	if err != nil || len(data) > maxXLSXSize {
		return nil, appErrors.BadRequest("file XLSX tidak dapat dibaca")
	}

	if len(data) < 4 || !bytes.Equal(data[:4], []byte{'P', 'K', 3, 4}) {
		return nil, appErrors.BadRequest("file bukan XLSX yang valid")
	}

	parsed := parseXLSX(data)
	key := fmt.Sprintf("analysis-uploads/%s/%s.xlsx", userID, uuid.NewString())
	err = s.storage.UploadXLSX(data, key)
	if err != nil {
		return nil, appErrors.ServiceUnavailable("gagal menyimpan file XLSX")
	}

	committed := false
	defer func() {
		if !committed {
			_ = s.storage.DeleteXLSX(key)
		}
	}()

	status := constants.AnalysisUploadValid
	if len(parsed.errors) > 0 {
		status = constants.AnalysisUploadInvalid
	}

	hash := sha256.Sum256(data)
	upload := &entity.AnalysisUpload{AnalysisUploadID: uuid.New(), UserID: userID, OriginalFilename: file.Filename, StorageObjectKey: key, FileSizeBytes: int64(len(data)), SHA256: hex.EncodeToString(hash[:]), Status: status, SKUName: parsed.skuName, ValidRowCount: len(parsed.rows), ErrorRowCount: len(parsed.errors)}
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, appErrors.InternalServer("gagal menyimpan upload")
	}
	defer tx.Rollback()

	if err := s.analysisRepo.CreateUpload(tx, upload); err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan upload")
	}

	for i := range parsed.rows {
		parsed.rows[i].AnalysisUploadID = upload.AnalysisUploadID
	}

	for i := range parsed.errors {
		parsed.errors[i].AnalysisUploadID = upload.AnalysisUploadID
	}

	err = s.analysisRepo.CreateUploadRows(tx, parsed.rows)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan preview upload")
	}

	err = s.analysisRepo.CreateValidationErrors(tx, parsed.errors)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan error validasi")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan upload")
	}

	committed = true
	return uploadResponse(upload, parsed.rows, parsed.errors), nil
}

func (s *AnalysisService) CreateSession(userID, uploadID uuid.UUID, req model.CreateAnalysisSessionRequest) (*model.AnalysisSessionResponse, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, appErrors.InternalServer("gagal memulai analisis")
	}
	defer tx.Rollback()

	upload, err := s.analysisRepo.GetUploadOwned(tx, uploadID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.NotFound("upload tidak ditemukan")
	}
	if err != nil {
		return nil, appErrors.InternalServer("gagal memeriksa upload")
	}
	if upload.Status != constants.AnalysisUploadValid {
		return nil, appErrors.BadRequest("perbaiki data XLSX yang tidak valid sebelum memulai analisis")
	}

	rows, err := s.analysisRepo.ListUploadRows(tx, upload.AnalysisUploadID)
	if err != nil || len(rows) == 0 {
		return nil, appErrors.BadRequest("data penjualan valid tidak tersedia")
	}

	account, err := s.creditRepo.GetByUserIDForUpdate(tx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.BadRequest("akun kredit belum tersedia")
	}
	if err != nil {
		return nil, appErrors.InternalServer("gagal memeriksa kredit")
	}
	if err := s.creditRepo.ReserveAnalysisCredit(tx, userID, 1); err != nil {
		if errors.Is(err, repository.ErrInsufficientCredits) {
			return nil, appErrors.BadRequest("kredit tidak mencukupi")
		}
		return nil, appErrors.InternalServer("gagal mereservasi kredit")
	}

	sku, err := s.analysisRepo.GetOrCreateSKU(tx, &entity.SKU{
		SKUID:  uuid.New(),
		UserID: userID,
		Name:   upload.SKUName,
	})
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan SKU")
	}

	session := &entity.AnalysisSession{
		AnalysisSessionID: uuid.New(),
		UserID:            userID,
		SKUID:             sku.SKUID,
		AnalysisUploadID:  upload.AnalysisUploadID,
		CurrentStock:      req.CurrentStock,
		LeadTimeDays:      req.LeadTimeDays,
		CreditCost:        1,
		Status:            constants.AnalysisSessionPendingAI,
	}
	err = s.analysisRepo.CreateSession(tx, session)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat sesi analisis")
	}

	history := make([]entity.SalesHistory, len(rows))
	payloadRows := make([]model.AISalesHistoryRow, len(rows))
	for i, row := range rows {
		history[i] = entity.SalesHistory{
			SalesHistoryID:    uuid.New(),
			AnalysisSessionID: session.AnalysisSessionID,
			SaleDate:          row.SaleDate,
			QuantitySold:      row.QuantitySold,
			UnitPrice:         row.UnitPrice,
		}
		payloadRows[i] = model.AISalesHistoryRow{
			SaleDate:     row.SaleDate.Format("2006-01-02"),
			QuantitySold: row.QuantitySold,
			UnitPrice:    row.UnitPrice,
		}
	}

	err = s.analysisRepo.CreateSalesHistories(tx, history)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan data penjualan")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat sesi analisis")
	}

	return &model.AnalysisSessionResponse{
		SessionID:        session.AnalysisSessionID,
		Status:           session.Status,
		AvailableCredits: account.Balance - account.ReservedCredits - 1,
		AIPayload: &model.AIAnalysisPayload{
			AnalysisSessionID: session.AnalysisSessionID,
			SKUName:           sku.Name,
			CurrentStock:      session.CurrentStock,
			LeadTimeDays:      session.LeadTimeDays,
			SalesHistory:      payloadRows},
	}, nil
}

func (s *AnalysisService) GetSession(userID, sessionID uuid.UUID) (*model.AnalysisSessionResponse, error) {
	session, sku, result, err := s.analysisRepo.GetSessionOwned(s.db, sessionID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.NotFound("sesi analisis tidak ditemukan")
	}
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil sesi analisis")
	}

	account, err := s.creditRepo.GetByUserID(s.db, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.NotFound("akun kredit tidak ditemukan")
	}
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil kredit")
	}

	rows, err := s.analysisRepo.ListSalesHistories(s.db, sessionID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil data analisis")
	}

	payloadRows := make([]model.AISalesHistoryRow, len(rows))
	for i, row := range rows {
		payloadRows[i] = model.AISalesHistoryRow{
			SaleDate:     row.SaleDate.Format("2006-01-02"),
			QuantitySold: row.QuantitySold,
			UnitPrice:    row.UnitPrice,
		}
	}

	response := &model.AnalysisSessionResponse{
		SessionID:        session.AnalysisSessionID,
		Status:           session.Status,
		AvailableCredits: account.Balance - account.ReservedCredits,
		FailureCode:      session.FailureCode,
		FailureMessage:   session.FailureMessage,
		AIPayload: &model.AIAnalysisPayload{
			AnalysisSessionID: session.AnalysisSessionID,
			SKUName:           sku.Name,
			CurrentStock:      session.CurrentStock,
			LeadTimeDays:      session.LeadTimeDays,
			SalesHistory:      payloadRows},
	}

	if result != nil {
		p50 := []float64{}
		p90 := []float64{}
		_ = json.Unmarshal([]byte(result.ForecastP50), &p50)
		_ = json.Unmarshal([]byte(result.ForecastP90), &p90)
		response.AIPayload = nil
		response.Recommendation = &model.RecommendationResponse{
			DemandCategory:  stringValue(session.DemandCategory),
			ADIValue:        session.ADIValue,
			CVSquaredValue:  session.CVSquaredValue,
			ForecastP50:     p50,
			ForecastP90:     p90,
			ReorderPoint:    result.ReorderPoint,
			ReorderQuantity: result.ReorderQuantity,
			RiskLabel:       result.RiskLabel,
			RiskReason:      result.RiskReason,
			ExplanationText: result.ExplanationText,
		}
		response.Result = buildAnalysisResult(session, sku, rows, result, p50, p90)
	}
	return response, nil
}

func (s *AnalysisService) GetHistory(userID uuid.UUID, query model.AnalysisHistoryQuery) (*model.AnalysisHistoryResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 {
		query.Limit = 8
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.Sort != "oldest" {
		query.Sort = "newest"
	}

	summary, err := s.analysisRepo.GetHistorySummary(s.db, userID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil ringkasan riwayat analisis")
	}
	items, totalItems, err := s.analysisRepo.ListHistory(s.db, userID, query)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil riwayat analisis")
	}

	totalPages := (totalItems + query.Limit - 1) / query.Limit
	return &model.AnalysisHistoryResponse{
		Summary: *summary,
		Items:   items,
		Pagination: model.PaginationResponse{
			Page:       query.Page,
			Limit:      query.Limit,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	}, nil
}

func buildAnalysisResult(session *entity.AnalysisSession, sku *entity.SKU, rows []entity.SalesHistory, result *entity.RecommendationResult, p50, p90 []float64) *model.AnalysisResultResponse {
	historicalData := model.HistoricalDataSummary{
		RowCount: len(rows),
	}
	forecastStart := session.CreatedAt.UTC().AddDate(0, 0, 1)
	if len(rows) > 0 {
		startDate := rows[0].SaleDate.UTC()
		endDate := rows[len(rows)-1].SaleDate.UTC()
		historicalData.StartDate = startDate.Format("2006-01-02")
		historicalData.EndDate = endDate.Format("2006-01-02")
		historicalData.PeriodDays = int(endDate.Sub(startDate).Hours()/24) + 1
		forecastStart = endDate.AddDate(0, 0, 1)
	}

	horizon := min(len(p50), len(p90))
	points := make([]model.ForecastPoint, horizon)
	var totalDemand float64
	for i := 0; i < horizon; i++ {
		points[i] = model.ForecastPoint{
			Date: forecastStart.AddDate(0, 0, i).Format("2006-01-02"),
			P50:  p50[i],
			P90:  p90[i],
		}
		totalDemand += p50[i]
	}
	averageDemand := float64(0)
	if horizon > 0 {
		averageDemand = totalDemand / float64(horizon)
	}
	averageDemand = math.Round(averageDemand*100) / 100

	return &model.AnalysisResultResponse{
		SKU: model.AnalysisResultSKU{
			ID:   sku.SKUID,
			Name: sku.Name,
		},
		AnalysisDate:       session.CreatedAt.UTC().Format("2006-01-02"),
		HistoricalData:     historicalData,
		CurrentStock:       session.CurrentStock,
		LeadTimeDays:       session.LeadTimeDays,
		TargetServiceLevel: constants.TargetServiceLevel,
		DemandCategory:     stringValue(session.DemandCategory),
		AverageDailyDemand: averageDemand,
		Forecast:           model.ForecastResponse{HorizonDays: horizon, Points: points},
		ReorderPoint:       result.ReorderPoint,
		ReorderQuantity:    result.ReorderQuantity,
		Risk:               model.RiskResponse{Label: result.RiskLabel, Reason: result.RiskReason},
		ExplanationText:    result.ExplanationText,
	}
}

func (s *AnalysisService) CompleteFromAI(sessionID uuid.UUID, req model.AIResultRequest) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return appErrors.InternalServer("gagal memproses hasil AI")
	}
	defer tx.Rollback()

	session, err := s.analysisRepo.GetSessionForUpdate(tx, sessionID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.NotFound("sesi analisis tidak ditemukan")
	}
	if err != nil {
		return appErrors.InternalServer("gagal memeriksa sesi analisis")
	}
	if session.Status == constants.AnalysisSessionCompleted && req.Status == "SUCCESS" {
		return tx.Commit().Error
	}
	if session.Status != constants.AnalysisSessionPendingAI {
		return appErrors.Conflict("sesi analisis sudah memiliki hasil")
	}
	if req.Status == "SUCCESS" {
		if req.DemandCategory == "" || req.RiskLabel == "" || req.RiskReason == "" || req.ExplanationText == "" || len(req.ForecastP50) == 0 || len(req.ForecastP90) == 0 {
			return appErrors.BadRequest("hasil AI sukses tidak lengkap")
		}
		p50, _ := json.Marshal(req.ForecastP50)
		p90, _ := json.Marshal(req.ForecastP90)
		err := s.analysisRepo.CreateRecommendationResult(tx, &entity.RecommendationResult{
			RecommendationResultID: uuid.New(),
			AnalysisSessionID:      sessionID,
			ReorderPoint:           req.ReorderPoint,
			ReorderQuantity:        req.ReorderQuantity,
			RiskLabel:              req.RiskLabel,
			RiskReason:             req.RiskReason,
			ExplanationText:        req.ExplanationText,
			ForecastP50:            string(p50),
			ForecastP90:            string(p90),
		})
		if err != nil {
			return appErrors.InternalServer("gagal menyimpan rekomendasi")
		}

		err = s.creditRepo.CompleteAnalysisDebit(tx, session.UserID, sessionID, session.CreditCost)
		if err != nil {
			return appErrors.InternalServer("gagal mendebit kredit")
		}

		err = s.analysisRepo.UpdateSessionResult(tx, sessionID, constants.AnalysisSessionCompleted, &req.DemandCategory, req.ADIValue, req.CVSquaredValue, nil, nil)
		if err != nil {
			return appErrors.InternalServer("gagal menyelesaikan sesi analisis")
		}
	} else {
		status := constants.AnalysisSessionAIFailed
		if req.Status == "INSUFFICIENT_DATA" {
			status = constants.AnalysisSessionInsufficientData
		}

		err = s.creditRepo.ReleaseAnalysisCredit(tx, session.UserID, session.CreditCost)
		if err != nil {
			return appErrors.InternalServer("gagal melepas reservasi kredit")
		}

		err = s.analysisRepo.UpdateSessionResult(tx, sessionID, status, nil, nil, nil, &req.ErrorCode, &req.ErrorMessage)
		if err != nil {
			return appErrors.InternalServer("gagal memperbarui sesi analisis")
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return appErrors.InternalServer("gagal memproses hasil AI")
	}

	return nil
}

type parsedXLSX struct {
	skuName string
	rows    []entity.AnalysisUploadRow
	errors  []entity.UploadValidationError
}

func parseXLSX(data []byte) parsedXLSX {
	result := parsedXLSX{}
	book, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		result.errors = append(result.errors, validationError(0, "INVALID_XLSX", "File XLSX tidak dapat dibaca."))
		return result
	}
	defer book.Close()

	sheets := book.GetSheetList()
	if len(sheets) == 0 {
		result.errors = append(result.errors, validationError(0, "EMPTY_SHEET", "Workbook tidak memiliki sheet."))
		return result
	}

	values, err := book.GetRows(sheets[0])
	if err != nil || len(values) < 2 {
		result.errors = append(result.errors, validationError(0, "EMPTY_DATA", "File XLSX tidak memiliki data penjualan."))
		return result
	}

	headers := map[string]int{}
	for i, value := range values[0] {
		headers[strings.ToLower(strings.TrimSpace(value))] = i
	}

	required := []string{"tanggal", "jumlah_terjual", "nama_sku", "harga_satuan"}
	for _, name := range required {
		if _, ok := headers[name]; !ok {
			result.errors = append(result.errors, validationError(1, "MISSING_COLUMN", fmt.Sprintf("Berkas tidak memiliki kolom %s.", name)))
		}
	}
	if len(result.errors) > 0 {
		return result
	}

	dates := map[string]bool{}
	expectedSKU := ""
	for index, row := range values[1:] {
		rowNumber := index + 2
		value := func(name string) string {
			i := headers[name]
			if i >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[i])
		}
		dateText, quantityText, skuName, priceText := value("tanggal"), value("jumlah_terjual"), value("nama_sku"), value("harga_satuan")
		date, dateErr := time.Parse("2006-01-02", dateText)
		quantity, quantityErr := strconv.Atoi(quantityText)
		price, priceErr := strconv.ParseFloat(priceText, 64)
		invalid := false
		if dateErr != nil {
			result.errors = append(result.errors, validationError(rowNumber, "INVALID_DATE_FORMAT", fmt.Sprintf("Format tanggal tidak valid pada baris %d. Gunakan format YYYY-MM-DD.", rowNumber)))
			invalid = true
		} else if dates[dateText] {
			result.errors = append(result.errors, validationError(rowNumber, "DUPLICATE_DATE", fmt.Sprintf("Ditemukan tanggal ganda pada baris %d.", rowNumber)))
			invalid = true
		} else {
			dates[dateText] = true
		}
		if quantityErr != nil || quantity < 0 {
			result.errors = append(result.errors, validationError(rowNumber, "INVALID_QUANTITY", fmt.Sprintf("Nilai penjualan tidak valid pada baris %d. Harus berupa angka nol atau lebih.", rowNumber)))
			invalid = true
		}
		if skuName == "" {
			result.errors = append(result.errors, validationError(rowNumber, "INVALID_SKU", fmt.Sprintf("Nama SKU wajib diisi pada baris %d.", rowNumber)))
			invalid = true
		} else if expectedSKU != "" && skuName != expectedSKU {
			result.errors = append(result.errors, validationError(rowNumber, "MULTIPLE_SKU", fmt.Sprintf("Nama SKU pada baris %d harus sama dengan baris lain.", rowNumber)))
			invalid = true
		}
		if priceErr != nil || price < 0 {
			result.errors = append(result.errors, validationError(rowNumber, "INVALID_UNIT_PRICE", fmt.Sprintf("Harga satuan tidak valid pada baris %d.", rowNumber)))
			invalid = true
		}
		if invalid {
			continue
		}
		if expectedSKU == "" {
			expectedSKU = skuName
		}
		result.rows = append(result.rows, entity.AnalysisUploadRow{
			AnalysisUploadRowID: uuid.New(),
			RowNumber:           rowNumber,
			SaleDate:            date,
			QuantitySold:        quantity,
			SKUName:             skuName,
			UnitPrice:           price,
		})
	}
	result.skuName = expectedSKU
	if len(result.rows) == 0 && len(result.errors) == 0 {
		result.errors = append(result.errors, validationError(0, "EMPTY_DATA", "File XLSX tidak memiliki data penjualan yang valid."))
	}

	return result
}

func validationError(row int, code, message string) entity.UploadValidationError {
	return entity.UploadValidationError{UploadValidationErrorID: uuid.New(), RowNumber: row, Code: code, Message: message}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func uploadResponse(upload *entity.AnalysisUpload, rows []entity.AnalysisUploadRow, validationErrors []entity.UploadValidationError) *model.AnalysisUploadResponse {
	preview := make([]model.AnalysisUploadPreviewRow, 0, min(len(rows), 5))
	for i, row := range rows {
		if i == 5 {
			break
		}
		preview = append(preview, model.AnalysisUploadPreviewRow{
			RowNumber:    row.RowNumber,
			SaleDate:     row.SaleDate.Format("2006-01-02"),
			QuantitySold: row.QuantitySold,
			SKUName:      row.SKUName,
			UnitPrice:    row.UnitPrice,
		})
	}

	errorsResponse := make([]model.UploadValidationErrorResponse, 0, min(len(validationErrors), 5))
	for i, item := range validationErrors {
		if i == 5 {
			break
		}
		errorsResponse = append(errorsResponse, model.UploadValidationErrorResponse{
			RowNumber: item.RowNumber,
			Code:      item.Code,
			Message:   item.Message,
		})
	}

	return &model.AnalysisUploadResponse{
		UploadID:      upload.AnalysisUploadID,
		Status:        upload.Status,
		SKUName:       upload.SKUName,
		ValidRowCount: upload.ValidRowCount,
		ErrorRowCount: upload.ErrorRowCount,
		ValidRows:     preview,
		Errors:        errorsResponse,
	}
}
