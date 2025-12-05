package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService,
	}
}

type NotificationParams struct {
	OnlyUnacknowledged bool   `query:"onlyUnacknowledged"`
	Page               uint64 `query:"page"`
	PerPage            uint64 `query:"perPage"`
}

func (nh *NotificationHandler) Index(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	var params NotificationParams
	if err := c.Bind(&params); err != nil {
		return &utils.ParameterError{Err: err}
	}
	if params.PerPage == 0 {
		params.PerPage = 50
	}
	totalCount, totalPageCount, notifications, err := nh.notificationService.GetNotifications(c.Request().Context(), services.GetNotificationsForUserParams{
		UserId:             authCtx.User.UserId,
		PerPage:            params.PerPage,
		Page:               params.Page,
		Paginate:           true,
		OnlyUnacknowledged: params.OnlyUnacknowledged,
	})
	if err != nil {
		return err
	}
	paginated := resources.PaginatedNotifications{
		Notifications:  resources.NotificationsFromModelSlice(notifications),
		PerPage:        uint64(params.PerPage),
		Page:           uint64(params.Page),
		TotalPageCount: totalPageCount,
		TotalCount:     totalCount,
	}
	return c.JSON(http.StatusOK, paginated)
}

func (nh *NotificationHandler) Acknowledge(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	notificationId := c.Param("notificationId")
	err := nh.notificationService.AcknowledgeNotification(c.Request().Context(), services.AcknowledgeNotificationParams{
		NotificationId: notificationId,
		UserId:         authCtx.User.UserId,
	})
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
