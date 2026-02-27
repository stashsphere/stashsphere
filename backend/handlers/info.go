package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/config"
)

type InfoHandler struct {
	inviteRequired bool
	oidcProviders  []config.OIDCProviderConfig
}

func NewInfoHandler(inviteRequired bool, oidcProviders []config.OIDCProviderConfig) *InfoHandler {
	return &InfoHandler{
		inviteRequired: inviteRequired,
		oidcProviders:  oidcProviders,
	}
}

type OIDCProviderInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type InfoGetResponse struct {
	InviteRequired bool               `json:"inviteRequired"`
	OIDCProviders  []OIDCProviderInfo `json:"oidcProviders"`
}

func (h *InfoHandler) InfoHandlerGet(c echo.Context) error {
	providers := make([]OIDCProviderInfo, 0, len(h.oidcProviders))
	for _, p := range h.oidcProviders {
		providers = append(providers, OIDCProviderInfo{
			Name:        p.Name,
			DisplayName: p.DisplayName,
		})
	}
	return c.JSON(http.StatusOK, InfoGetResponse{
		InviteRequired: h.inviteRequired,
		OIDCProviders:  providers,
	})
}
