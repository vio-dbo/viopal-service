package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"viopal-service/internal/constant"
	"viopal-service/internal/middleware"
	"viopal-service/internal/model"
	"viopal-service/internal/response"
	"viopal-service/internal/service"

	"github.com/go-chi/chi/v5"
)

type Invoice struct {
	invoiceService *service.InvoiceService
}

func NewInvoice(invoiceService *service.InvoiceService) *Invoice {
	return &Invoice{invoiceService: invoiceService}
}

// RequestInvoices godoc
// @Summary Create invoice
// @Description Create new invoice request for merchant
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.InvoiceRequest true "Invoice Request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/invoices [post]
func (h *Invoice) RequestInvoices(w http.ResponseWriter, r *http.Request) {
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

	var req model.InvoiceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Fail(w, model.ErrorBody{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: "Invalid Body Request",
			Status:  http.StatusBadRequest,
		})
		return
	}

	res := h.invoiceService.RequestInvoices(userID, req.DueInDays, req.Amount, req.CustName, req.CustEmail)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

// ListInvoice godoc
// @Summary List invoices
// @Description Get paginated invoice list with optional filters (status, customer name, email)
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default 1)"
// @Param perpage query int false "Items per page (default 10)"
// @Param status query string false "Invoice status (PENDING, PAID, EXPIRED)"
// @Param cust_name query string false "Customer name (partial match)"
// @Param cust_email query string false "Customer email (partial match)"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/invoices [get]
func (h *Invoice) ListInvoice(w http.ResponseWriter, r *http.Request) {
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

	var req model.ListInvoiceRequest
	userID := claims.UserID
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perpage")
	status := r.URL.Query().Get("status")
	custName := r.URL.Query().Get("cust_name")
	custEmail := r.URL.Query().Get("cust_email")

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

	if status != "" && !IsValidInvoiceStatus(status) {
		http.Error(w, "invalid status value", http.StatusBadRequest)
		return
	}
	req.Status = status

	if custName != "" {
		req.CustName = custName
	}

	if custEmail != "" {
		req.CustEmail = custEmail
	}

	res := h.invoiceService.ListInvoice(userID, req)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

func IsValidInvoiceStatus(status string) bool {
	switch status {
	case constant.InvoiceStatusPending,
		constant.InvoiceStatusPaid,
		constant.InvoiceStatusExpired:
		return true
	default:
		return false
	}
}

// InvoiceByCode godoc
// @Summary Get invoice by code
// @Description Get invoice detail using invoice code for authenticated user
// @Tags Invoice
// @Produce json
// @Security BearerAuth
// @Param code path string true "Invoice code"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/invoices/{code} [get]
func (h *Invoice) InvoiceByCode(w http.ResponseWriter, r *http.Request) {
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
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "missing invoice code", http.StatusBadRequest)
		return
	}

	res := h.invoiceService.InvoiceByCode(userID, code)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

// InvoiceStatistic godoc
// @Summary Get invoice statistics
// @Description Retrieve invoice statistics filtered by merchant_id and created date range
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param merchant_id query int false "Merchant ID"
// @Param start_created_date query string false "Start created date (YYYY-MM-DD)"
// @Param end_created_date query string false "End created date (YYYY-MM-DD)"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/stats [get]
func (h *Invoice) InvoiceStatistic(w http.ResponseWriter, r *http.Request) {
	var req model.InvoiceStatisticReq
	query := r.URL.Query()
	if m := query.Get("merchant_id"); m != "" {
		id, err := strconv.Atoi(m)
		if err != nil || id <= 0 {
			http.Error(w, "invalid merchant_id", http.StatusBadRequest)
			return
		}
		req.MerchantID = &id
	}

	req.StartCreatedDate = query.Get("start_created_date")
	req.EndCreatedDate = query.Get("end_created_date")

	if req.StartCreatedDate != "" && !IsValidDateOnly(req.StartCreatedDate) {
		http.Error(w, "invalid start_created_date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	if req.EndCreatedDate != "" && !IsValidDateOnly(req.EndCreatedDate) {
		http.Error(w, "invalid end_created_date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	if req.StartCreatedDate != "" && req.EndCreatedDate != "" {
		start, _ := time.Parse("2006-01-02", req.StartCreatedDate)
		end, _ := time.Parse("2006-01-02", req.EndCreatedDate)

		if start.After(end) {
			http.Error(w, "start_created_date cannot be after end_created_date", http.StatusBadRequest)
			return
		}
	}

	res := h.invoiceService.InvoiceStatistic(req)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

func IsValidDateOnly(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}
