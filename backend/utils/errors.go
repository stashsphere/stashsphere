package utils

import (
	"bytes"
	"errors"
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

var (
	ErrWrongInviteCode           = errors.New("WrongInviteCode")
	ErrEntityDoesNotBelongToUser = errors.New("EntityDoesNotBelongToUser")
	ErrUserHasNoAccessRights     = errors.New("UserHasNoAccessRights")
	ErrEntityInUse               = errors.New("EntityInUse")
	ErrFriendRequestNotPending   = errors.New("FriendRequestNotPending")
	// file is not of the allowed mime types
	ErrIllegalMimeType = errors.New("IllegalMimeType")
)
