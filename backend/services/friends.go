package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type FriendService struct {
	db *sql.DB
	ns *NotificationService
}

func NewFriendService(db *sql.DB, ns *NotificationService) *FriendService {
	return &FriendService{db, ns}
}

type CreateFriendRequestParams struct {
	UserId     string
	ReceiverId string
}

func (fs *FriendService) CreateFriendRequest(ctx context.Context, params CreateFriendRequestParams) (*models.FriendRequest, error) {
	var outerRequest *models.FriendRequest
	err := utils.Tx(ctx, fs.db, func(tx *sql.Tx) error {
		pendingFriendRequests, err := models.FriendRequests(models.FriendRequestWhere.State.EQ(models.FriendRequestStatePending), models.FriendRequestWhere.SenderID.EQ(params.UserId)).Count(ctx, tx)
		if err != nil {
			return err
		}
		if pendingFriendRequests > 0 {
			return utils.PendingFriendRequestExistsError{}
		}

		existingFriendShip, err := models.Friendships(
			qm.Expr(models.FriendshipWhere.Friend1ID.EQ(params.UserId),
				qm.Or2(models.FriendshipWhere.Friend2ID.EQ(params.UserId)))).Count(ctx, tx)
		if err != nil {
			return err
		}
		if existingFriendShip > 0 {
			return utils.FriendShipExistsError{}
		}
		requestId, err := gonanoid.New()
		if err != nil {
			return err
		}
		request := models.FriendRequest{
			ID:         requestId,
			SenderID:   params.UserId,
			ReceiverID: params.ReceiverId,
			CreatedAt:  time.Now(),
		}
		err = request.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}
		outerRequest = &request
		return nil
	})
	if err != nil {
		return nil, err
	}
	receiver, err := operations.FindUserByID(ctx, fs.db, outerRequest.ReceiverID)
	if err != nil {
		return nil, err
	}
	sender, err := operations.FindUserByID(ctx, fs.db, outerRequest.SenderID)
	if err != nil {
		return nil, err
	}

	err = fs.ns.CreateFriendRequest(ctx,
		CreateFriendRequestNotificationParams{
			ReceiverId:    outerRequest.ReceiverID,
			SenderId:      outerRequest.SenderID,
			ReceiverName:  receiver.Name,
			ReceiverEmail: receiver.Email,
			SenderName:    sender.Name,
			RequestId:     outerRequest.ID,
		})
	if err != nil {
		log.Error().Msgf("Could not create notification: %v", err)
	}
	return fs.GetFriendRequest(ctx, outerRequest.ID)
}

func (fs *FriendService) GetFriendRequest(ctx context.Context, id string) (*models.FriendRequest, error) {
	friendRequest, err := models.FriendRequests(
		models.FriendRequestWhere.ID.EQ(id),
		qm.Load(models.FriendRequestRels.Sender),
		qm.Load(models.FriendRequestRels.Receiver),
	).One(ctx, fs.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "FriendRequest"}
		}
		return nil, err
	}
	return friendRequest, err
}

type CancelFriendRequestParams struct {
	UserId    string
	RequestId string
}

func (fs *FriendService) CancelFriendRequest(ctx context.Context, params CancelFriendRequestParams) (*models.FriendRequest, error) {
	request, err := models.FriendRequests(models.FriendRequestWhere.ID.EQ(params.RequestId)).One(ctx, fs.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "FriendRequest"}
		}
		return nil, err
	}
	if request.SenderID != params.UserId {
		return nil, utils.EntityDoesNotBelongToUserError{}
	}
	_, err = request.Delete(ctx, fs.db)
	if err != nil {
		return nil, err
	}
	return request, nil
}

type FriendRequestsResult struct {
	Received models.FriendRequestSlice
	Sent     models.FriendRequestSlice
}

func (fs *FriendService) GetFriendRequests(ctx context.Context, userId string) (*FriendRequestsResult, error) {
	received, err := models.FriendRequests(
		models.FriendRequestWhere.ReceiverID.EQ(userId),
		qm.Load(models.FriendRequestRels.Sender),
		qm.Load(models.FriendRequestRels.Receiver),
	).All(ctx, fs.db)
	if err != nil {
		return nil, err
	}
	sent, err := models.FriendRequests(
		models.FriendRequestWhere.SenderID.EQ(userId),
		qm.Load(models.FriendRequestRels.Sender),
		qm.Load(models.FriendRequestRels.Receiver),
	).All(ctx, fs.db)
	if err != nil {
		return nil, err
	}
	return &FriendRequestsResult{
		Received: received,
		Sent:     sent,
	}, nil
}

type ReactFriendRequestParams struct {
	FriendRequestId string
	UserId          string
	Accept          bool
}

func (fs *FriendService) ReactFriendRequest(ctx context.Context, params ReactFriendRequestParams) (*models.FriendRequest, error) {
	var outerRequest *models.FriendRequest
	err := utils.Tx(ctx, fs.db, func(tx *sql.Tx) error {
		request, err := models.FriendRequests(models.FriendRequestWhere.ID.EQ(params.FriendRequestId)).One(ctx, tx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return utils.NotFoundError{EntityName: "FriendRequest"}
			}
			return err
		}
		// only the receiver can accept or reject a friend request``
		if request.ReceiverID != params.UserId {
			return utils.EntityDoesNotBelongToUserError{}
		}
		// only requests that are pending can be accepted or rejected
		if request.State != models.FriendRequestStatePending {
			return utils.FriendRequestNotPendingError{}
		}

		if !params.Accept {
			request.State = models.FriendRequestStateRejected
		} else {
			request.State = models.FriendRequestStateAccepted
			// create friendship
			friendship := models.Friendship{
				Friend1ID:       request.ReceiverID,
				Friend2ID:       request.SenderID,
				FriendRequestID: request.ID,
				CreatedAt:       time.Now(),
			}
			err = friendship.Insert(ctx, tx, boil.Infer())
			if err != nil {
				return err
			}
		}
		_, err = request.Update(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}
		outerRequest = request
		return nil
	})
	if err != nil {
		return nil, err
	}
	receiver, err := operations.FindUserByID(ctx, fs.db, outerRequest.ReceiverID)
	if err != nil {
		return nil, err
	}
	sender, err := operations.FindUserByID(ctx, fs.db, outerRequest.SenderID)
	if err != nil {
		return nil, err
	}
	err = fs.ns.CreateFriendRequestReaction(ctx, CreateFriendRequestReactionParams{
		RequestId:    outerRequest.ID,
		ReceiverId:   outerRequest.ReceiverID,
		Accepted:     params.Accept,
		SenderEmail:  receiver.Email,
		ReceiverName: receiver.Name,
		SenderName:   sender.Name,
	})
	if err != nil {
		log.Error().Msgf("Could not create notification: %v", err)
	}
	return fs.GetFriendRequest(ctx, outerRequest.ID)
}

func (fs *FriendService) GetFriends(ctx context.Context, userId string) (models.FriendshipSlice, error) {
	return models.Friendships(
		qm.Expr(models.FriendshipWhere.Friend1ID.EQ(userId),
			qm.Or2(models.FriendshipWhere.Friend2ID.EQ(userId))),
		qm.Load(models.FriendshipRels.Friend1),
		qm.Load(models.FriendshipRels.Friend2),
		qm.Load(models.FriendshipRels.FriendRequest),
		qm.Load(qm.Rels(models.FriendshipRels.FriendRequest, models.FriendRequestRels.Sender)),
		qm.Load(qm.Rels(models.FriendshipRels.FriendRequest, models.FriendRequestRels.Receiver)),
	).All(ctx, fs.db)
}

func (fs *FriendService) Unfriend(ctx context.Context, userId string, friendId string) error {
	friendShip, err := models.Friendships(
		qm.Expr(models.FriendshipWhere.Friend1ID.EQ(userId),
			qm.Or2(models.FriendshipWhere.Friend2ID.EQ(userId))),
	).One(ctx, fs.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.NotFoundError{EntityName: "Friend"}
		}
	}
	_, err = friendShip.Delete(ctx, fs.db)
	return err
}
