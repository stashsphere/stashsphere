package resources

import "time"
import "github.com/stashsphere/backend/models"

type ExportStatus struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt *time.Time `json:"expiresAt"`
	Error     *string    `json:"error"`
}

func ExportStatusFromModel(e *models.Export) ExportStatus {
	return ExportStatus{
		ID:        e.ID,
		Status:    string(e.Status),
		CreatedAt: e.CreatedAt,
		ExpiresAt: e.ExpiresAt.Ptr(),
		Error:     e.ErrorMSG.Ptr(),
	}
}
