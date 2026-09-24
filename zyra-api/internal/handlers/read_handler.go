package handlers

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) Get(kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var out any
		var e error
		switch kind {
		case "clients":
			v, err := h.Service.Client(c.Request.Context(), c.Param("id"))
			e = err
			out = v
		case "databases":
			v, err := h.Service.Database(c.Request.Context(), c.Param("id"))
			e = err
			out = v
		case "assessments":
			v, err := h.Service.Assessment(c.Request.Context(), c.Param("id"))
			e = err
			out = v
		}
		respond(c, 200, out, e)
	}
}
