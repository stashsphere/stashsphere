package resources

import (
	"github.com/stashsphere/backend/models"
)

type UserProfile struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	FullName    *string       `json:"fullName"`
	Information *string       `json:"information"`
	Image       *ReducedImage `json:"image"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func UserFromModel(user *models.User) User {
	return User{
		ID:   user.ID,
		Name: user.Name,
	}
}

func UsersFromModelSlice(mUsers models.UserSlice) []User {
	users := make([]User, len(mUsers))
	for i, user := range mUsers {
		users[i] = UserFromModel(user)
	}
	return users
}

func UserProfileFromModel(user *models.User) UserProfile {
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

	return UserProfile{
		ID:          user.ID,
		Name:        user.Name,
		Image:       image,
		FullName:    fullName,
		Information: information,
	}
}

func UserProfilesFromModelSlice(mUsers models.UserSlice) []User {
	users := make([]User, len(mUsers))
	for i, user := range mUsers {
		users[i] = UserFromModel(user)
	}
	return users
}
