package utils

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	ErrStashSphereValidation      = "stashsphere-validation"
	ErrParameterError             = "parameter-error"
	ErrNotFoundError              = "not-found"
	ErrWrongInviteCode            = "wrong-invite-code"
	ErrEntityDoesNotBelongToUser  = "entity-does-not-belong-to-user"
	ErrUserHasNoAccessRights      = "user-has-no-access-rights"
	ErrEntityInUse                = "entity-in-use"
	ErrFriendRequestNotPending    = "friend-request-not-pending"
	ErrNoAuthContext              = "no-auth-context"
	ErrNotAuthenticated           = "not-authenticated"
	ErrIllegalMimeType            = "illegal-mime-type"
	ErrPendingFriendRequestExists  = "pending-friend-request-exists"
	ErrFriendShipExists            = "friend-ship-exists"
	ErrInvalidVerificationCode     = "invalid-verification-code"
	ErrVerificationCodeExpired     = "verification-code-expired"
	ErrOIDCProviderNotFound        = "oidc-provider-not-found"
	ErrOIDCCallbackFailed          = "oidc-callback-failed"
	ErrOIDCLinkChallengExpired     = "oidc-link-challenge-expired"
	ErrOIDCLinkIncorrectPassword   = "oidc-link-incorrect-password"
)

type StashsphereError interface {
	Error() string
	ErrorType() string
}

type StashSphereValidationError struct {
	Errors map[string]string
}

func (ie StashSphereValidationError) Error() string {
	buff := bytes.NewBufferString("")

	for k, v := range ie.Errors {
		buff.WriteString(k)
		buff.WriteString(": ")
		buff.WriteString(v)
		buff.WriteString("\n")
	}

	return strings.TrimSpace(buff.String())

}

func (ie StashSphereValidationError) ErrorType() string { return ErrStashSphereValidation }

type ParameterError struct {
	Err error
}

func (r ParameterError) Error() string {
	return fmt.Sprintf("ParameterError: %v", r.Err)
}

func (r ParameterError) ErrorType() string { return ErrParameterError }

type NotFoundError struct {
	EntityName string
}

func (r NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", r.EntityName)
}

func (r NotFoundError) ErrorType() string { return ErrNotFoundError }

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

type InvalidVerificationCodeError struct{}

func (r InvalidVerificationCodeError) ErrorType() string { return ErrInvalidVerificationCode }
func (r InvalidVerificationCodeError) Error() string     { return "Invalid verification code" }

type VerificationCodeExpiredError struct{}

func (r VerificationCodeExpiredError) ErrorType() string { return ErrVerificationCodeExpired }
func (r VerificationCodeExpiredError) Error() string     { return "Verification code has expired" }

type OIDCProviderNotFoundError struct{}

func (r OIDCProviderNotFoundError) ErrorType() string { return ErrOIDCProviderNotFound }
func (r OIDCProviderNotFoundError) Error() string     { return "OIDC provider not found" }

type OIDCCallbackFailedError struct {
	Err error
}

func (r OIDCCallbackFailedError) ErrorType() string { return ErrOIDCCallbackFailed }
func (r OIDCCallbackFailedError) Error() string     { return fmt.Sprintf("OIDC callback failed: %v", r.Err) }

type OIDCLinkChallengeExpiredError struct{}

func (r OIDCLinkChallengeExpiredError) ErrorType() string { return ErrOIDCLinkChallengExpired }
func (r OIDCLinkChallengeExpiredError) Error() string     { return "Link challenge has expired" }

type OIDCLinkIncorrectPasswordError struct{}

func (r OIDCLinkIncorrectPasswordError) ErrorType() string { return ErrOIDCLinkIncorrectPassword }
func (r OIDCLinkIncorrectPasswordError) Error() string     { return "Incorrect password" }
