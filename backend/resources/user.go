package resources

import (
	"github.com/stashsphere/backend/models"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func UserFromModel(value *models.User) User {
	return User{
		ID:   value.ID,
		Name: value.Name,
	}
}

func UsersFromModelSlice(mUsers models.UserSlice) []User {
	users := make([]User, len(mUsers))
	for i, user := range mUsers {
		users[i] = UserFromModel(user)
	}
	return users
}
