package resources

import (
	"time"

	"github.com/stashsphere/backend/models"
)

type Notification struct {
	Id           string      `json:"id"`
	CreatedAt    time.Time   `json:"createdAt"`
	Acknowledged bool        `json:"acknowledged"`
	ContentType  string      `json:"contentType"`
	Content      interface{} `json:"content"`
}

type PaginatedNotifications struct {
	Notifications  []Notification `json:"notifications"`
	PerPage        uint64         `json:"perPage"`
	Page           uint64         `json:"page"`
	TotalPageCount uint64         `json:"totalPageCount"`
	TotalCount     uint64         `json:"totalCount"`
}

func NotificationFromModel(model *models.Notification) *Notification {
	return &Notification{
		Id:           model.ID,
		CreatedAt:    model.CreatedAt,
		Acknowledged: model.AcknowledgedAt.Valid,
		ContentType:  model.ContentType,
		Content:      model.Content,
	}
}

func NotificationsFromModelSlice(mNotifications models.NotificationSlice) []Notification {
	notifications := make([]Notification, len(mNotifications))
	for i, model := range mNotifications {
		notifications[i] = *NotificationFromModel(model)
	}
	return notifications
}
