package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"zyra-api/internal/apperrors"
	"zyra-api/internal/models"
	"zyra-api/internal/services"
)

func actor(c *gin.Context) models.User { v, _ := c.Get("user"); return v.(models.User) }
func respond(c *gin.Context, status int, v any, err error) {
	if err == nil {
		c.JSON(status, v)
		return
	}
	code := http.StatusInternalServerError
	message := "internal server error"
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		code = 404
		message = "not found"
	case errors.Is(err, apperrors.ErrConflict):
		code = 409
		message = "conflict"
	case errors.Is(err, apperrors.ErrInvalidInput):
		code = 400
		message = err.Error()
	case errors.Is(err, apperrors.ErrUnauthorized):
		code = 401
		message = "invalid credentials or expired session"
	case errors.Is(err, apperrors.ErrForbidden):
		code = 403
		message = "permission denied"
	case errors.Is(err, services.ErrRecoveryUnavailable):
		code = 503
		message = err.Error()
	default:
		log.Printf("request failed: %v", err)
	}
	c.JSON(code, gin.H{"error": message})
}
func bind(c *gin.Context, v any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		c.JSON(400, gin.H{"error": "invalid JSON request"})
		return false
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		c.JSON(400, gin.H{"error": "expected one JSON object"})
		return false
	}
	return true
}
