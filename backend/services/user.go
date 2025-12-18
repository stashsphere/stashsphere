package services

import (
	"context"
	"database/sql"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
)

type UserService struct {
	db             *sql.DB
	inviteCode     string
	inviteRequired bool
}

func NewUserService(db *sql.DB, inviteRequired bool, inviteCode string) *UserService {
	return &UserService{db, inviteCode, inviteRequired}
}

type CreateUserParams struct {
	Name       string
	Email      string
	Password   string
	InviteCode string
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
