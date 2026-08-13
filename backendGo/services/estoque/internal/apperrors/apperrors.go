package apperrors

import "net/http"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *AppError) Error() string {
	if e.Code == "" {
		return e.Message
	}
	return e.Code + ": " + e.Message
}

const CodeDuplicado = "IDEMPOTENCY_DUPLICATE"

func BadRequest(msg string) *AppError {
	return &AppError{Code: "BAD_REQUEST", Message: msg, StatusCode: http.StatusBadRequest}
}

func NotFound(msg string) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: msg, StatusCode: http.StatusNotFound}
}

func Conflict(msg string) *AppError {
	return &AppError{Code: "CONFLICT", Message: msg, StatusCode: http.StatusConflict}
}

func Unprocessable(msg string) *AppError {
	return &AppError{Code: "UNPROCESSABLE_ENTITY", Message: msg, StatusCode: http.StatusUnprocessableEntity}
}

func Internal(msg string) *AppError {
	return &AppError{Code: "INTERNAL_ERROR", Message: msg, StatusCode: http.StatusInternalServerError}
}

func ServiceUnavailable(msg string) *AppError {
	return &AppError{Code: "SERVICE_UNAVAILABLE", Message: msg, StatusCode: http.StatusServiceUnavailable}
}

func Duplicado(msg string) *AppError {
	return &AppError{Code: CodeDuplicado, Message: msg, StatusCode: http.StatusConflict}
}
