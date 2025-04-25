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

type ListService struct {
	db *sql.DB
}

func NewListService(db *sql.DB) *ListService {
	return &ListService{db}
}

type CreateListParams struct {
	Name     string
	ThingIds []string
	OwnerId  string
}

func (ls *ListService) CreateList(ctx context.Context, params CreateListParams) (*models.List, error) {
	var outerList *models.List
	err := utils.Tx(ctx, ls.db, func(tx *sql.Tx) error {
		listID, err := gonanoid.New()
		if err != nil {
			return err
		}

		list := models.List{
			ID:      listID,
			Name:    params.Name,
			OwnerID: params.OwnerId,
		}

		err = list.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}
		for _, thingId := range params.ThingIds {
			thing, err := models.Things(models.ThingWhere.ID.EQ(thingId)).One(ctx, tx)
			if err != nil {
				return err
			}
			if thing.OwnerID != params.OwnerId {
				return utils.ErrEntityDoesNotBelongToUser
			}
			err = thing.AddLists(ctx, tx, false, &list)
			if err != nil {
				return err
			}
		}

		outerList = &list
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ls.GetList(ctx, outerList.ID, outerList.OwnerID)
}

type UpdateListParams struct {
	Name     string
	ThingIds []string
}

func (ls *ListService) UpdateList(ctx context.Context, listId string, userId string, params UpdateListParams) (*models.List, error) {
	var outerList *models.List
	err := utils.Tx(ctx, ls.db, func(tx *sql.Tx) error {
		list, err := models.Lists(qm.Load(models.ListRels.Things), models.ListWhere.ID.EQ(listId)).One(ctx, tx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return utils.ErrNotFoundError{EntityName: "List"}
			}
			return err
		}
		if list.OwnerID != userId {
			return utils.ErrEntityDoesNotBelongToUser
		}

		list.Name = params.Name
		_, err = list.Update(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}

		err = list.RemoveThings(ctx, tx, list.R.Things...)
		if err != nil {
			return err
		}

		for _, thingId := range params.ThingIds {
			thing, err := models.Things(models.ThingWhere.ID.EQ(thingId)).One(ctx, tx)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return utils.ErrNotFoundError{EntityName: "Thing"}
				}
				return err
			}
			if thing.OwnerID != userId {
				return utils.ErrEntityDoesNotBelongToUser
			}
			err = thing.AddLists(ctx, tx, false, list)
			if err != nil {
				return err
			}
		}
		outerList = list
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ls.GetList(ctx, outerList.ID, outerList.OwnerID)
}

type GetListsForUserParams struct {
	UserId   string
	PerPage  uint64
	Page     uint64
	Paginate bool
}

func (ls *ListService) GetListsForUser(ctx context.Context, params GetListsForUserParams) (uint64, uint64, models.ListSlice, error) {
	userId, perPage, page, paginate := params.UserId, params.PerPage, params.Page, params.Paginate

	tx, err := ls.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return 0, 0, nil, err
	}

	type ListIdRow struct {
		ListId string `boil:"list_id"`
	}
	var sharedListIdRows []ListIdRow
	sharedListIds := []interface{}{}
	err = models.NewQuery(
		qm.Distinct("list_id"),
		qm.From("shares_lists"),
		qm.InnerJoin("shares on share_id = id"),
		qm.Where("target_user_id=?", userId),
	).Bind(ctx, tx, &sharedListIdRows)
	if err != nil {
		return 0, 0, nil, err
	}
	for _, listIdRow := range sharedListIdRows {
		sharedListIds = append(sharedListIds, listIdRow.ListId)
	}

	searchCond := qm.Expr(
		models.ListWhere.OwnerID.EQ(userId),
		qm.OrIn("id in ?", sharedListIds...),
	)

	listCount, err := models.Lists(searchCond).Count(ctx, tx)
	if err != nil {
		return 0, 0, nil, err
	}

	// empty expr for no pagination
	listQuery := []qm.QueryMod{}
	if paginate {
		listQuery = append(listQuery, qm.Offset(int(perPage*page)), qm.Limit(int(perPage)))
	}

	sortCond := qm.OrderBy(models.ThingColumns.CreatedAt)

	listQuery = append(listQuery,
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.Owner)),
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.ThingImages)),
		qm.Load(models.ListRels.Owner),
		searchCond,
		sortCond,
	)

	lists, err := models.Lists(listQuery...).All(ctx, tx)
	if err != nil {
		return 0, 0, nil, err
	}

	totalPages := uint64(math.Ceil(float64(listCount) / float64(perPage)))

	return uint64(listCount), totalPages, lists, nil
}

func (ls *ListService) GetListsWhereThingIsPartOf(ctx context.Context, thingId string) (models.ListSlice, error) {
	return models.Lists(qm.InnerJoin("lists_things on lists.id = list_things.list_id", qm.Where("list_things.thingId = ?", thingId))).All(ctx, ls.db)
}

func (ls *ListService) GetList(ctx context.Context, listId string, userId string) (*models.List, error) {
	list, err := operations.GetListUnchecked(ctx, ls.db, listId)
	if err != nil {
		return nil, err
	}
	sharedListsForUsers, err := operations.GetSharedListIdsForUser(ctx, ls.db, userId)
	if err != nil {
		return nil, err
	}
	authorized := func() bool {
		for _, id := range sharedListsForUsers {
			if id == listId {
				return true
			}
		}
		if userId == list.OwnerID {
			return true
		}
		return false
	}()
	if !authorized {
		return nil, utils.ErrUserHasNoAccessRights
	}
	return list, nil
}

func (ls *ListService) AddThingToList(ctx context.Context, thingId string, listId string, userId string) (*models.List, error) {
	thing, err := models.FindThing(ctx, ls.db, thingId)
	if err != nil {
		return nil, err
	}
	list, err := models.FindList(ctx, ls.db, listId)
	if err != nil {
		return nil, err
	}
	if thing.OwnerID != userId || list.OwnerID != userId {
		return nil, utils.ErrEntityDoesNotBelongToUser
	}
	err = list.AddThings(ctx, ls.db, true, thing)
	if err != nil {
		return nil, err
	}
	return ls.GetList(ctx, listId, list.OwnerID)
}

func (ls *ListService) RemoveThingFromList(ctx context.Context, thingId string, listId string, userId string) (*models.List, error) {
	thing, err := models.FindThing(ctx, ls.db, thingId)
	if err != nil {
		return nil, err
	}
	list, err := models.FindList(ctx, ls.db, listId)
	if err != nil {
		return nil, err
	}
	if thing.OwnerID != userId || list.OwnerID != userId {
		return nil, utils.ErrEntityDoesNotBelongToUser
	}
	err = list.RemoveThings(ctx, ls.db, thing)
	if err != nil {
		return nil, err
	}
	return ls.GetList(ctx, listId, userId)
}

func (ls *ListService) GetSharedListIdsForUser(ctx context.Context, userId string) ([]string, error) {
	return operations.GetSharedListIdsForUser(ctx, ls.db, userId)
}
