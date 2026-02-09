package dto

type ErrorResponse struct {
	Error string `json:"error"`
}

type DeleteResponse struct {
	Message string `json:"message"`
}
