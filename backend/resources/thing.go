package resources

import (
	"sort"
	"time"

	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/utils"
)

type Thing struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	PrivateNote  *string        `json:"privateNote"`
	CreatedAt    time.Time      `json:"createdAt"`
	Owner        User           `json:"owner"`
	Lists        []ReducedList  `json:"lists"`
	Images       []ReducedImage `json:"images"`
	Properties   []interface{}  `json:"properties"`
	Shares       []ReducedShare `json:"shares"`
	SharingState *string        `json:"sharingState"`
	Actions      Actions        `json:"actions"`
	Quantity     int64          `json:"quantity"`
	QuantityUnit string         `json:"quantityUnit"`
}

func SumQuantityEntries(entries models.QuantityEntrySlice) int64 {
	sum := int64(0)
	for _, entry := range entries {
		sum += int64(entry.DeltaValue)
	}
	return sum
}

func ThingFromModel(thing *models.Thing, userId string, sharedListIds []string) *Thing {
	shares := []ReducedShare{}
	if thing.OwnerID == userId {
		shares = ReducedSharesFromModelSlice(thing.R.Shares)
	}
	filteredLists := []ReducedList{}
	lists := ReducedListsFromModelSlice(thing.R.Lists, userId)
	if thing.OwnerID == userId {
		filteredLists = lists
	} else {
		for _, list := range lists {
			if utils.Contains(sharedListIds, list.ID) {
				filteredLists = append(filteredLists, list)
			}
		}
	}
	canEdit := thing.OwnerID == userId
	canShare := thing.OwnerID == userId
	canDelete := thing.OwnerID == userId

	var privateNote *string
	var sharingState *string
	sharingStateString := thing.SharingState.String()
	if thing.OwnerID == userId {
		privateNote = &thing.PrivateNote
		sharingState = &sharingStateString
	}

	imageThings := make([]models.ImagesThing, len(thing.R.ImagesThings))
	for i, imageThing := range thing.R.ImagesThings {
		imageThings[i] = *imageThing
	}
	sort.Slice(imageThings, func(i, j int) bool {
		return imageThings[i].Pos < imageThings[j].Pos
	})
	images := make([]models.Image, len(imageThings))
	for i, imageThing := range imageThings {
		images[i] = *imageThing.R.Image
	}

	return &Thing{
		ID:           thing.ID,
		Name:         thing.Name,
		PrivateNote:  privateNote,
		Description:  thing.Description,
		CreatedAt:    thing.CreatedAt,
		Owner:        UserFromModel(thing.R.Owner),
		Lists:        filteredLists,
		Images:       ReducedImagesFromModel(images),
		Properties:   PropertiesFromModelSlice(thing.R.Properties),
		Shares:       shares,
		SharingState: sharingState,
		Actions: Actions{
			CanEdit:   canEdit,
			CanDelete: canDelete,
			CanShare:  canShare,
		},
		Quantity:     SumQuantityEntries(thing.R.QuantityEntries),
		QuantityUnit: thing.QuantityUnit,
	}
}

func ThingsFromModel(mThings []models.Thing, userId string, sharedListIds []string) []Thing {
	things := make([]Thing, len(mThings))
	for i, thing := range mThings {
		things[i] = *ThingFromModel(&thing, userId, sharedListIds)
	}
	return things
}

func ThingsFromModelSlice(mThings models.ThingSlice, userId string, sharedListIds []string) []Thing {
	things := make([]Thing, len(mThings))
	for i, thing := range mThings {
		things[i] = *ThingFromModel(thing, userId, sharedListIds)
	}
	return things
}

type ReducedThing struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PrivateNote *string   `json:"privateNote"`
	CreatedAt   time.Time `json:"createdAt"`
	Owner       User      `json:"owner"`
	Quantity    int64     `json:"quantity"`
}

func ReducedThingFromModel(thing *models.Thing, userId string) *ReducedThing {
	var privateNote *string
	if thing.OwnerID == userId {
		privateNote = &thing.PrivateNote
	}
	return &ReducedThing{
		ID:          thing.ID,
		Name:        thing.Name,
		PrivateNote: privateNote,
		Description: thing.Description,
		CreatedAt:   thing.CreatedAt,
		Owner:       UserFromModel(thing.R.Owner),
		Quantity:    SumQuantityEntries(thing.R.QuantityEntries),
	}
}

func ReducedThingsFromModelSlice(things models.ThingSlice, userId string) []ReducedThing {
	reducedThings := make([]ReducedThing, len(things))
	for i, thing := range things {
		reducedThings[i] = *ReducedThingFromModel(thing, userId)
	}
	return reducedThings
}

func ReducedThingsFromModel(things []models.Thing, userId string) []ReducedThing {
	reducedThings := make([]ReducedThing, len(things))
	for i, thing := range things {
		reducedThings[i] = *ReducedThingFromModel(&thing, userId)
	}
	return reducedThings
}

type PaginatedThings struct {
	Things         []Thing `json:"things"`
	PerPage        uint64  `json:"perPage"`
	Page           uint64  `json:"page"`
	TotalPageCount uint64  `json:"totalPageCount"`
	TotalCount     uint64  `json:"totalCount"`
}
