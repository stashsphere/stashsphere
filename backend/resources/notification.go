package resources

import (
	"encoding/json"
	"time"

	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/notifications"
)

// Notification is one of the *Notification types below, discriminated by its
// `contentType` field.
type Notification interface {
	NotificationContentType() string
}

type NotificationBase struct {
	Id           string    `json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	Acknowledged bool      `json:"acknowledged"`
	ContentType  string    `json:"contentType"`
}

func (n NotificationBase) NotificationContentType() string { return n.ContentType }

type FriendRequestNotification struct {
	NotificationBase
	Content notifications.FriendRequest `json:"content"`
}

type FriendRequestReactionNotification struct {
	NotificationBase
	Content notifications.FriendRequestReaction `json:"content"`
}

type ThingSharedNotification struct {
	NotificationBase
	Content notifications.ThingShared `json:"content"`
}

type ListSharedNotification struct {
	NotificationBase
	Content notifications.ListShared `json:"content"`
}

type ThingsAddedToListNotification struct {
	NotificationBase
	Content notifications.ThingsAddedToList `json:"content"`
}

// unknownNotification passes through content of types this version does not
// know, so stored notifications keep being returned as before.
type unknownNotification struct {
	NotificationBase
	Content json.RawMessage `json:"content"`
}

type PaginatedNotifications struct {
	Notifications  []Notification `json:"notifications"`
	PerPage        uint64         `json:"perPage"`
	Page           uint64         `json:"page"`
	TotalPageCount uint64         `json:"totalPageCount"`
	TotalCount     uint64         `json:"totalCount"`
}

func NotificationFromModel(model *models.Notification) Notification {
	base := NotificationBase{
		Id:           model.ID,
		CreatedAt:    model.CreatedAt,
		Acknowledged: model.AcknowledgedAt.Valid,
		ContentType:  model.ContentType,
	}
	var notification Notification
	var err error
	switch model.ContentType {
	case notifications.NotifyFriendRequestSent:
		n := FriendRequestNotification{NotificationBase: base}
		err = json.Unmarshal(model.Content, &n.Content)
		notification = n
	case notifications.NotifyFriendRequestReaction:
		n := FriendRequestReactionNotification{NotificationBase: base}
		err = json.Unmarshal(model.Content, &n.Content)
		notification = n
	case notifications.NotifyThingShared:
		n := ThingSharedNotification{NotificationBase: base}
		err = json.Unmarshal(model.Content, &n.Content)
		notification = n
	case notifications.NotifyListShared:
		n := ListSharedNotification{NotificationBase: base}
		err = json.Unmarshal(model.Content, &n.Content)
		notification = n
	case notifications.NotifyThingsAddedToList:
		n := ThingsAddedToListNotification{NotificationBase: base}
		err = json.Unmarshal(model.Content, &n.Content)
		notification = n
	}
	if notification == nil || err != nil {
		return unknownNotification{NotificationBase: base, Content: json.RawMessage(model.Content)}
	}
	return notification
}

func NotificationsFromModelSlice(mNotifications models.NotificationSlice) []Notification {
	notifications := make([]Notification, len(mNotifications))
	for i, model := range mNotifications {
		notifications[i] = NotificationFromModel(model)
	}
	return notifications
}
