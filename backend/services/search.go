package services

import (
	"context"
	"database/sql"
	"strings"

	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
)

type SearchService struct {
	db           *sql.DB
	thingService *ThingService
	listService  *ListService
}

func NewSearchService(db *sql.DB, thingService *ThingService, listService *ListService) *SearchService {
	return &SearchService{
		db,
		thingService,
		listService,
	}
}

type SearchParams struct {
	Query string `json:"query"`
}

type SearchResult struct {
	Things         []models.Thing
	Lists          []models.List
	ThingReasonMap map[string]operations.AccessReasonInformation
	ListReasonMap  map[string]operations.AccessReasonInformation
}

// Basic search using substring matchin
// TODO add pagination
func (sp *SearchService) Search(ctx context.Context, userId string, params *SearchParams) (*SearchResult, error) {
	thingsResult, err := sp.thingService.GetThingsForUser(ctx, GetThingsForUserParams{
		UserId:   userId,
		Paginate: false,
	})
	if err != nil {
		return nil, err
	}
	listsResult, err := sp.listService.GetListsForUser(ctx, GetListsForUserParams{
		UserId:   userId,
		Paginate: false,
	})
	if err != nil {
		return nil, err
	}
	filteredThings := []models.Thing{}
	thingReasonMap := make(map[string]operations.AccessReasonInformation)
	for _, thing := range thingsResult.Things {
		if strings.Contains(strings.ToLower(thing.Name), strings.ToLower(params.Query)) ||
			strings.Contains(strings.ToLower(thing.Description), strings.ToLower(params.Query)) ||
			(userId == thing.OwnerID && strings.Contains(strings.ToLower(thing.PrivateNote), strings.ToLower(params.Query))) {
			filteredThings = append(filteredThings, *thing)
			if reason, ok := thingsResult.ThingReasonMap[thing.ID]; ok {
				thingReasonMap[thing.ID] = reason
			}
		}
	}
	filteredLists := []models.List{}
	listReasonMap := make(map[string]operations.AccessReasonInformation)
	for _, list := range listsResult.Lists {
		if strings.Contains(strings.ToLower(list.Name), strings.ToLower(params.Query)) {
			filteredLists = append(filteredLists, *list)
			if reason, ok := listsResult.ListReasonMap[list.ID]; ok {
				listReasonMap[list.ID] = reason
			}
		}
	}
	result := &SearchResult{
		Things:         filteredThings,
		Lists:          filteredLists,
		ThingReasonMap: thingReasonMap,
		ListReasonMap:  listReasonMap,
	}
	return result, nil
}
