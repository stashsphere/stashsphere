package resources

import (
	"time"

	"github.com/stashsphere/backend/models"
)

type List struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	CreatedAt    time.Time      `json:"createdAt"`
	Owner        User           `json:"owner"`
	Things       []Thing        `json:"things"`
	Actions      Actions        `json:"actions"`
	Shares       []ReducedShare `json:"shares"`
	SharingState *string        `json:"sharingState"`
	Reason       *AccessReason  `json:"reason,omitempty"`
}

// requires an eager loaded list with things
func ListFromModel(list *models.List, userId string, sharedListIds []string, listReason *AccessReason, thingReasonMap map[string]*AccessReason) List {
	shares := []ReducedShare{}
	if list.OwnerID == userId {
		shares = ReducedSharesFromModelSlice(list.R.Shares)
	}
	thingResources := []Thing{}
	for _, e := range list.R.Things {
		thingResources = append(thingResources, *ThingFromModel(e, userId, sharedListIds, thingReasonMap))
	}
	canEdit := list.OwnerID == userId
	canShare := list.OwnerID == userId
	canDelete := list.OwnerID == userId

	sharingStateString := list.SharingState.String()
	var sharingState *string
	if list.OwnerID == userId {
		sharingState = &sharingStateString
	}

	return List{
		ID:           list.ID,
		Name:         list.Name,
		CreatedAt:    list.CreatedAt,
		Owner:        UserFromModel(list.R.Owner),
		Things:       thingResources,
		Shares:       shares,
		SharingState: sharingState,
		Actions: Actions{
			CanEdit:   canEdit,
			CanDelete: canDelete,
			CanShare:  canShare,
		},
		Reason: listReason,
	}
}

func ListsFromModelSlice(mLists models.ListSlice, userId string, sharedListIds []string, listReasonMap map[string]*AccessReason, thingReasonMap map[string]*AccessReason) []List {
	lists := make([]List, len(mLists))
	for i, list := range mLists {
		var listReason *AccessReason
		if listReasonMap != nil {
			listReason = listReasonMap[list.ID]
		}
		lists[i] = ListFromModel(list, userId, sharedListIds, listReason, thingReasonMap)
	}
	return lists
}

func ListsFromModel(mLists []models.List, userId string, sharedListIds []string, listReasonMap map[string]*AccessReason, thingReasonMap map[string]*AccessReason) []List {
	lists := make([]List, len(mLists))
	for i, list := range mLists {
		var listReason *AccessReason
		if listReasonMap != nil {
			listReason = listReasonMap[list.ID]
		}
		lists[i] = ListFromModel(&list, userId, sharedListIds, listReason, thingReasonMap)
	}
	return lists

}

type ReducedList struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	Owner     User      `json:"owner"`
	Actions   Actions   `json:"actions"`
}

func ReducedListFromModel(list *models.List, userId string) *ReducedList {
	canEdit := list.OwnerID == userId
	canShare := list.OwnerID == userId
	canDelete := list.OwnerID == userId
	return &ReducedList{
		ID:        list.ID,
		Name:      list.Name,
		CreatedAt: list.CreatedAt,
		Owner:     UserFromModel(list.R.Owner),
		Actions: Actions{
			CanEdit:   canEdit,
			CanDelete: canDelete,
			CanShare:  canShare,
		},
	}
}

func ReducedListsFromModelSlice(lists models.ListSlice, userId string) []ReducedList {
	reducedLists := make([]ReducedList, len(lists))
	for i, list := range lists {
		reducedLists[i] = *ReducedListFromModel(list, userId)
	}
	return reducedLists
}

type PaginatedLists struct {
	Things         []List `json:"lists"`
	PerPage        uint64 `json:"perPage"`
	Page           uint64 `json:"page"`
	TotalPageCount uint64 `json:"totalPageCount"`
	TotalCount     uint64 `json:"totalCount"`
}
