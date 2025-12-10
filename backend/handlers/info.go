package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type InfoHandler struct {
	inviteRequired bool
}

func NewInfoHandler(inviteRequired bool) *InfoHandler {
	return &InfoHandler{inviteRequired: inviteRequired}
}

type InfoGetResponse struct {
	InviteRequired bool `json:"inviteRequired"`
}

func (h *InfoHandler) InfoHandlerGet(c echo.Context) error {
	return c.JSON(http.StatusOK, InfoGetResponse{
		InviteRequired: h.inviteRequired,
	})
}
