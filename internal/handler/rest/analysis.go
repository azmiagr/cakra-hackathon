package rest

import (
	"net/http"
	"strconv"

	"github.com/azmiagr/cakra-hackathon/model"
	"github.com/azmiagr/cakra-hackathon/pkg/helper"
	"github.com/azmiagr/cakra-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) GetAnalysisHistory(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	page, err := optionalPositiveInt(c.Query("page"), 1)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "parameter page tidak valid", nil)
		return
	}
	limit, err := optionalPositiveInt(c.Query("limit"), 8)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "parameter limit tidak valid", nil)
		return
	}
	sort := c.DefaultQuery("sort", "newest")
	if sort != "newest" && sort != "oldest" {
		response.Error(c, http.StatusBadRequest, "parameter sort harus newest atau oldest", nil)
		return
	}

	result, err := r.service.AnalysisService.GetHistory(user.UserID, model.AnalysisHistoryQuery{
		Search:    c.Query("search"),
		RiskLabel: c.Query("risk_label"),
		Page:      page,
		Limit:     limit,
		Sort:      sort,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "riwayat analisis berhasil diambil", result)
}

func (r *Rest) GetCategories(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.AnalysisService.ListCategories(user.UserID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "kategori produk berhasil diambil", result)
}

func (r *Rest) GetDashboard(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.AnalysisService.GetDashboard(user.UserID, user.FullName, c.Query("search"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "dashboard berhasil diambil", result)
}

func optionalPositiveInt(raw string, defaultValue int) (int, error) {
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}

func (r *Rest) UploadAnalysisXLSX(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "file XLSX wajib diunggah", nil)
		return
	}

	result, err := r.service.AnalysisService.Upload(user.UserID, file)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "file XLSX berhasil divalidasi", result)
}

func (r *Rest) CreateAnalysisSession(c *gin.Context) {
	uploadID, err := helper.ParseUUIDParam(c, "uploadID", "id unggahan analisis tidak valid")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.CreateAnalysisSessionRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "request analisis tidak valid", nil)
		return
	}

	result, err := r.service.AnalysisService.CreateSession(user.UserID, uploadID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "sesi analisis berhasil dibuat", result)
}

func (r *Rest) GetAnalysisSession(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	sessionID, err := helper.ParseUUIDParam(c, "sessionID", "id sesi analisis tidak valid")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.AnalysisService.GetSession(user.UserID, sessionID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "sesi analisis berhasil diambil", result)
}

func (r *Rest) GetCreditAccount(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CreditService.GetBalance(user.UserID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "saldo kredit berhasil diambil", result)
}

func (r *Rest) CompleteAnalysisFromAI(c *gin.Context) {
	sessionID, err := helper.ParseUUIDParam(c, "sessionID", "id sesi analisis tidak valid")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AIResultRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "hasil AI tidak valid", nil)
		return
	}

	err = r.service.AnalysisService.CompleteFromAI(sessionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "hasil AI berhasil diproses", nil)
}
