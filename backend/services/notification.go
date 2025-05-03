package services

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/notifications"
	"github.com/stashsphere/backend/utils"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type NotificationService struct {
	db *sql.DB
}

func NewNotificationService(db *sql.DB) *NotificationService {
	return &NotificationService{db}
}

type CreateNotification struct {
	RecipientId string
	Content     notifications.StashsphereNotification
}

func (ns *NotificationService) CreateNotification(ctx context.Context, params CreateNotification) (*models.Notification, error) {
	notificationId, err := gonanoid.New()
	if err != nil {
		return nil, err
	}
	notification := models.Notification{
		ID:          notificationId,
		RecipientID: params.RecipientId,
		ContentType: params.Content.ContentType(),
	}
	err = notification.Content.Marshal(params.Content)
	if err != nil {
		return nil, err
	}
	err = notification.Insert(ctx, ns.db, boil.Infer())
	if err != nil {
		return nil, err
	}
	// TODO send out notification using mail/smtp
	err = notification.Reload(ctx, ns.db)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

type GetNotificationsForUserParams struct {
	UserId             string
	PerPage            uint64
	Page               uint64
	Paginate           bool
	OnlyUnacknowledged bool
}

func (ns *NotificationService) GetNotifications(ctx context.Context, params GetNotificationsForUserParams) (uint64, uint64, models.NotificationSlice, error) {
	userId, perPage, page, paginate, onlyUnacknowledged := params.UserId, params.PerPage, params.Page, params.Paginate, params.OnlyUnacknowledged

	searchCond := []qm.QueryMod{
		models.NotificationWhere.RecipientID.EQ(userId),
	}
	if onlyUnacknowledged {
		searchCond = append(searchCond, models.NotificationWhere.AcknowledgedAt.IsNull())
	}

	notificationCount, err := models.Notifications(searchCond...).Count(ctx, ns.db)
	if err != nil {
		return 0, 0, models.NotificationSlice{}, err
	}

	notificationQuery := []qm.QueryMod{
		qm.OrderBy(`created_at desc`),
	}
	if paginate {
		notificationQuery = append(notificationQuery, qm.Offset(int(perPage*page)), qm.Limit(int(perPage)))
	}
	for _, s := range searchCond {
		notificationQuery = append(notificationQuery, s)
	}
	notifications, err := models.Notifications(notificationQuery...).All(ctx, ns.db)
	if err != nil {
		return 0, 0, models.NotificationSlice{}, err
	}
	totalPages := uint64(math.Ceil(float64(notificationCount) / float64(perPage)))
	return uint64(notificationCount), totalPages, notifications, nil
}

type AcknowledgeNotificationParams struct {
	UserId         string
	NotificationId string
}

func (ns *NotificationService) AcknowledgeNotification(ctx context.Context, params AcknowledgeNotificationParams) error {
	notification, err := models.Notifications(models.NotificationWhere.ID.EQ(params.NotificationId)).One(ctx, ns.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.NotFoundError{EntityName: "Notification"}
		}
		return err
	}
	if notification.RecipientID != params.UserId {
		return utils.EntityDoesNotBelongToUserError{}
	}
	notification.AcknowledgedAt = null.NewTime(time.Now(), true)
	_, err = notification.Update(ctx, ns.db, boil.Whitelist(models.NotificationColumns.AcknowledgedAt))
	return err
}
