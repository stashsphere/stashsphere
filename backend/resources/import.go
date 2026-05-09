package resources

import (
	"time"

	"github.com/stashsphere/backend/models"
)

type ImportStatus struct {
	ID             string     `json:"id"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"createdAt"`
	CompletedAt    *time.Time `json:"completedAt"`
	Error          *string    `json:"error"`
	ThingsImported *int       `json:"thingsImported"`
	ListsImported  *int       `json:"listsImported"`
	ImagesImported *int       `json:"imagesImported"`
}

func ImportStatusFromModel(i *models.Import) ImportStatus {
	return ImportStatus{
		ID:             i.ID,
		Status:         string(i.Status),
		CreatedAt:      i.CreatedAt,
		CompletedAt:    i.CompletedAt.Ptr(),
		Error:          i.ErrorMSG.Ptr(),
		ThingsImported: i.ThingsImported.Ptr(),
		ListsImported:  i.ListsImported.Ptr(),
		ImagesImported: i.ImagesImported.Ptr(),
	}
}
