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

type Refund struct {
	refundService *service.RefundService
}

func NewRefund(refundService *service.RefundService) *Refund {
	return &Refund{refundService: refundService}
}

// RequestRefund godoc
// @Summary Request refund
// @Description Create a refund request based on payment intent code
// @Tags Refund
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.RefundRequest true "Refund Request Body"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/refunds [post]
func (h *Refund) RequestRefund(w http.ResponseWriter, r *http.Request) {

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

	var req model.RefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body request", http.StatusBadRequest)
		return
	}

	if req.PaymentIntentCode == "" {
		http.Error(w, "invalid payment intent id", http.StatusBadRequest)
		return
	}

	if req.Reason == "" {
		http.Error(w, "reason required", http.StatusBadRequest)
		return
	}

	res := h.refundService.RequestRefund(req.PaymentIntentCode, req.Reason, userID)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

// RequestRefund godoc
// @Summary List refund
// @Description Get paginated refund list with optional filters (status)
// @Tags Refund
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default 1)"
// @Param perpage query int false "Items per page (default 10)"
// @Param status query string false "Invoice status (REQUESTED → APPROVED / REJECTED → SUCCESS / FAILED)"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/refunds [get]
func (h *Refund) ListRefund(w http.ResponseWriter, r *http.Request) {
	h.ListRefundCore(w, r)
}

// RequestRefund godoc
// @Summary List refund (Admin)
// @Description Get paginated refund list with optional filters (status)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default 1)"
// @Param perpage query int false "Items per page (default 10)"
// @Param status query string false "Invoice status (REQUESTED → APPROVED / REJECTED → SUCCESS / FAILED)"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/refunds [get]
func (h *Refund) ListRefundAdmin(w http.ResponseWriter, r *http.Request) {
	h.ListRefundCore(w, r)
}

func (h *Refund) ListRefundCore(w http.ResponseWriter, r *http.Request) {
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

	var req model.ListRefundRequest
	userID := claims.UserID
	role := claims.Role
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perpage")
	status := r.URL.Query().Get("status")

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

	if status != "" && !IsValidRefundStatus(status) {
		http.Error(w, "invalid status value", http.StatusBadRequest)
		return
	}
	req.Status = status

	res := h.refundService.ListRefund(userID, role, req)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

// DecideRefund godoc
// @Summary Decide refund (approve/reject)
// @Description Approve or reject a refund request. Only refunds with status REQUESTED can be decided.
// @Tags [ADMIN] Refund
// @Accept json
// @Produce json
// @Param refund_number path string true "Refund Number"
// @Param request body model.RefundDecideRequestBody true "Refund decision payload"
// @Success 200 {object} model.RefundResponse
// @Failure 400 {object} model.ErrorBody
// @Failure 401 {object} model.ErrorBody
// @Router /api/v1/admin/refunds/{refund_number}/decision [patch]
// @Security BearerAuth
func (h *Refund) DecideRefund(w http.ResponseWriter, r *http.Request) {
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
	var req model.RefundDecideRequest
	var reqBody model.RefundDecideRequestBody

	req.RefundNumber = chi.URLParam(r, "refund_number")
	if req.RefundNumber == "" {
		http.Error(w, "missing refund number", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "invalid body request", http.StatusBadRequest)
		return
	}

	log.Print("reqBody.Status: ", reqBody.Status)
	if reqBody.Status != "" && !IsValidRefundStatus(reqBody.Status) {
		http.Error(w, "status required", http.StatusBadRequest)
		return
	}
	req.Status = reqBody.Status

	res := h.refundService.DecideRefund(userID, req)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

// ProcessRefund godoc
// @Summary Process refund
// @Description Execute the refund process for a previously approved refund. This will move the refund to a final state (e.g., SUCCESS or FAILED).
// @Tags [ADMIN] Refund
// @Accept json
// @Produce json
// @Param refund_number path string true "Refund Number"
// @Success 200 {object} model.RefundResponse
// @Failure 400 {object} model.ErrorBody
// @Failure 401 {object} model.ErrorBody
// @Router /api/v1/admin/refunds/{refund_number}/process [patch]
// @Security BearerAuth
func (h *Refund) ProcessRefund(w http.ResponseWriter, r *http.Request) {
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
	var req model.RefundProcessRequest

	req.RefundNumber = chi.URLParam(r, "refund_number")
	if req.RefundNumber == "" {
		http.Error(w, "missing refund number", http.StatusBadRequest)
		return
	}

	res := h.refundService.ProcessRefund(userID, req)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

func IsValidRefundStatus(status string) bool {
	switch status {
	case constant.RefundStatusRequested,
		constant.RefundStatusApproved,
		constant.RefundStatusRejected,
		constant.RefundStatusSuccess,
		constant.RefundStatusFailed:
		return true
	default:
		return false
	}
}
