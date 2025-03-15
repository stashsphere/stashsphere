package resources

import (
	"encoding/json"

	"github.com/stashsphere/backend/models"
)

type ShareType int

const (
	ThingShare ShareType = iota
	ListShare
)

func (st ShareType) String() string {
	switch st {
	case ThingShare:
		return "thing"
	case ListShare:
		return "list"
	default:
		return ""
	}
}

type Share struct {
	Type       ShareType   `json:"type"`
	TargetUser User        `json:"target_user"`
	Owner      User        `json:"owner"`
	Object     interface{} `json:"share"`
}

type ReducedShare struct {
	TargetUser User   `json:"target_user"`
	Owner      User   `json:"owner"`
	Id         string `json:"id"`
}

func (s *Share) MarshalJSON() ([]byte, error) {
	switch s.Type {
	case ThingShare:
		return json.Marshal(&struct {
			Type   string       `json:"type"`
			Object ReducedThing `json:"object"`
		}{
			Type:   s.Type.String(),
			Object: s.Object.(ReducedThing),
		})
	case ListShare:
		return json.Marshal(&struct {
			Type   string      `json:"type"`
			Object ReducedList `json:"object"`
		}{
			Type:   s.Type.String(),
			Object: s.Object.(ReducedList),
		})
	default:
		return nil, nil
	}
}

func ShareFromModel(share *models.Share, userId string) *Share {
	if len(share.R.Lists) > 0 {
		return &Share{
			Type:       ListShare,
			TargetUser: UserFromModel(share.R.TargetUser),
			Owner:      UserFromModel(share.R.Owner),
			Object:     *ReducedListFromModel(share.R.Lists[0], userId),
		}
	} else {
		return &Share{Type: ThingShare, Object: *ReducedThingFromModel(share.R.Things[0], userId)}
	}
}

func ReducedShareFromModel(s *models.Share) ReducedShare {
	return ReducedShare{TargetUser: UserFromModel(s.R.TargetUser), Owner: UserFromModel(s.R.Owner), Id: s.ID}
}

func SharesFromModelSlice(mShares models.ShareSlice, userId string) []Share {
	shares := make([]Share, len(mShares))
	for i, share := range mShares {
		shares[i] = *ShareFromModel(share, userId)
	}
	return shares
}

func ReducedSharesFromModelSlice(mShares models.ShareSlice) []ReducedShare {
	shares := make([]ReducedShare, len(mShares))
	for i, share := range mShares {
		shares[i] = ReducedShareFromModel(share)
	}
	return shares
}
