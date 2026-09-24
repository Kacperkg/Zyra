package handlers

import (
	"github.com/gin-gonic/gin"
	"zyra-api/internal/models"
)

func (h *Handler) Schedules(c *gin.Context) {
	if c.Request.Method == "GET" {
		v, e := h.Service.SourceSchedules(c.Request.Context(), c.Param("source_id"))
		respond(c, 200, v, e)
		return
	}
	var in struct {
		Schedules []models.Schedule `json:"schedules"`
	}
	if !bind(c, &in) {
		return
	}
	v, e := h.Service.SaveSourceSchedules(c.Request.Context(), actor(c), c.Param("source_id"), in.Schedules)
	respond(c, 200, v, e)
}

func (h *Handler) EvaluateSchedules(c *gin.Context) {
	var in struct {
		Date string `json:"date"`
	}
	if !bind(c, &in) {
		return
	}
	n, e := h.Service.EvaluateSchedules(c.Request.Context(), actor(c), in.Date)
	respond(c, 200, gin.H{"created": n}, e)
}
