package handlers

import (
	"zyra-api/internal/services"
)

type Handler struct{ Service *services.Service }

func New(s *services.Service) *Handler { return &Handler{s} }
