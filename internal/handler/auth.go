package handler

import (
	"encoding/json"
	"net/http"

	"viopal-service/internal/model"
	"viopal-service/internal/response"
	"viopal-service/internal/service"

	"github.com/go-playground/validator/v10"
)

type Auth struct {
	authService *service.AuthService
}

var validate = validator.New()

func NewAuth(authService *service.AuthService) *Auth {
	return &Auth{authService: authService}
}

// Register godoc
// @Summary Register merchant account
// @Description Create new merchant user and merchant profile
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Register Request"
// @Success 201 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Fail(w, model.ErrorBody{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: "Failed Get Param",
			Status:  http.StatusBadRequest,
		})
		return
	}

	if err := validate.Struct(req); err != nil {
		response.Fail(w, model.ErrorBody{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: "Failed Get Param",
			Status:  http.StatusBadRequest,
		})
		return
	}

	res := h.authService.Register(req)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res.Register)
}

// Login godoc
// @Summary Login user
// @Description Login using email and password, returns JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login Request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Fail(w, model.ErrorBody{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: "Invalid Body Request",
			Status:  http.StatusBadRequest,
		})
		return
	}

	res := h.authService.Login(req)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get profile data by email
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param email query string true "User email"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/me [get]
func (h *Auth) Me(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	if email == "" {
		response.Fail(w, model.ErrorBody{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: "Email is Required",
			Status:  http.StatusBadRequest,
		})
		return
	}

	res := h.authService.Me(email)
	if res.Error != nil {
		response.Fail(w, *res.Error)
		return
	}

	response.Success(w, res)
}
