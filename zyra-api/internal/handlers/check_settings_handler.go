package handlers

import (
	"github.com/gin-gonic/gin"
	"zyra-api/internal/models"
)

func (h *Handler) Settings(c *gin.Context) {
	v, e := h.Service.GetSettings(c.Request.Context(), c.Param("id"))
	respond(c, 200, v, e)
}
func (h *Handler) PutSettings(c *gin.Context) {
	var in models.CheckSettings
	if !bind(c, &in) {
		return
	}
	e := h.Service.PutSettings(c.Request.Context(), actor(c), c.Param("id"), in)
	respond(c, 200, in, e)
}
