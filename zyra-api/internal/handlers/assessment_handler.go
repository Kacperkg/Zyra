package handlers

import (
	"github.com/gin-gonic/gin"
	"zyra-api/internal/services"
)

func (h *Handler) Submit(c *gin.Context) {
	var in services.ReportInput
	if !bind(c, &in) {
		return
	}
	v, e := h.Service.SubmitReport(c.Request.Context(), actor(c), c.Param("id"), in)
	respond(c, 200, v, e)
}
func (h *Handler) Raw(c *gin.Context) {
	body, e := h.Service.AssessmentRawBody(c.Request.Context(), c.Param("id"))
	if e != nil {
		respond(c, 200, nil, e)
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	c.String(200, "%s", body)
}
