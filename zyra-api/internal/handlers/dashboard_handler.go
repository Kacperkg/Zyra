package handlers

import "github.com/gin-gonic/gin"

func (h *Handler) Dashboard(c *gin.Context) {
	v, e := h.Service.Dashboard(c.Request.Context())
	respond(c, 200, v, e)
}
