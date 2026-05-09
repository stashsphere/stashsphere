package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type ImportHandler struct {
	importService  *services.ImportService
	maxUploadBytes int64
}

func NewImportHandler(importService *services.ImportService, maxUploadBytes int64) *ImportHandler {
	return &ImportHandler{importService, maxUploadBytes}
}

func (ih *ImportHandler) Post(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}

	// Cap the total request body before the multipart parser reads it.
	c.Request().Body = http.MaxBytesReader(c.Response().Writer, c.Request().Body, ih.maxUploadBytes)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return utils.InvalidImportFileError{Msg: "failed to read uploaded file: " + err.Error()}
	}

	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	imp, err := ih.importService.QueueImport(c.Request().Context(), authCtx.User.UserId, src, ih.maxUploadBytes)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusAccepted, resources.ImportStatusFromModel(imp))
}

func (ih *ImportHandler) Get(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}

	imp, err := ih.importService.GetImport(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.ImportStatusFromModel(imp))
}
