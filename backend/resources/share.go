package resources

import (
	"github.com/stashsphere/backend/models"
)

type ShareType string

const (
	ThingShareType ShareType = "thing"
	ListShareType  ShareType = "list"
)

// Share is a ThingShare or a ListShare, discriminated by its `type` field.
type Share interface {
	ShareType() ShareType
}

func (ThingShare) ShareType() ShareType { return ThingShareType }
func (ListShare) ShareType() ShareType  { return ListShareType }

type ThingShare struct {
	Id         string       `json:"id"`
	Type       ShareType    `json:"type"`
	TargetUser User         `json:"targetUser"`
	Owner      User         `json:"owner"`
	Object     ReducedThing `json:"object"`
}

type ListShare struct {
	Id         string      `json:"id"`
	Type       ShareType   `json:"type"`
	TargetUser User        `json:"targetUser"`
	Owner      User        `json:"owner"`
	Object     ReducedList `json:"object"`
}

type ReducedShare struct {
	TargetUser User   `json:"targetUser"`
	Owner      User   `json:"owner"`
	Id         string `json:"id"`
}

func ShareFromModel(share *models.Share, userId string) Share {
	targetUser := UserFromModel(share.R.TargetUser)
	owner := UserFromModel(share.R.Owner)
	if len(share.R.Lists) > 0 {
		return ListShare{
			Id:         share.ID,
			Type:       ListShareType,
			TargetUser: targetUser,
			Owner:      owner,
			Object:     *ReducedListFromModel(share.R.Lists[0], userId),
		}
	}
	return ThingShare{
		Id:         share.ID,
		Type:       ThingShareType,
		TargetUser: targetUser,
		Owner:      owner,
		Object:     *ReducedThingFromModel(share.R.Things[0], userId),
	}
}

func ReducedShareFromModel(s *models.Share) ReducedShare {
	return ReducedShare{TargetUser: UserFromModel(s.R.TargetUser), Owner: UserFromModel(s.R.Owner), Id: s.ID}
}

func ReducedSharesFromModelSlice(mShares models.ShareSlice) []ReducedShare {
	shares := make([]ReducedShare, len(mShares))
	for i, share := range mShares {
		shares[i] = ReducedShareFromModel(share)
	}
	return shares
}
