package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
)

type UserService struct {
	db                  *sql.DB
	inviteCode          string
	inviteRequired      bool
	gracePeriodMinutes  int
	notificationService *NotificationService
}

func NewUserService(db *sql.DB, inviteRequired bool, inviteCode string, gracePeriodMinutes int, notificationService *NotificationService) *UserService {
	return &UserService{
		db:                  db,
		inviteCode:          inviteCode,
		inviteRequired:      inviteRequired,
		gracePeriodMinutes:  gracePeriodMinutes,
		notificationService: notificationService,
	}
}

type CreateUserParams struct {
	Name                      string
	Email                     string
	Password                  string
	InviteCode                string
	SendEmailVerification     bool
}

func (us *UserService) CreateUser(ctx context.Context, params CreateUserParams) (*models.User, error) {
	if us.inviteRequired && params.InviteCode != us.inviteCode {
		return nil, utils.WrongInviteCodeError{}
	}

	passwordHash, err := operations.HashPassword(params.Password)
	if err != nil {
		return nil, err
	}

	userID, err := gonanoid.New()
	if err != nil {
		return nil, err
	}

	user := models.User{
		ID:           userID,
		Name:         params.Name,
		Email:        params.Email,
		PasswordHash: string(passwordHash),
	}

	err = user.Insert(ctx, us.db, boil.Infer())
	if err != nil {
		return nil, err
	}

	if params.SendEmailVerification {
		if err := us.RequestEmailVerification(ctx, user.ID); err != nil {
			log.Error().Err(err).Str("userId", user.ID).Msg("Failed to send verification email on registration")
		}
	}

	return &user, nil
}

func (us *UserService) FindUserByID(ctx context.Context, userId string) (*models.User, error) {
	return operations.FindUserWithProfileByID(ctx, us.db, userId)
}

type UpdateUserParams struct {
	UserId      string
	Name        string
	FullName    string
	Information string
	ImageId     *string
}

func (us *UserService) UpdateUser(ctx context.Context, params UpdateUserParams) (*models.User, error) {
	err := utils.Tx(ctx, us.db, func(tx *sql.Tx) error {
		user, err := operations.FindUserWithProfileByID(ctx, tx, params.UserId)
		if err != nil {
			return err
		}
		user.Name = params.Name
		_, err = user.Update(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}
		var profile *models.Profile
		var insertProfile bool
		if user.R.Profile != nil {
			profile = user.R.Profile
			insertProfile = false
		} else {
			profileId, err := gonanoid.New()
			if err != nil {
				return err
			}
			profile = &models.Profile{
				ID: profileId,
			}
			insertProfile = true
		}
		profile.FullName = params.FullName
		profile.Information = params.Information
		if params.ImageId != nil {
			res, err := operations.ImageBelongsToUser(ctx, tx, params.UserId, *params.ImageId)
			if err != nil {
				return err
			}
			if !res {
				return utils.EntityDoesNotBelongToUserError{}
			}
			profile.ImageID = null.NewString(*params.ImageId, true)
		} else {
			profile.ImageID = null.NewString("", false)
		}
		if insertProfile {
			err = user.SetProfile(ctx, tx, insertProfile, profile)
			if err != nil {
				return err
			}
		} else {
			_, err := profile.Update(ctx, tx, boil.Infer())
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return us.FindUserByID(ctx, params.UserId)
}

type UpdatePasswordParams struct {
	UserId      string
	OldPassword string
	NewPassword string
}

func (us *UserService) UpdatePassword(ctx context.Context, params UpdatePasswordParams) error {
	err := utils.Tx(ctx, us.db, func(tx *sql.Tx) error {
		user, err := operations.AuthenticateUserByID(ctx, tx, params.UserId, params.OldPassword)
		if err != nil {
			return utils.ParameterError{Err: errors.New("Incorrect old password.")}
		}
		passwordHash, err := operations.HashPassword(params.NewPassword)
		if err != nil {
			return err
		}
		user.PasswordHash = string(passwordHash)
		_, err = user.Update(ctx, tx, boil.Infer())
		return err
	})
	return err
}

func (us *UserService) GetAllUsers(ctx context.Context) (models.UserSlice, error) {
	users, err := models.Users(qm.Load(models.UserRels.Profile),
		qm.Load(qm.Rels(models.UserRels.Profile, models.ProfileRels.Image)),
		qm.Load(qm.Rels(models.UserRels.Profile, models.ProfileRels.Image, models.ImageRels.Owner)),
	).All(ctx, us.db)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (us *UserService) ScheduleDeletion(ctx context.Context, userId string, password string) (*models.User, error) {
	_, err := operations.AuthenticateUserByID(ctx, us.db, userId, password)
	if err != nil {
		return nil, utils.ParameterError{Err: errors.New("Incorrect password.")}
	}

	existing, err := operations.FindUserByID(ctx, us.db, userId)
	if err != nil {
		return nil, err
	}
	if existing.PurgeAt.Valid {
		return us.FindUserByID(ctx, userId)
	}

	purgeAt := time.Now().UTC().Add(time.Duration(us.gracePeriodMinutes) * time.Minute)

	user, err := operations.ScheduleUserDeletion(ctx, us.db, userId, purgeAt)
	if err != nil {
		return nil, err
	}

	err = us.notificationService.AccountDeletionScheduled(ctx, AccountDeletionScheduledParams{
		UserId:    userId,
		UserName:  user.Name,
		UserEmail: user.Email,
		PurgeAt:   purgeAt,
	})
	if err != nil {
		return nil, err
	}

	return us.FindUserByID(ctx, userId)
}

func (us *UserService) CancelDeletion(ctx context.Context, userId string) (*models.User, error) {
	user, err := operations.FindUserByID(ctx, us.db, userId)
	if err != nil {
		return nil, err
	}

	if !user.PurgeAt.Valid {
		return nil, utils.NotFoundError{EntityName: "scheduled deletion"}
	}

	user.PurgeAt = null.Time{}
	_, err = user.Update(ctx, us.db, boil.Whitelist(models.UserColumns.PurgeAt))
	if err != nil {
		return nil, err
	}

	return us.FindUserByID(ctx, userId)
}

func (us *UserService) RequestEmailVerification(ctx context.Context, userId string) error {
	user, err := operations.FindUserByID(ctx, us.db, userId)
	if err != nil {
		return err
	}

	validUntil := time.Now().Add(30 * time.Minute)
	code, err := operations.CreateVerificationCode(ctx, us.db, user.ID, user.Email, validUntil)
	if err != nil {
		return err
	}

	return us.notificationService.EmailVerification(ctx, EmailVerificationParams{
		UserName:  user.Name,
		UserEmail: user.Email,
		DigitCode: code,
	})
}

func (us *UserService) VerifyEmail(ctx context.Context, userId string, email string, code string) error {
	return utils.Tx(ctx, us.db, func(tx *sql.Tx) error {
		return operations.VerifyCode(ctx, tx, userId, email, code)
	})
}

func (us *UserService) GetEmailVerificationStatus(ctx context.Context, userId string) (*models.EmailVerification, error) {
	return operations.GetEmailVerificationStatus(ctx, us.db, userId)
}

