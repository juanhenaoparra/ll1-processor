package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"
)

type ErrResponse struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.StatusCode)
	return nil
}

func NewAPIError(code int, message string) render.Renderer {
	return &ErrResponse{
		StatusCode: code,
		Message:    message,
	}
}

func CloseOrLog(closable io.Closer) {
	if err := closable.Close(); err != nil {
		fmt.Println("closing_object_failed")
	}
}
