package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"viopal-service/internal/constant"
	"viopal-service/internal/middleware"
	"viopal-service/internal/model"
	"viopal-service/internal/response"
	"viopal-service/internal/service"

	"github.com/go-chi/chi/v5"
)

type Balance struct {
	balanceService *service.BalanceService
}

func NewBalance(balanceService *service.BalanceService) *Balance {
	return &Balance{balanceService: balanceService}
}

// MerchantBalance godoc
// @Summary Get merchant wallet balance
// @Description Get merchant balance from authenticated user token
// @Tags Merchant Wallet
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/wallet [get]
func (h *Balance) MerchantBalance(w http.ResponseWriter, r *http.Request) {
	userValue := r.Context().Value(middleware.UserContextKey)
	if userValue == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims, ok := userValue.(*middleware.Claims)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID

	res := h.balanceService.MerchantBalance(userID)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

// RequestTopUp godoc
// @Summary Request top up balance
// @Description Merchant request balance top up
// @Tags Merchant Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.TopUpRequest true "Top Up Request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/balance-requests [post]
func (h *Balance) RequestTopUp(w http.ResponseWriter, r *http.Request) {
	userValue := r.Context().Value(middleware.UserContextKey)
	if userValue == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims, ok := userValue.(*middleware.Claims)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID

	var req model.TopUpRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Fail(w, model.ErrorBody{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: "Invalid Body Request",
			Status:  http.StatusBadRequest,
		})
		return
	}

	res := h.balanceService.RequestTopUp(userID, req.Amount)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

// ListRequestTopUp godoc
// @Summary List top up requests
// @Description Get paginated list of top up requests for authenticated merchant with optional filters
// @Tags Merchant Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default 1)" example(1)
// @Param perpage query int false "Items per page (default 10)" example(10)
// @Param status query string false "Top up status" Enums(PENDING,SUCCESS,FAILED)
// @Param ref_number query string false "Top up reference number" example(TOPUP-20260504-0001)
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/balance-requests [get]
// @Tags Merchant Wallet
// @Router /api/v1/balance-requests [get]
func (h *Balance) ListRequestTopUp(w http.ResponseWriter, r *http.Request) {
	h.ListRequestTopUpCore(w, r)
}

// ListBalanceRequestAdmin godoc
// @Summary List top up requests (Admin)
// @Description Get paginated list of all merchant top up requests with optional filters
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default 1)" example(1)
// @Param perpage query int false "Items per page (default 10)" example(10)
// @Param status query string false "Top up status" Enums(PENDING,SUCCESS,FAILED)
// @Param ref_number query string false "Top up reference number" example(TOPUP-20260504-0001)
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/balance-requests [get]
func (h *Balance) ListRequestTopUpAdmin(w http.ResponseWriter, r *http.Request) {
	h.ListRequestTopUpCore(w, r)
}

func (h *Balance) ListRequestTopUpCore(w http.ResponseWriter, r *http.Request) {
	var req model.ListTopUpRequest
	userValue := r.Context().Value(middleware.UserContextKey)
	if userValue == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims, ok := userValue.(*middleware.Claims)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID
	role := claims.Role
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perpage")
	status := r.URL.Query().Get("status")
	refNumber := r.URL.Query().Get("ref_number")

	req.Page = 1
	req.PerPage = 10
	var err error

	if pageStr != "" {
		req.Page, err = strconv.Atoi(pageStr)
		if err != nil || req.Page < 1 {
			http.Error(w, "invalid page", http.StatusBadRequest)
			return
		}
	}

	if perPageStr != "" {
		req.PerPage, err = strconv.Atoi(perPageStr)
		if err != nil || req.PerPage < 1 {
			http.Error(w, "invalid perpage", http.StatusBadRequest)
			return
		}
	}

	if status != "" && !IsValidTopUptStatus(status) {
		http.Error(w, "invalid status value", http.StatusBadRequest)
		return
	}
	req.Status = status

	if refNumber != "" {
		req.RefNumber = refNumber
	}

	res := h.balanceService.ListRequestTopUp(userID, role, req)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

// UpdateTopUp godoc
// @Summary Update top up request status
// @Description Update top up request status by reference number (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ref_number path string true "Top Up Reference Number"
// @Param request body model.UpdateTopUpRequestBody true "Status only"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/balance-requests/{ref_number}/status [patch]
func (h *Balance) UpdateTopUp(w http.ResponseWriter, r *http.Request) {
	userValue := r.Context().Value(middleware.UserContextKey)
	if userValue == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims, ok := userValue.(*middleware.Claims)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID
	var req model.UpdateTopUpRequest
	var reqBody model.UpdateTopUpRequestBody

	req.RefNumber = chi.URLParam(r, "ref_number")
	if req.RefNumber == "" {
		http.Error(w, "missing top up reference number", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "invalid body request", http.StatusBadRequest)
		return
	}
	log.Print("reqBody.Status: ", reqBody.Status)

	if reqBody.Status != "" && !IsValidTopUptStatus(reqBody.Status) {
		http.Error(w, "status required", http.StatusBadRequest)
		return
	}
	req.Status = reqBody.Status

	log.Print(req)
	res := h.balanceService.UpdateTopUp(userID, req)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

func IsValidTopUptStatus(status string) bool {
	switch status {
	case constant.TopUpStatusPending,
		constant.TopUpStatusSuccess,
		constant.TopUpStatusFailed:
		return true
	default:
		return false
	}
}
