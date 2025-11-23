package protocol

// Response is sent from daemon back to client.
type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// NewSuccessResponse creates a successful response.
func NewSuccessResponse(message string) *Response {
	return &Response{
		Success: true,
		Message: message,
	}
}

// NewErrorResponse creates an error response.
func NewErrorResponse(err error) *Response {
	return &Response{
		Success: false,
		Error:   err.Error(),
	}
}
