package utils

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	ErrInventoryValidation        = "inventory-validation"
	ErrParameterError             = "parameter-error"
	ErrNotFoundError              = "not-found"
	ErrWrongInviteCode            = "wrong-invite-code"
	ErrEntityDoesNotBelongToUser  = "entity-does-not-belong-to-user"
	ErrUserHasNoAccessRights      = "user-has-no-access-rights"
	ErrEntityInUse                = "entity-in-use"
	ErrFriendRequestNotPending    = "friend-request-not-pending"
	ErrNoAuthContext              = "no-auth-context"
	ErrNotAuthenticated           = "not_authenticated"
	ErrIllegalMimeType            = "illegal-mime-type"
	ErrPendingFriendRequestExists = "pending-friend-request-exists"
	ErrFriendShipExists           = "friend-ship-exists"
)

type StashsphereError interface {
	Error() string
	ErrorType() string
}

type StashsphereValidationError struct {
	Errors map[string]string
}

func (ie StashsphereValidationError) Error() string {
	buff := bytes.NewBufferString("")

	for k, v := range ie.Errors {
		buff.WriteString(k)
		buff.WriteString(": ")
		buff.WriteString(v)
		buff.WriteString("\n")
	}

	return strings.TrimSpace(buff.String())

}

type ParameterError struct {
	Err error
}

func (r ParameterError) Error() string {
	return fmt.Sprintf("ParameterError: %v", r.Err)
}

type NotFoundError struct {
	EntityName string
}

func (r NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", r.EntityName)
}

type WrongInviteCodeError struct{}

func (r WrongInviteCodeError) ErrorType() string { return ErrWrongInviteCode }
func (r WrongInviteCodeError) Error() string     { return "Invalid invite code" }

type EntityDoesNotBelongToUserError struct{}

func (r EntityDoesNotBelongToUserError) ErrorType() string { return ErrEntityDoesNotBelongToUser }
func (r EntityDoesNotBelongToUserError) Error() string     { return "Entity does not belong to user" }

type UserHasNoAccessRightsError struct{}

func (r UserHasNoAccessRightsError) ErrorType() string { return ErrUserHasNoAccessRights }
func (r UserHasNoAccessRightsError) Error() string     { return "User has no access rights" }

type EntityInUseError struct{}

func (r EntityInUseError) ErrorType() string { return ErrEntityInUse }
func (r EntityInUseError) Error() string     { return "Entity is in use" }

type FriendRequestNotPendingError struct{}

func (r FriendRequestNotPendingError) ErrorType() string { return ErrFriendRequestNotPending }
func (r FriendRequestNotPendingError) Error() string     { return "Friend request is not pending" }

type NoAuthContextError struct{}

func (r NoAuthContextError) ErrorType() string { return ErrNoAuthContext }
func (r NoAuthContextError) Error() string     { return "No authentication context found" }

type NotAuthenticatedError struct{}

func (r NotAuthenticatedError) ErrorType() string { return ErrNotAuthenticated }
func (r NotAuthenticatedError) Error() string     { return "User is not authenticated" }

type IllegalMimeTypeError struct{}

func (r IllegalMimeTypeError) ErrorType() string { return ErrIllegalMimeType }
func (r IllegalMimeTypeError) Error() string     { return "Invalid MIME type" }

type PendingFriendRequestExistsError struct{}

func (r PendingFriendRequestExistsError) ErrorType() string { return ErrPendingFriendRequestExists }
func (r PendingFriendRequestExistsError) Error() string     { return "Pending friend request exists" }

type FriendShipExistsError struct{}

func (r FriendShipExistsError) ErrorType() string { return ErrFriendShipExists }
func (r FriendShipExistsError) Error() string     { return "Friendship already exists" }
