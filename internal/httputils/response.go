package httputils

type Message struct {
	Message string `json:"message"`
}

type ErrorMessage struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func Ok() *Message {
	return &Message{
		Message: "ok",
	}
}

func Error(message string) *ErrorMessage {
	return &ErrorMessage{
		Message: message,
	}
}

func ErrorWithDetails(message string, err error) *ErrorMessage {
	return &ErrorMessage{
		Message: message,
		Details: err.Error(),
	}
}
