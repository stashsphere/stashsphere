package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/utils"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func CreateStashSphereHTTPErrorHandler(echoInstance *echo.Echo) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		statusCode := http.StatusInternalServerError
		message := "Internal Server Error"

		// Check if the error implements ErrorInterface
		switch e := err.(type) {
		case utils.StashsphereError:
			switch e.ErrorType() {
			case utils.ErrStashSphereValidation:
				statusCode = http.StatusBadRequest
				message = e.Error()
			case utils.ErrParameterError:
				statusCode = http.StatusBadRequest
				message = e.Error()
			case utils.ErrNotFoundError:
				statusCode = http.StatusNotFound
				message = e.Error()
			case utils.ErrWrongInviteCode:
				statusCode = http.StatusBadRequest
				message = "Invalid invite code"
			case utils.ErrEntityDoesNotBelongToUser:
				statusCode = http.StatusForbidden
				message = "Entity does not belong to user"
			case utils.ErrUserHasNoAccessRights:
				statusCode = http.StatusForbidden
				message = "Insufficient permissions"
			case utils.ErrEntityInUse:
				statusCode = http.StatusBadRequest
				message = "Entity is currently in use"
			case utils.ErrFriendRequestNotPending:
				statusCode = http.StatusBadRequest
				message = "Friend request not pending"
			case utils.ErrNoAuthContext, utils.ErrNotAuthenticated:
				statusCode = http.StatusUnauthorized
				message = "Authentication required"
			case utils.ErrIllegalMimeType:
				statusCode = http.StatusBadRequest
				message = "Invalid file type"
			case utils.ErrPendingFriendRequestExists:
				statusCode = http.StatusConflict
				message = "A pending request already exists"
			}
		default:
			echoInstance.DefaultHTTPErrorHandler(err, c)
			return
		}

		// Construct the error response
		response := &ErrorResponse{
			Code:    statusCode,
			Message: message,
		}

		// Send the JSON response with appropriate status code
		if err := c.JSON(statusCode, response); err != nil {
			// Fallback if JSON encoding fails
			echoInstance.DefaultHTTPErrorHandler(err, c)
		}
		return
	}
}
