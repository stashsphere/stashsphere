package services

import (
	"context"
	"database/sql"
	"strings"

	"github.com/stashsphere/backend/models"
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
	Things []models.Thing `json:"things"`
	Lists  []models.List  `json:"lists"`
}

// Basic search using substring matchin
// TODO add pagination
func (sp *SearchService) Search(ctx context.Context, userId string, params *SearchParams) (*SearchResult, error) {
	_, _, userThings, err := sp.thingService.GetThingsForUser(ctx, GetThingsForUserParams{
		UserId:   userId,
		Paginate: false,
	})
	if err != nil {
		return nil, err
	}
	_, _, userLists, err := sp.listService.GetListsForUser(ctx, GetListsForUserParams{
		UserId:   userId,
		Paginate: false,
	})
	if err != nil {
		return nil, err
	}
	filteredThings := []models.Thing{}
	for _, thing := range userThings {
		if strings.Contains(strings.ToLower(thing.Name), strings.ToLower(params.Query)) ||
			strings.Contains(strings.ToLower(thing.Description), strings.ToLower(params.Query)) ||
			(userId == thing.OwnerID && strings.Contains(strings.ToLower(thing.PrivateNote), strings.ToLower(params.Query))) {
			filteredThings = append(filteredThings, *thing)
		}
	}
	filteredLists := []models.List{}
	for _, list := range userLists {
		if strings.Contains(strings.ToLower(list.Name), strings.ToLower(params.Query)) {
			filteredLists = append(filteredLists, *list)
		}
	}
	result := &SearchResult{
		Things: filteredThings,
		Lists:  filteredLists,
	}
	return result, nil
}
