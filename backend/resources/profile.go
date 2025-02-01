package resources

import (
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/models"
)

type Profile struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func ProfileFromUserContext(ctx *middleware.UserContext) Profile {
	return Profile{
		ID:    ctx.ID,
		Name:  ctx.Name,
		Email: ctx.Email,
	}
}

func ProfileFromModel(user *models.User) Profile {
	return Profile{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}

func ProfilesFromModelSlice(mProfiles models.UserSlice) []Profile {
	profiles := make([]Profile, len(mProfiles))
	for i, profile := range mProfiles {
		profiles[i] = ProfileFromModel(profile)
	}
	return profiles
}
