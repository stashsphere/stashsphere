package utils

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type InventoryValidationError struct {
	Errors map[string]string
}

func (ie InventoryValidationError) Error() string {
	buff := bytes.NewBufferString("")

	for k, v := range ie.Errors {
		buff.WriteString(k)
		buff.WriteString(": ")
		buff.WriteString(v)
		buff.WriteString("\n")
	}

	return strings.TrimSpace(buff.String())

}

type ErrParameterError struct {
	Err error
}

func (r ErrParameterError) Error() string {
	return fmt.Sprintf("ParameterError: %v", r.Err)
}

type ErrNotFoundError struct {
	EntityName string
}

func (r ErrNotFoundError) Error() string {
	return fmt.Sprintf("%s not found", r.EntityName)
}

var (
	ErrWrongInviteCode           = errors.New("WrongInviteCode")
	ErrEntityDoesNotBelongToUser = errors.New("EntityDoesNotBelongToUser")
	ErrUserHasNoAccessRights     = errors.New("UserHasNoAccessRights")
	ErrEntityInUse               = errors.New("EntityInUse")
	ErrFriendRequestNotPending   = errors.New("FriendRequestNotPending")
	// internal error: no auth context has been found
	ErrNoAuthContext = errors.New("NoAuthContext")
	// the user is not authenticated, request w/o a token
	ErrNotAuthenticated = errors.New("NotAuthenticated")
	// file is not of the allowed mime types
	ErrIllegalMimeType            = errors.New("IllegalMimeType")
	ErrPendingFriendRequestExists = errors.New("PendingFriendRequestExists")
	ErrFriendShipExists           = errors.New("FriendshipExists")
)
