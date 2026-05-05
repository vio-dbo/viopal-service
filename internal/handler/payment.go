package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"viopal-service/internal/constant"
	"viopal-service/internal/middleware"
	"viopal-service/internal/model"
	"viopal-service/internal/response"
	"viopal-service/internal/service"

	"github.com/go-chi/chi/v5"
)

type Payment struct {
	paymentService *service.PaymentService
}

func NewPayment(paymentService *service.PaymentService) *Payment {
	return &Payment{paymentService: paymentService}
}

// Payment godoc
// @Summary Process payment by token
// @Description Process payment using payment link token
// @Tags Payment
// @Produce json
// @Param token path string true "Payment Token"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/public/pay/{token} [get]
func (h *Payment) Payment(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "missing payment token", http.StatusBadRequest)
		return
	}

	res := h.paymentService.Payment(token)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

// CreatePaymentIntent godoc
// @Summary Create payment intent
// @Description Create payment intent using payment link token and selected payment method
// @Tags Payment
// @Accept json
// @Produce json
// @Param token path string true "Payment Token"
// @Param request body model.PaymentIntentRequest true "Payment Method Request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/public/pay/{token}/intents [post]
func (h *Payment) PaymentIntent(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "missing payment token", http.StatusBadRequest)
		return
	}

	var paymentIntentRequest model.PaymentIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&paymentIntentRequest); err != nil {
		http.Error(w, "invalid body request", http.StatusBadRequest)
		return
	}

	if paymentIntentRequest.PaymentMethodID <= 0 {
		http.Error(w, "invalid payment_method_id", http.StatusBadRequest)
		return
	}

	res := h.paymentService.PaymentIntent(token, paymentIntentRequest.PaymentMethodID)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

// PaymentIntentByCode godoc
// @Summary Get payment intent by code
// @Description Get payment intent detail using public intent code
// @Tags Payment
// @Produce json
// @Param code path string true "Payment Intent Code"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/public/intents/{code} [get]
func (h *Payment) PaymentIntentByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "missing invoice code", http.StatusBadRequest)
		return
	}

	res := h.paymentService.PaymentIntentByCode(code)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

// ListPaymentIntent godoc
// @Summary List payment intents
// @Description Get payment intent list with pagination and optional filters
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" example(1)
// @Param perpage query int false "Items per page" example(10)
// @Param status query string false "Payment intent status" Enums(PENDING,SUCCESS,FAILED,EXPIRED)
// @Param payment_intent_number query string false "Payment intent number"
// @Param payment_method_id query int false "Payment method ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/payment-intents [get]
func (h *Payment) ListPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var req model.ListPaymentIntentRequest
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perpage")
	status := r.URL.Query().Get("status")
	paymentIntentNumber := r.URL.Query().Get("payment_intent_number")
	paymentMethodID := r.URL.Query().Get("payment_method_id")

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

	if status != "" && !IsValidPaymentIntentStatus(status) {
		http.Error(w, "invalid status value", http.StatusBadRequest)
		return
	}
	req.Status = status

	if paymentIntentNumber != "" {
		req.PaymentIntentNumber = paymentIntentNumber
	}

	if paymentMethodID != "" {
		req.PaymentMethodID, err = strconv.Atoi(paymentMethodID)
		if err != nil || req.PaymentMethodID < 1 {
			http.Error(w, "invalid payment method", http.StatusBadRequest)
			return
		}
	}

	res := h.paymentService.ListPaymentIntent(req)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

func IsValidPaymentIntentStatus(status string) bool {
	switch status {
	case constant.PaymentIntentStatusPending,
		constant.PaymentIntentStatusSuccess,
		constant.PaymentIntentStatusFailed:
		return true
	default:
		return false
	}
}

// UpdatePaymentIntent godoc
// @Summary Update payment intent status
// @Description Update payment intent status by payment intent number (Admin only). If status is FAILED, failure_reason is required.
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment_intent_number path string true "Payment Intent Number"
// @Param request body model.UpdatePaymentIntentRequestBody true "Update Payment Intent Request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/payment-intents/{payment_intent_number}/status [patch]
func (h *Payment) UpdatePaymentIntent(w http.ResponseWriter, r *http.Request) {
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
	var req model.UpdatePaymentIntentRequest
	var reqBody model.UpdatePaymentIntentRequestBody

	req.PaymentIntentNumber = chi.URLParam(r, "payment_intent_number")
	if req.PaymentIntentNumber == "" {
		http.Error(w, "missing payment intent number", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "invalid body request", http.StatusBadRequest)
		return
	}

	if reqBody.Status != "" && !IsValidTopUptStatus(reqBody.Status) {
		http.Error(w, "status required", http.StatusBadRequest)
		return
	}
	req.Status = reqBody.Status

	reqBody.FailureReason = chi.URLParam(r, "failure_reason")
	if reqBody.Status == constant.PaymentIntentStatusFailed && reqBody.FailureReason == "" {
		http.Error(w, "missing payment intent number", http.StatusBadRequest)
		return
	}
	req.FailureReason = reqBody.FailureReason

	res := h.paymentService.UpdatePaymentIntent(userID, req)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}
