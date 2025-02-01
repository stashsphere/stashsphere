package resources

import "github.com/stashsphere/backend/services"

type SearchResult struct {
	Things []Thing `json:"things"`
	Lists  []List  `json:"lists"`
}

func SearchResultsFromModel(result *services.SearchResult, userId string, sharedListIds []string) *SearchResult {
	return &SearchResult{
		Things: ThingsFromModel(result.Things, userId, sharedListIds),
		Lists:  ListsFromModel(result.Lists, userId, sharedListIds),
	}
}
