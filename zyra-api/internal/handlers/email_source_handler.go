package handlers

import (
	"github.com/gin-gonic/gin"
	"zyra-api/internal/models"
)

func (h *Handler) Source(c *gin.Context) {
	v, e := h.Service.EmailSource(c.Request.Context(), c.Param("source_id"))
	respond(c, 200, v, e)
}

func (h *Handler) DisableSource(c *gin.Context) {
	e := h.Service.DisableSource(c.Request.Context(), actor(c), c.Param("source_id"))
	respond(c, 200, gin.H{"enabled": false}, e)
}

func (h *Handler) Sources(c *gin.Context) {
	v, e := h.Service.EmailSources(c.Request.Context(), c.Param("id"))
	respond(c, 200, v, e)
}
func (h *Handler) SaveSource(c *gin.Context) {
	var in models.EmailSource
	if !bind(c, &in) {
		return
	}
	dbID := c.Param("id")
	if c.Param("source_id") != "" {
		v, e := h.Service.SaveSourceByID(c.Request.Context(), actor(c), c.Param("source_id"), in)
		respond(c, 200, v, e)
		return
	} else {
		in.ID = ""
	}
	v, e := h.Service.SaveSource(c.Request.Context(), actor(c), dbID, in)
	respond(c, 200, v, e)
}
