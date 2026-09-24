package handlers

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) Ticket(c *gin.Context) {
	v, e := h.Service.TicketDetail(c.Request.Context(), c.Param("id"))
	respond(c, 200, v, e)
}
func (h *Handler) Action(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if action == "close" || action == "reopen" {
			v, e := h.Service.TicketAction(c.Request.Context(), actor(c), c.Param("id"), action, "")
			respond(c, 200, v, e)
			return
		}
		var in struct {
			Comment string `json:"comment"`
		}
		if !bind(c, &in) {
			return
		}
		v, e := h.Service.TicketAction(c.Request.Context(), actor(c), c.Param("id"), action, in.Comment)
		respond(c, 200, v, e)
	}
}
