package response

import (
	"encoding/json"
	"net/http"
	"viopal-service/internal/model"
)

type Meta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	TotalItems int `json:"total_items,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

type SuccessResponse struct {
	Data interface{} `json:"data,omitempty"`
	Meta interface{} `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Error model.ErrorBody `json:"error"`
}

func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, SuccessResponse{
		Data: data,
	})
}

func SuccessPaginated(w http.ResponseWriter, data interface{}, meta Meta) {
	JSON(w, http.StatusOK, SuccessResponse{
		Data: data,
		Meta: meta,
	})
}

func Fail(w http.ResponseWriter, err model.ErrorBody) {
	JSON(w, err.Status, ErrorResponse{
		Error: model.ErrorBody{
			Code:    err.Code,
			Message: err.Message,
			Status:  err.Status,
		},
	})
}
