package resources

import "github.com/stashsphere/backend/services"

type SearchResult struct {
	Things []Thing `json:"things"`
	Lists  []List  `json:"lists"`
}

func SearchResultsFromModel(result *services.SearchResult, userId string, sharedListIds []string) *SearchResult {
	thingReasonMap := make(map[string]*AccessReason)
	for thingId, reason := range result.ThingReasonMap {
		thingReasonMap[thingId] = AccessReasonFromOperations(reason)
	}

	listReasonMap := make(map[string]*AccessReason)
	for listId, reason := range result.ListReasonMap {
		listReasonMap[listId] = AccessReasonFromOperations(reason)
	}

	return &SearchResult{
		Things: ThingsFromModel(result.Things, userId, sharedListIds, thingReasonMap),
		Lists:  ListsFromModel(result.Lists, userId, sharedListIds, listReasonMap, thingReasonMap),
	}
}
