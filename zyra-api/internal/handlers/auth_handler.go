package handlers

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) ChangePassword(c *gin.Context) {
	var in struct {
		Current  string `json:"current_password"`
		Password string `json:"password"`
	}
	if !bind(c, &in) {
		return
	}
	e := h.Service.ChangePassword(c.Request.Context(), actor(c), in.Current, in.Password)
	respond(c, 200, gin.H{"message": "Password changed; sign in again."}, e)
}

func (h *Handler) Login(c *gin.Context) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !bind(c, &in) {
		return
	}
	v, e := h.Service.Login(c.Request.Context(), in.Email, in.Password)
	respond(c, 200, v, e)
}
func (h *Handler) Refresh(c *gin.Context) {
	var in struct {
		Token string `json:"refresh_token"`
	}
	if !bind(c, &in) {
		return
	}
	v, e := h.Service.Refresh(c.Request.Context(), in.Token)
	respond(c, 200, v, e)
}
func (h *Handler) Logout(c *gin.Context) {
	e := h.Service.Logout(c.Request.Context(), c.GetString("session_id"))
	respond(c, 200, gin.H{"ok": true}, e)
}
func (h *Handler) Forgot(c *gin.Context) {
	var in struct {
		Email string `json:"email"`
	}
	if !bind(c, &in) {
		return
	}
	e := h.Service.ForgotPassword(c.Request.Context(), in.Email)
	respond(c, 202, gin.H{"message": "If eligible, recovery instructions will be sent."}, e)
}
func (h *Handler) Reset(c *gin.Context) {
	var in struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !bind(c, &in) {
		return
	}
	e := h.Service.ResetPassword(c.Request.Context(), in.Token, in.Password)
	respond(c, 200, gin.H{"ok": true}, e)
}
