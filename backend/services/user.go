package services

import (
	"context"
	"database/sql"
	"errors"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
	"github.com/volatiletech/sqlboiler/v4/boil"
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
	return operations.FindUserByID(ctx, us.db, userId)
}

func (us *UserService) UpdateUser(ctx context.Context, userId string, name string) (*models.User, error) {
	user, err := models.Users(models.UserWhere.ID.EQ(userId)).One(ctx, us.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "user"}
		}
		return nil, err
	}
	user.Name = name
	_, err = user.Update(ctx, us.db, boil.Infer())
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (us *UserService) GetAllUsers(ctx context.Context) (models.UserSlice, error) {
	users, err := models.Users().All(ctx, us.db)
	if err != nil {
		return nil, err
	}
	return users, nil
}
