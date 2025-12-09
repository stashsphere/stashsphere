package services

import (
	"context"
	"database/sql"
	"errors"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ShareService struct {
	db *sql.DB
	ns *NotificationService
}

func NewShareService(db *sql.DB, ns *NotificationService) *ShareService {
	return &ShareService{
		db,
		ns,
	}
}

type CreateThingShareParams struct {
	ThingId      string
	OwnerId      string
	TargetUserId string
}

func (ss *ShareService) CreateThingShare(ctx context.Context, params CreateThingShareParams) (*models.Share, error) {
	var outerShare *models.Share
	err := utils.Tx(ctx, ss.db, func(tx *sql.Tx) error {
		// check whether the thing exists and belongs to the owner
		thing, err := operations.GetThingUnchecked(ctx, tx, params.ThingId)
		if err != nil {
			return err
		}
		// only the owner of the thing can share it
		if thing.OwnerID != params.OwnerId {
			return utils.EntityDoesNotBelongToUserError{}
		}
		shareId, err := gonanoid.New()
		if err != nil {
			return err
		}
		share := &models.Share{
			ID:           shareId,
			TargetUserID: params.TargetUserId,
			OwnerID:      params.OwnerId,
		}
		err = share.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}
		err = share.AddThings(ctx, tx, false, thing)
		if err != nil {
			return err
		}
		outerShare = share
		return nil
	})
	if err != nil {
		return nil, err
	}

	sharer, err := operations.FindUserByID(ctx, ss.db, params.OwnerId)
	if err != nil {
		return nil, err
	}

	targetUser, err := operations.FindUserByID(ctx, ss.db, params.TargetUserId)
	if err != nil {
		return nil, err
	}

	err = ss.ns.ThingShared(ctx, ThingSharedParams{
		ThingId:         params.ThingId,
		SharerName:      sharer.Name,
		SharedId:        sharer.ID,
		TargetUserId:    params.TargetUserId,
		TargetUserName:  targetUser.Name,
		TargetUserEmail: targetUser.Email,
	})
	if err != nil {
		log.Error().Msgf("Could not create notification: %v", err)
	}
	return ss.GetShare(ctx, outerShare.ID, outerShare.TargetUserID)
}

type CreateListShareParams struct {
	ListId       string
	OwnerId      string
	TargetUserId string
}

func (ss *ShareService) CreateListShare(ctx context.Context, params CreateListShareParams) (*models.Share, error) {
	var outerShare *models.Share
	err := utils.Tx(ctx, ss.db, func(tx *sql.Tx) error {
		// check whether the list exists and belongs to the owner
		list, err := operations.GetListUnchecked(ctx, tx, params.ListId)
		if err != nil {
			return err
		}
		// only the owner of the list can share it
		if list.OwnerID != params.OwnerId {
			return utils.EntityDoesNotBelongToUserError{}
		}
		shareId, err := gonanoid.New()
		if err != nil {
			return err
		}
		share := &models.Share{
			ID:           shareId,
			TargetUserID: params.TargetUserId,
			OwnerID:      params.OwnerId,
		}
		err = share.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}
		err = share.AddLists(ctx, tx, false, list)
		if err != nil {
			return err
		}
		outerShare = share
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = ss.ns.ListShared(ctx, ListSharedParams{
		ListId:       params.ListId,
		SharedId:     params.OwnerId,
		TargetUserId: params.TargetUserId,
	})

	if err != nil {
		log.Error().Msgf("Could not create notification: %v", err)
	}
	return ss.GetShare(ctx, outerShare.ID, outerShare.TargetUserID)
}

type CreateShareParams struct {
	ObjectId     string
	TargetUserId string
	OwnerId      string
}

func (ss *ShareService) CreateShare(ctx context.Context, params CreateShareParams) (*models.Share, error) {
	_, err := operations.GetThingUnchecked(ctx, ss.db, params.ObjectId)
	if err == nil {
		return ss.CreateThingShare(ctx, CreateThingShareParams{
			ThingId:      params.ObjectId,
			OwnerId:      params.OwnerId,
			TargetUserId: params.TargetUserId,
		})
	} else {
		return ss.CreateListShare(ctx, CreateListShareParams{
			ListId:       params.ObjectId,
			OwnerId:      params.OwnerId,
			TargetUserId: params.TargetUserId,
		})
	}
}

func (ss *ShareService) GetShare(ctx context.Context, shareId string, requestingUser string) (*models.Share, error) {
	share, err := models.Shares(
		qm.Load(qm.Rels(models.ShareRels.Things, models.ThingRels.Owner)),
		qm.Load(qm.Rels(models.ShareRels.Lists, models.ListRels.Owner)),
		qm.Load(models.ShareRels.TargetUser),
		qm.Load(models.ShareRels.Owner),
		models.ShareWhere.ID.EQ(shareId),
	).One(ctx, ss.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "Share"}
		}
		return nil, err
	}
	if share.OwnerID != requestingUser && share.TargetUserID != requestingUser {
		return nil, utils.EntityDoesNotBelongToUserError{}
	}
	return share, nil
}

func (ss *ShareService) DeleteShare(ctx context.Context, shareId string, requestingUser string) error {
	err := utils.Tx(ctx, ss.db, func(tx *sql.Tx) error {
		share, err := models.Shares(
			models.ShareWhere.ID.EQ(shareId),
			qm.Load(models.ShareRels.Things),
			qm.Load(models.ShareRels.Lists),
			qm.Load(qm.Rels(models.ShareRels.Lists, models.ListRels.Things)),
		).One(ctx, tx)
		if err != nil {
			return err
		}
		if share.OwnerID != requestingUser {
			return utils.EntityDoesNotBelongToUserError{}
		}
		// all shared things that might no longer be accessible by users
		thingIds := []string{}
		for _, thing := range share.R.Things {
			thingIds = append(thingIds, thing.ID)
		}
		for _, list := range share.R.Lists {
			for _, thing := range list.R.Things {
				thingIds = append(thingIds, thing.ID)
			}
		}
		err = share.RemoveThings(ctx, tx, share.R.Things...)
		if err != nil {
			return err
		}
		err = share.RemoveLists(ctx, tx, share.R.Lists...)
		if err != nil {
			return err
		}
		_, err = share.Delete(ctx, tx)
		if err != nil {
			return err
		}
		err = operations.RemoveForbiddenThingsFromCarts(ctx, tx, thingIds)
		return err
	})
	return err
}
