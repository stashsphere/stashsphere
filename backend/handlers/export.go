package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type ExportHandler struct {
	exportService *services.ExportService
	exportDir     string
}

func NewExportHandler(exportService *services.ExportService, exportDir string) *ExportHandler {
	return &ExportHandler{exportService, exportDir}
}

func (eh *ExportHandler) Post(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	export, err := eh.exportService.CreateExport(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusAccepted, resources.ExportStatusFromModel(export))
}

func (eh *ExportHandler) Get(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	export, err := eh.exportService.GetExport(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.ExportStatusFromModel(export))
}

func (eh *ExportHandler) DownloadGet(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	export, err := eh.exportService.GetDoneExport(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}
	exportFilePath := filepath.Clean(filepath.Join(eh.exportDir, export.FilePath.String))
	if !strings.HasPrefix(exportFilePath, filepath.Clean(eh.exportDir)+string(os.PathSeparator)) {
		return echo.NewHTTPError(http.StatusForbidden, "invalid export path")
	}

	f, err := os.Open(exportFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	filename := "stashsphere-export-" + export.CreatedAt.UTC().Format("2006-01-02-1504") + ".zip"
	c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	return c.Stream(http.StatusOK, "application/zip", f)
}
