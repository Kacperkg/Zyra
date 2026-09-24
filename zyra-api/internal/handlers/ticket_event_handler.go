package handlers

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func (h *Handler) Events(c *gin.Context) {
	page := 1
	if raw := c.Query("page"); raw != "" {
		n, e := strconv.Atoi(raw)
		if e != nil || n < 1 || n > 100000 {
			c.JSON(400, gin.H{"error": "invalid page"})
			return
		}
		page = n
	}
	v, e := h.Service.TicketEvents(c.Request.Context(), c.Param("id"), page)
	respond(c, 200, v, e)
}
