package handlers

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) Archive(kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		e := h.Service.Archive(c.Request.Context(), actor(c), kind, c.Param("id"))
		respond(c, 200, gin.H{"archived": true}, e)
	}
}
