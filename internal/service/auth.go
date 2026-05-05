package service

import (
	"context"
	"net/http"
	"time"
	"viopal-service/internal/model"
	"viopal-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authRepo *repository.AuthRepository
}

func NewAuthService(authRepo *repository.AuthRepository) *AuthService {
	return &AuthService{
		authRepo: authRepo,
	}
}

func (s *AuthService) Register(req model.RegisterRequest) model.RegisterResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.authRepo.BeginTx(ctx)
	if err != nil {
		return model.RegisterResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	defer tx.Rollback()

	// hash password
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return model.RegisterResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	// insert user
	userID, err := s.authRepo.InsertUser(
		ctx,
		tx,
		req.Name,
		req.Email,
		string(hash),
	)
	if err != nil {
		return model.RegisterResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	// insert merchant
	err = s.authRepo.InsertMerchant(
		ctx,
		tx,
		userID,
		req.BusinessName,
		req.PhoneNumber,
	)
	if err != nil {
		return model.RegisterResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	tx.Commit()

	return model.RegisterResponse{Register: &req}
}

func (s *AuthService) Login(req model.LoginRequest) model.LoginResponse {
	user, err := s.authRepo.FindUserByEmail(req.Email)
	if err != nil {
		return model.LoginResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	// compare plain password vs hash
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	)

	if err != nil {
		return model.LoginResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusUnauthorized),
				Message: "Invalid Email or Password",
				Status:  http.StatusUnauthorized,
			},
		}
	}

	// generate JWT
	token, err := generateToken(user.ID, user.Role)
	if err != nil {
		return model.LoginResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusUnauthorized),
				Message: err.Error(),
				Status:  http.StatusUnauthorized,
			},
		}
	}

	return model.LoginResponse{
		Login: &req,
		Token: token,
	}
}

func generateToken(userID int64, role string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     now.Add(8 * time.Hour).Unix(),
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
		"iss":     "viopal-api",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your-secret-key"))
}

func (s *AuthService) Me(email string) model.MeResponse {
	var me model.MeResponse
	me = s.authRepo.FindUser(email)
	return me
}
