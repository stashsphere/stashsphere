package services

import (
	"context"
	"database/sql"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ShareService struct {
	db *sql.DB
}

func NewShareService(db *sql.DB) *ShareService {
	return &ShareService{
		db,
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
			return utils.ErrEntityDoesNotBelongToUser
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
			return utils.ErrEntityDoesNotBelongToUser
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
		return nil, err
	}
	if share.OwnerID != requestingUser && share.TargetUserID != requestingUser {
		return nil, utils.ErrEntityDoesNotBelongToUser
	}
	return share, nil
}
