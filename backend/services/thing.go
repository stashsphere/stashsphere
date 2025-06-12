package services

import (
	"context"
	"database/sql"
	"errors"
	"math"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ThingService struct {
	db           *sql.DB
	imageService *ImageService
}

func NewThingService(db *sql.DB, imageService *ImageService) *ThingService {
	return &ThingService{db, imageService}
}

type CreateThingParams struct {
	Name         string
	Description  string
	PrivateNote  string
	OwnerId      string
	Properties   []operations.CreatePropertyParams
	ImagesIds    []string
	Quantity     uint64
	QuantityUnit string
	SharingState string
}

func (ts *ThingService) CreateThing(ctx context.Context, params CreateThingParams) (*models.Thing, error) {
	var outerThing *models.Thing
	err := utils.Tx(ctx, ts.db, func(tx *sql.Tx) error {
		for _, imageId := range params.ImagesIds {
			res, err := operations.ImageBelongsToUser(ctx, tx, params.OwnerId, imageId)
			if err != nil {
				return err
			}
			if !res {
				return utils.EntityDoesNotBelongToUserError{}
			}
		}

		thingID, err := gonanoid.New()
		if err != nil {
			return err
		}

		sharingState := models.SharingStatePrivate
		switch params.SharingState {
		case "friends":
			sharingState = models.SharingStateFriends
		case "friends-of-friends":
			sharingState = models.SharingStateFriendsOfFriends
		}

		thing := &models.Thing{
			ID:           thingID,
			Name:         params.Name,
			Description:  params.Description,
			PrivateNote:  params.PrivateNote,
			OwnerID:      params.OwnerId,
			QuantityUnit: params.QuantityUnit,
			SharingState: sharingState,
		}

		err = thing.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}

		for _, prop := range params.Properties {
			_, err = operations.CreateProperty(ctx, tx, thingID, prop)
			if err != nil {
				return err
			}
		}

		quantityID, err := gonanoid.New()
		if err != nil {
			return err
		}
		err = thing.AddQuantityEntries(ctx, tx, true, &models.QuantityEntry{DeltaValue: int64(params.Quantity), ID: quantityID})
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}

		images, err := models.Images(models.ImageWhere.ID.IN(params.ImagesIds)).All(ctx, tx)
		if err != nil {
			return err
		}
		// TODO make sure the image belongs to the owner of the thing
		err = thing.AddImages(ctx, tx, false, images...)
		if err != nil {
			return err
		}
		outerThing = thing
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ts.GetThing(ctx, outerThing.ID, outerThing.OwnerID)
}

func (ts *ThingService) GetThing(ctx context.Context, thingId string, userId string) (*models.Thing, error) {
	thing, err := operations.GetThingUnchecked(ctx, ts.db, thingId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "Thing"}
		}
		return nil, err
	}
	sharedThingsForUser, err := operations.GetSharedThingIdsForUser(ctx, ts.db, userId)
	if err != nil {
		return nil, err
	}
	authorized := func() bool {
		for _, id := range sharedThingsForUser {
			if id == thingId {
				return true
			}
		}
		if userId == thing.OwnerID {
			return true
		}
		return false
	}()
	if !authorized {
		return nil, utils.UserHasNoAccessRightsError{}
	}

	return thing, nil
}

type UpdateThingParams struct {
	Name         string
	Description  string
	PrivateNote  string
	Properties   []operations.CreatePropertyParams
	ImagesIds    []string
	Quantity     uint64
	QuantityUnit string
	SharingState string
}

func (ts *ThingService) EditThing(ctx context.Context, thingId string, userId string, params UpdateThingParams) (*models.Thing, error) {
	var outerThing *models.Thing
	err := utils.Tx(ctx, ts.db, func(tx *sql.Tx) error {
		thing, err := models.Things(
			qm.Load(models.ThingRels.Properties),
			qm.Load(models.ThingRels.Images),
			qm.Load(models.ThingRels.QuantityEntries),
			models.ThingWhere.ID.EQ(thingId),
		).One(ctx, tx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return utils.NotFoundError{EntityName: "Thing"}
			}
			return err
		}
		if thing.OwnerID != userId {
			return utils.EntityDoesNotBelongToUserError{}
		}

		sharingState := models.SharingStatePrivate
		switch params.SharingState {
		case "friends":
			sharingState = models.SharingStateFriends
		case "friends-of-friends":
			sharingState = models.SharingStateFriendsOfFriends
		}

		thing.PrivateNote = params.PrivateNote
		thing.Name = params.Name
		thing.Description = params.Description
		thing.QuantityUnit = params.QuantityUnit
		thing.SharingState = sharingState

		_, err = thing.Update(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}

		_, err = thing.R.Properties.DeleteAll(ctx, tx)
		if err != nil {
			return err
		}

		for _, prop := range params.Properties {
			_, err = operations.CreateProperty(ctx, tx, thingId, prop)
			if err != nil {
				return err
			}
		}

		for _, imageId := range params.ImagesIds {
			res, err := operations.ImageBelongsToUser(ctx, tx, userId, imageId)
			if err != nil {
				return err
			}
			if !res {
				return utils.EntityDoesNotBelongToUserError{}
			}
		}

		quantityID, err := gonanoid.New()
		if err != nil {
			return err
		}
		err = thing.AddQuantityEntries(ctx, tx, true, &models.QuantityEntry{
			DeltaValue: operations.DeltaQuantity(thing, params.Quantity),
			ID:         quantityID,
		})
		if err != nil {
			return err
		}

		images, err := models.Images(models.ImageWhere.ID.IN(params.ImagesIds)).All(ctx, tx)
		if err != nil {
			return err
		}

		err = thing.SetImages(ctx, tx, false, images...)
		if err != nil {
			return err
		}
		outerThing = thing
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ts.GetThing(ctx, thingId, outerThing.OwnerID)
}

type GetThingsForUserParams struct {
	UserId   string
	PerPage  uint64
	Page     uint64
	Paginate bool
}

func (ts *ThingService) GetThingsForUser(ctx context.Context, params GetThingsForUserParams) (uint64, uint64, models.ThingSlice, error) {
	userId, perPage, page, paginate := params.UserId, params.PerPage, params.Page, params.Paginate

	tx, err := ts.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return 0, 0, nil, err
	}

	sharedThingIds, err := operations.GetSharedThingIdsForUser(ctx, tx, userId)
	if err != nil {
		return 0, 0, nil, err
	}
	interfaceIds := make([]interface{}, len(sharedThingIds))
	for i, s := range sharedThingIds {
		interfaceIds[i] = s
	}

	searchCond := qm.Expr(
		models.ThingWhere.OwnerID.EQ(userId),
		qm.OrIn("id in ?", interfaceIds...),
	)

	thingCount, err := models.Things(searchCond).Count(ctx, tx)
	if err != nil {
		return 0, 0, models.ThingSlice{}, err
	}

	// empty expr for no pagination
	thingQuery := []qm.QueryMod{}
	if paginate {
		thingQuery = append(thingQuery, qm.Offset(int(perPage*page)), qm.Limit(int(perPage)))
	}

	sortCond := qm.OrderBy(models.ThingColumns.CreatedAt)

	thingQuery = append(thingQuery,
		qm.Load(models.ThingRels.Properties),
		qm.Load(models.ThingRels.QuantityEntries),
		qm.Load(qm.Rels(models.ThingRels.Lists, models.ListRels.Owner)),
		qm.Load(models.ThingRels.Owner),
		qm.Load(qm.Rels(models.ThingRels.Shares, models.ShareRels.Owner)),
		qm.Load(qm.Rels(models.ThingRels.Shares, models.ShareRels.TargetUser)),
		qm.Load(models.ThingRels.Images),
		searchCond,
		sortCond,
	)

	things, err := models.Things(thingQuery...).All(ctx, tx)
	if err != nil {
		return 0, 0, models.ThingSlice{}, err
	}
	totalPages := uint64(math.Ceil(float64(thingCount) / float64(perPage)))
	return uint64(thingCount), totalPages, things, nil
}
