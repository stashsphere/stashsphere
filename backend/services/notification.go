package services

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"html/template"
	"math"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/notifications"
	"github.com/stashsphere/backend/notifications/templates"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
)

type NotificationData struct {
	FrontendUrl  string
	InstanceName string
}

type NotificationService struct {
	db           *sql.DB
	data         NotificationData
	emailService EmailService
}

func NewNotificationService(db *sql.DB, data NotificationData, emailService EmailService) *NotificationService {
	return &NotificationService{db, data, emailService}
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

type CreateFriendRequestNotificationParams struct {
	ReceiverId string
	RequestId  string
	SenderId   string
}

type createFriendRequestNotificationParamsFull struct {
	ReceiverId    string
	ReceiverName  string
	ReceiverEmail string
	RequestId     string
	SenderId      string
}

func (ns *NotificationService) CreateFriendRequest(ctx context.Context, params CreateFriendRequestNotificationParams) error {
	receiver, err := operations.FindUserByID(ctx, ns.db, params.ReceiverId)
	if err != nil {
		return err
	}
	return ns.createFriendRequest(ctx, createFriendRequestNotificationParamsFull{
		ReceiverId:    params.ReceiverId,
		ReceiverName:  receiver.Name,
		ReceiverEmail: receiver.Email,
		RequestId:     params.RequestId,
		SenderId:      params.SenderId,
	})
}

func (ns *NotificationService) createFriendRequest(ctx context.Context, params createFriendRequestNotificationParamsFull) error {
	_, err := ns.CreateNotification(ctx, CreateNotification{
		RecipientId: params.ReceiverId,
		Content: notifications.FriendRequest{
			RequestId: params.RequestId,
			SenderId:  params.SenderId,
		},
	})
	if err != nil {
		return err
	}

	bodyTempl, err := template.ParseFS(templates.FS, "friend_request.body.txt")
	if err != nil {
		return err
	}

	subjectTempl, err := template.ParseFS(templates.FS, "friend_request.subject.txt")
	if err != nil {
		return err
	}

	type BodyData struct {
		RecipientName string
		FrontendUrl   string
	}

	type SubjectData struct {
		InstanceName string
	}

	var body bytes.Buffer
	err = bodyTempl.Execute(&body, BodyData{
		RecipientName: params.ReceiverName,
		FrontendUrl:   ns.data.FrontendUrl,
	})
	if err != nil {
		return err
	}

	var subject bytes.Buffer
	err = subjectTempl.Execute(&subject, SubjectData{
		InstanceName: ns.data.InstanceName,
	})
	if err != nil {
		return err
	}

	return ns.emailService.Deliver(params.ReceiverEmail, subject.String(), body.String())
}

type CreateFriendRequestReactionParams struct {
	RequestId  string
	ReceiverId string
	SenderId   string
	Accepted   bool
}

type createFriendRequestReactionParamsFull struct {
	RequestId    string
	ReceiverId   string
	SenderId     string
	SenderEmail  string
	ReceiverName string
	Accepted     bool
	SenderName   string
}

func (ns *NotificationService) CreateFriendRequestReaction(ctx context.Context, params CreateFriendRequestReactionParams) error {
	receiver, err := operations.FindUserByID(ctx, ns.db, params.ReceiverId)
	if err != nil {
		return err
	}
	sender, err := operations.FindUserByID(ctx, ns.db, params.SenderId)
	if err != nil {
		return err
	}
	return ns.createFriendRequestReaction(ctx, createFriendRequestReactionParamsFull{
		RequestId:    params.RequestId,
		ReceiverId:   params.ReceiverId,
		SenderId:     params.SenderId,
		SenderEmail:  sender.Email,
		ReceiverName: receiver.Name,
		Accepted:     params.Accepted,
		SenderName:   sender.Name,
	})
}

func (ns *NotificationService) createFriendRequestReaction(ctx context.Context, params createFriendRequestReactionParamsFull) error {
	_, err := ns.CreateNotification(ctx, CreateNotification{
		RecipientId: params.ReceiverId,
		Content: notifications.FriendRequestReaction{
			RequestId: params.RequestId,
			Accepted:  params.Accepted,
		},
	})
	if err != nil {
		return err
	}

	_, err = ns.CreateNotification(ctx, CreateNotification{
		RecipientId: params.SenderId,
		Content: notifications.FriendRequestReaction{
			RequestId: params.RequestId,
			Accepted:  params.Accepted,
		},
	})
	if err != nil {
		return err
	}

	bodyTempl, err := template.ParseFS(templates.FS, "friend_request_reaction.body.txt")
	if err != nil {
		return err
	}

	subjectTempl, err := template.ParseFS(templates.FS, "friend_request_reaction.subject.txt")
	if err != nil {
		return err
	}

	type BodyData struct {
		Accepted      bool
		RecipientName string
		SenderName    string
	}

	type SubjectData struct {
		Accepted     bool
		InstanceName string
	}

	var body bytes.Buffer
	err = bodyTempl.Execute(&body, BodyData{
		RecipientName: params.ReceiverName,
		SenderName:    params.SenderName,
		Accepted:      params.Accepted,
	})
	if err != nil {
		return err
	}

	var subject bytes.Buffer
	err = subjectTempl.Execute(&subject, SubjectData{
		InstanceName: ns.data.InstanceName,
		Accepted:     params.Accepted,
	})
	if err != nil {
		return err
	}
	return ns.emailService.Deliver(params.SenderEmail, subject.String(), body.String())
}

type ThingSharedParams struct {
	ThingId      string
	SharerId     string
	TargetUserId string
}

type thingSharedParamsFull struct {
	ThingId         string
	SharerName      string
	SharerId        string
	TargetUserId    string
	TargetUserName  string
	TargetUserEmail string
}

func (ns *NotificationService) ThingShared(ctx context.Context, params ThingSharedParams) error {
	sharer, err := operations.FindUserByID(ctx, ns.db, params.SharerId)
	if err != nil {
		return err
	}
	targetUser, err := operations.FindUserByID(ctx, ns.db, params.TargetUserId)
	if err != nil {
		return err
	}
	return ns.thingShared(ctx, thingSharedParamsFull{
		ThingId:         params.ThingId,
		SharerName:      sharer.Name,
		SharerId:        params.SharerId,
		TargetUserId:    params.TargetUserId,
		TargetUserName:  targetUser.Name,
		TargetUserEmail: targetUser.Email,
	})
}

func (ns *NotificationService) thingShared(ctx context.Context, params thingSharedParamsFull) error {
	_, err := ns.CreateNotification(ctx, CreateNotification{
		RecipientId: params.TargetUserId,
		Content: notifications.ThingShared{
			ThingId:  params.ThingId,
			SharerId: params.SharerId,
		},
	})
	if err != nil {
		return err
	}

	bodyTempl, err := template.ParseFS(templates.FS, "thing_shared.body.txt")
	if err != nil {
		return err
	}

	subjectTempl, err := template.ParseFS(templates.FS, "thing_shared.subject.txt")
	if err != nil {
		return err
	}

	type BodyData struct {
		TargetUserName string
		SharerName     string
		FrontendUrl    string
	}

	type SubjectData struct {
		InstanceName string
	}

	var body bytes.Buffer
	err = bodyTempl.Execute(&body, BodyData{
		TargetUserName: params.TargetUserName,
		SharerName:     params.SharerName,
		FrontendUrl:    ns.data.FrontendUrl,
	})
	if err != nil {
		return err
	}

	var subject bytes.Buffer
	err = subjectTempl.Execute(&subject, SubjectData{
		InstanceName: ns.data.InstanceName,
	})
	if err != nil {
		return err
	}
	return ns.emailService.Deliver(params.TargetUserEmail, subject.String(), body.String())
}

type ListSharedParams struct {
	ListId       string
	SharedId     string
	TargetUserId string
}

type listSharedParamsFull struct {
	ListId          string
	SharerName      string
	SharedId        string
	TargetUserId    string
	TargetUserName  string
	TargetUserEmail string
}

func (ns *NotificationService) ListShared(ctx context.Context, params ListSharedParams) error {
	sharer, err := operations.FindUserByID(ctx, ns.db, params.SharedId)
	if err != nil {
		return err
	}
	targetUser, err := operations.FindUserByID(ctx, ns.db, params.TargetUserId)
	if err != nil {
		return err
	}
	return ns.listShared(ctx, listSharedParamsFull{
		ListId:          params.ListId,
		SharerName:      sharer.Name,
		SharedId:        sharer.ID,
		TargetUserName:  targetUser.Name,
		TargetUserId:    params.TargetUserId,
		TargetUserEmail: targetUser.Email,
	})
}

func (ns *NotificationService) listShared(ctx context.Context, params listSharedParamsFull) error {
	_, err := ns.CreateNotification(ctx, CreateNotification{
		RecipientId: params.TargetUserId,
		Content: notifications.ListShared{
			ListId:   params.ListId,
			SharerId: params.SharedId,
		},
	})
	if err != nil {
		return err
	}

	bodyTempl, err := template.ParseFS(templates.FS, "list_shared.body.txt")
	if err != nil {
		return err
	}

	subjectTempl, err := template.ParseFS(templates.FS, "list_shared.subject.txt")
	if err != nil {
		return err
	}

	type BodyData struct {
		TargetUserName string
		SharerName     string
		FrontendUrl    string
	}

	type SubjectData struct {
		InstanceName string
	}

	var body bytes.Buffer
	err = bodyTempl.Execute(&body, BodyData{
		TargetUserName: params.TargetUserName,
		SharerName:     params.SharerName,
		FrontendUrl:    ns.data.FrontendUrl,
	})
	if err != nil {
		return err
	}

	var subject bytes.Buffer
	err = subjectTempl.Execute(&subject, SubjectData{
		InstanceName: ns.data.InstanceName,
	})
	if err != nil {
		return err
	}
	return ns.emailService.Deliver(params.TargetUserEmail, subject.String(), body.String())
}

type ThingsAddedToListParams struct {
	ListId       string
	OwnerId      string
	TargetUserId string
}

func (ns *NotificationService) ThingsAddedToList(ctx context.Context, params ThingsAddedToListParams) error {
	owner, err := operations.FindUserByID(ctx, ns.db, params.OwnerId)
	if err != nil {
		return err
	}
	targetUser, err := operations.FindUserByID(ctx, ns.db, params.TargetUserId)
	if err != nil {
		return err
	}
	return ns.thingsAddedToList(ctx, thingsAddedToListParamsFull{
		OwnerId:         params.OwnerId,
		ListId:          params.ListId,
		OwnerName:       owner.Name,
		TargetUserId:    params.TargetUserId,
		TargetUserName:  targetUser.Name,
		TargetUserEmail: targetUser.Email,
	})
}

type thingsAddedToListParamsFull struct {
	ListId          string
	OwnerId         string
	OwnerName       string
	TargetUserId    string
	TargetUserName  string
	TargetUserEmail string
}

func (ns *NotificationService) thingsAddedToList(ctx context.Context, params thingsAddedToListParamsFull) error {
	_, err := ns.CreateNotification(ctx, CreateNotification{
		RecipientId: params.TargetUserId,
		Content: notifications.ThingsAddedToList{
			ListId:    params.ListId,
			AddedById: params.OwnerId,
		},
	})
	if err != nil {
		return err
	}

	bodyTempl, err := template.ParseFS(templates.FS, "things_added_to_list.body.txt")
	if err != nil {
		return err
	}

	subjectTempl, err := template.ParseFS(templates.FS, "things_added_to_list.subject.txt")
	if err != nil {
		return err
	}

	type BodyData struct {
		TargetUserName string
		OwnerName      string
		FrontendUrl    string
	}

	type SubjectData struct {
		InstanceName string
		OwnerName    string
	}

	var body bytes.Buffer
	err = bodyTempl.Execute(&body, BodyData{
		TargetUserName: params.TargetUserName,
		FrontendUrl:    ns.data.FrontendUrl,
		OwnerName:      params.OwnerName,
	})
	if err != nil {
		return err
	}

	var subject bytes.Buffer
	err = subjectTempl.Execute(&subject, SubjectData{
		InstanceName: ns.data.InstanceName,
		OwnerName:    params.OwnerName,
	})
	if err != nil {
		return err
	}

	return ns.emailService.Deliver(params.TargetUserEmail, subject.String(), body.String())
}
