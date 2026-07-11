package resources

import (
	"sort"
	"time"

	"github.com/stashsphere/backend/models"
)

// PublicThing is the anonymous view of a thing. It deliberately excludes
// the private note, shares, sharing state, actions and list memberships.
// Only the owner name is exposed, not the owner id.
type PublicThing struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	CreatedAt    time.Time      `json:"createdAt"`
	OwnerName    string         `json:"ownerName"`
	Images       []ReducedImage `json:"images"`
	Properties   []interface{}  `json:"properties"`
	Quantity     int64          `json:"quantity"`
	QuantityUnit string         `json:"quantityUnit"`
}

func PublicThingFromModel(thing *models.Thing) PublicThing {
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

	return PublicThing{
		ID:           thing.ID,
		Name:         thing.Name,
		Description:  thing.Description,
		CreatedAt:    thing.CreatedAt,
		OwnerName:    thing.R.Owner.Name,
		Images:       ReducedImagesFromModel(images),
		Properties:   PropertiesFromModelSlice(thing.R.Properties),
		Quantity:     SumQuantityEntries(thing.R.QuantityEntries),
		QuantityUnit: thing.QuantityUnit,
	}
}

type PublicList struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	CreatedAt time.Time     `json:"createdAt"`
	OwnerName string        `json:"ownerName"`
	Things    []PublicThing `json:"things"`
}

func PublicListFromModel(list *models.List) PublicList {
	things := make([]PublicThing, len(list.R.Things))
	for i, thing := range list.R.Things {
		things[i] = PublicThingFromModel(thing)
	}
	return PublicList{
		ID:        list.ID,
		Name:      list.Name,
		CreatedAt: list.CreatedAt,
		OwnerName: list.R.Owner.Name,
		Things:    things,
	}
}

// PublicShare is the anonymous response for a public share token.
type PublicShare struct {
	Token string       `json:"token"`
	Type  string       `json:"type"` // "thing" or "list"
	Thing *PublicThing `json:"thing,omitempty"`
	List  *PublicList  `json:"list,omitempty"`
}

func PublicShareFromModel(share *models.PublicShare) PublicShare {
	if share.ThingID.Valid {
		thing := PublicThingFromModel(share.R.Thing)
		return PublicShare{
			Token: share.ID,
			Type:  "thing",
			Thing: &thing,
		}
	}
	list := PublicListFromModel(share.R.List)
	return PublicShare{
		Token: share.ID,
		Type:  "list",
		List:  &list,
	}
}

// PublicShareInfo is the owner facing representation of a public share
// as embedded in Thing and List resources.
type PublicShareInfo struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

func PublicShareInfoFromModel(share *models.PublicShare) PublicShareInfo {
	return PublicShareInfo{
		ID:        share.ID,
		CreatedAt: share.CreatedAt,
	}
}

func PublicShareInfosFromModelSlice(shares models.PublicShareSlice) []PublicShareInfo {
	infos := make([]PublicShareInfo, len(shares))
	for i, share := range shares {
		infos[i] = PublicShareInfoFromModel(share)
	}
	return infos
}

// PublicShareIndexEntry is the owner facing representation of a public share
// in the index of all public shares of a user.
type PublicShareIndexEntry struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	Type       string    `json:"type"` // "thing" or "list"
	ObjectId   string    `json:"objectId"`
	ObjectName string    `json:"objectName"`
}

func PublicShareIndexEntryFromModel(share *models.PublicShare) PublicShareIndexEntry {
	entry := PublicShareIndexEntry{
		ID:        share.ID,
		CreatedAt: share.CreatedAt,
	}
	if share.ThingID.Valid {
		entry.Type = "thing"
		entry.ObjectId = share.R.Thing.ID
		entry.ObjectName = share.R.Thing.Name
	} else {
		entry.Type = "list"
		entry.ObjectId = share.R.List.ID
		entry.ObjectName = share.R.List.Name
	}
	return entry
}

func PublicShareIndexEntriesFromModelSlice(shares models.PublicShareSlice) []PublicShareIndexEntry {
	entries := make([]PublicShareIndexEntry, len(shares))
	for i, share := range shares {
		entries[i] = PublicShareIndexEntryFromModel(share)
	}
	return entries
}
