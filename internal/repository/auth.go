package repository

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"viopal-service/internal/model"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *AuthRepository) InsertUser(ctx context.Context, tx *sql.Tx, name string, email string, passwordHash string) (int64, error) {
	query := `INSERT INTO users (name, email, password_hash, role_id) VALUES (?, ?, ?, ?)`

	result, err := tx.ExecContext(ctx, query, name, email, passwordHash, 2)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *AuthRepository) InsertMerchant(ctx context.Context, tx *sql.Tx, userID int64, businessName string, phoneNumber string) error {
	query := `INSERT INTO merchant (user_id, bussiness_name, phone_number, balance) VALUES (?, ?, ?, ?)`
	_, err := tx.ExecContext(ctx, query, userID, businessName, phoneNumber, 0)
	return err
}

func (r *AuthRepository) FindUserByEmail(email string) (model.User, error) {
	var user model.User

	query := `
		SELECT u.id, u.name, u.email, u.password_hash, r.name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE email = ? LIMIT 1
	`

	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role)
	if err == sql.ErrNoRows {
		return user, errors.New("Invalid Email or Password")
	}

	if err != nil {
		return user, err
	}

	return user, nil
}

func (r *AuthRepository) FindUser(email string) model.MeResponse {
	var user model.Me

	query := `
		SELECT u.name, u.email, r.name, m.bussiness_name, m.phone_number
		FROM users u
		LEFT JOIN merchant m ON u.id = m.user_id 
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE email = ? LIMIT 1
	`

	err := r.db.QueryRow(query, email).Scan(&user.Name, &user.Email, &user.Role, &user.BusinessName, &user.PhoneNumber)
	if err == sql.ErrNoRows {
		return model.MeResponse{
			Me: &user,
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			},
		}
	}

	if err != nil {
		return model.MeResponse{
			Me: &user,
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	return model.MeResponse{Me: &user}
}

func (r *AuthRepository) GetMerchant(userId int) (model.Merchant, error) {
	var merchant model.Merchant

	query := `SELECT id, user_id, bussiness_name, phone_number, balance FROM merchant m WHERE user_id = ? LIMIT 1`

	err := r.db.QueryRow(query, userId).Scan(&merchant.ID, &merchant.UserID, &merchant.BusinessName, &merchant.PhoneNumber, &merchant.Balance)
	if err == sql.ErrNoRows {
		return merchant, err
	}

	if err != nil {
		return merchant, err
	}

	return merchant, nil
}

func (r *AuthRepository) GetMerchantById(merchantID int) (model.Merchant, error) {
	var merchant model.Merchant

	query := `SELECT id, user_id, bussiness_name, phone_number, balance FROM merchant m WHERE id = ? LIMIT 1`

	err := r.db.QueryRow(query, merchantID).Scan(&merchant.ID, &merchant.UserID, &merchant.BusinessName, &merchant.PhoneNumber, &merchant.Balance)
	if err == sql.ErrNoRows {
		return merchant, err
	}

	if err != nil {
		return merchant, err
	}

	return merchant, nil
}
