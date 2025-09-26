package resources

import (
	"time"

	"github.com/stashsphere/backend/models"
)

type CartEntry struct {
	CreatedAt time.Time `json:"createdAt"`
	ThingId   string    `json:"thingId"`
	OwnerId   string    `json:"ownerId"`
}

type Cart struct {
	Entries []CartEntry `json:"entries"`
}

func CartEntryFromModel(entry *models.CartEntry) CartEntry {
	return CartEntry{
		CreatedAt: entry.CreatedAt,
		ThingId:   entry.ThingID,
		OwnerId:   entry.R.Thing.OwnerID,
	}
}

func CartFromModelSlice(mCartEntries models.CartEntrySlice) Cart {
	entries := make([]CartEntry, len(mCartEntries))
	for i, entry := range mCartEntries {
		entries[i] = CartEntryFromModel(entry)
	}
	return Cart{
		Entries: entries,
	}
}
