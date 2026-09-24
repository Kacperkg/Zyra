package handlers

import (
	"github.com/gin-gonic/gin"
	"zyra-api/internal/services"
)

func (h *Handler) Me(c *gin.Context) { c.JSON(200, actor(c)) }
func (h *Handler) SaveUser(c *gin.Context) {
	var in services.UserInput
	if !bind(c, &in) {
		return
	}
	id := c.Param("id")
	if c.FullPath() == "/api/users/me" {
		id = actor(c).ID
	}
	v, e := h.Service.SaveUser(c.Request.Context(), actor(c), id, in)
	status := 200
	if id == "" {
		status = 201
	}
	respond(c, status, v, e)
}
