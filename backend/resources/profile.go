package resources

import (
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/models"
)

type Profile struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	FullName    *string       `json:"fullName"`
	Information *string       `json:"information"`
	Email       string        `json:"email"`
	Image       *ReducedImage `json:"image"`
}

func ProfileFromUserContext(ctx *middleware.UserContext) Profile {
	return Profile{
		ID:    ctx.UserId,
		Name:  ctx.Name,
		Email: ctx.Email,
	}
}

func ProfileFromModel(user *models.User) Profile {
	var image *ReducedImage
	var information *string
	var fullName *string
	if user.R.Profile != nil {
		if user.R.Profile.R.Image != nil {
			reduced := ReducedImageFromModel(user.R.Profile.R.Image)
			image = &reduced
		}
		fullName = &user.R.Profile.FullName
		information = &user.R.Profile.Information
	}

	return Profile{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Image:       image,
		FullName:    fullName,
		Information: information,
	}
}

func ProfilesFromModelSlice(mProfiles models.UserSlice) []Profile {
	profiles := make([]Profile, len(mProfiles))
	for i, profile := range mProfiles {
		profiles[i] = ProfileFromModel(profile)
	}
	return profiles
}
