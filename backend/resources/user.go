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
