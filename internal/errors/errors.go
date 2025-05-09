package errors

import "encoding/json"

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError - конструктор для новой ошибки Error
func NewAPIError(code int, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

func (e *APIError) ToJSON() []byte {
	jsonData, _ := json.Marshal(e)
	return jsonData
}
