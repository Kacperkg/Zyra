package handlers

import (
	"github.com/gin-gonic/gin"
	"zyra-api/internal/services"
)

func (h *Handler) SaveClient(c *gin.Context) {
	var in services.ClientInput
	if !bind(c, &in) {
		return
	}
	v, e := h.Service.SaveClient(c.Request.Context(), actor(c), c.Param("id"), in)
	status := 200
	if c.Param("id") == "" {
		status = 201
	}
	respond(c, status, v, e)
}
