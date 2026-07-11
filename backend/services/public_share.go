package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
)

type PublicShareService struct {
	db *sql.DB
}

func NewPublicShareService(db *sql.DB) *PublicShareService {
	return &PublicShareService{
		db,
	}
}

type CreatePublicShareParams struct {
	ObjectId string
	OwnerId  string
}

func (ps *PublicShareService) CreatePublicShare(ctx context.Context, params CreatePublicShareParams) (*models.PublicShare, error) {
	var outerShare *models.PublicShare
	err := utils.Tx(ctx, ps.db, func(tx *sql.Tx) error {
		share := &models.PublicShare{
			OwnerID: params.OwnerId,
		}
		thing, err := operations.GetThingUnchecked(ctx, tx, params.ObjectId)
		if err == nil {
			// only the owner of the thing can share it
			if thing.OwnerID != params.OwnerId {
				return utils.EntityDoesNotBelongToUserError{}
			}
			share.ThingID = null.StringFrom(thing.ID)
		} else {
			list, err := operations.GetListUnchecked(ctx, tx, params.ObjectId)
			if err != nil {
				return err
			}
			// only the owner of the list can share it
			if list.OwnerID != params.OwnerId {
				return utils.EntityDoesNotBelongToUserError{}
			}
			share.ListID = null.StringFrom(list.ID)
		}
		shareId, err := gonanoid.New()
		if err != nil {
			return err
		}
		share.ID = shareId
		err = share.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}
		outerShare = share
		return nil
	})
	if err != nil {
		return nil, err
	}
	return outerShare, nil
}

func (ps *PublicShareService) GetPublicShare(ctx context.Context, token string) (*models.PublicShare, error) {
	share, err := models.PublicShares(
		models.PublicShareWhere.ID.EQ(token),
		qm.Load(qm.Rels(models.PublicShareRels.Thing, models.ThingRels.Owner)),
		qm.Load(qm.Rels(models.PublicShareRels.Thing, models.ThingRels.Properties)),
		qm.Load(qm.Rels(models.PublicShareRels.Thing, models.ThingRels.QuantityEntries)),
		qm.Load(qm.Rels(models.PublicShareRels.Thing, models.ThingRels.ImagesThings, models.ImagesThingRels.Image)),
		qm.Load(qm.Rels(models.PublicShareRels.List, models.ListRels.Owner)),
		qm.Load(qm.Rels(models.PublicShareRels.List, models.ListRels.Things, models.ThingRels.Owner)),
		qm.Load(qm.Rels(models.PublicShareRels.List, models.ListRels.Things, models.ThingRels.Properties)),
		qm.Load(qm.Rels(models.PublicShareRels.List, models.ListRels.Things, models.ThingRels.QuantityEntries)),
		qm.Load(qm.Rels(models.PublicShareRels.List, models.ListRels.Things, models.ThingRels.ImagesThings, models.ImagesThingRels.Image)),
	).One(ctx, ps.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "PublicShare"}
		}
		return nil, err
	}
	return share, nil
}

func (ps *PublicShareService) GetPublicSharesForUser(ctx context.Context, userId string) (models.PublicShareSlice, error) {
	return models.PublicShares(
		models.PublicShareWhere.OwnerID.EQ(userId),
		qm.Load(models.PublicShareRels.Thing),
		qm.Load(models.PublicShareRels.List),
		qm.OrderBy("created_at DESC"),
	).All(ctx, ps.db)
}

func (ps *PublicShareService) DeletePublicShare(ctx context.Context, token string, requestingUser string) error {
	return utils.Tx(ctx, ps.db, func(tx *sql.Tx) error {
		share, err := models.PublicShares(
			models.PublicShareWhere.ID.EQ(token),
		).One(ctx, tx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return utils.NotFoundError{EntityName: "PublicShare"}
			}
			return err
		}
		if share.OwnerID != requestingUser {
			return utils.EntityDoesNotBelongToUserError{}
		}
		_, err = share.Delete(ctx, tx)
		return err
	})
}
