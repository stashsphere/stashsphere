package resources

import (
	"time"

	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/services"
)

type FriendRequest struct {
	Id        string    `json:"id"`
	Sender    User      `json:"sender"`
	Receiver  User      `json:"receiver"`
	CreatedAt time.Time `json:"createdAt"`
	State     string    `json:"state"`
}

func FriendRequestFromModel(model *models.FriendRequest, userId string) *FriendRequest {
	return &FriendRequest{
		Id:        model.ID,
		Sender:    UserFromModel(model.R.Sender),
		Receiver:  UserFromModel(model.R.Receiver),
		CreatedAt: model.CreatedAt,
		State:     model.State.String(),
	}
}

type FriendRequestResponse struct {
	Received []FriendRequest `json:"received"`
	Sent     []FriendRequest `json:"sent"`
}

func FriendRequestsFromModelSlice(mRequests models.FriendRequestSlice, userId string) []FriendRequest {
	requests := make([]FriendRequest, len(mRequests))
	for i, model := range mRequests {
		requests[i] = *FriendRequestFromModel(model, userId)
	}
	return requests
}

func FriendRequestsResponseFromResult(result *services.FriendRequestsResult, userId string) *FriendRequestResponse {
	return &FriendRequestResponse{
		Received: FriendRequestsFromModelSlice(result.Received, userId),
		Sent:     FriendRequestsFromModelSlice(result.Sent, userId),
	}
}

type FriendShip struct {
	Friend        User           `json:"friend"`
	FriendRequest *FriendRequest `json:"request"`
}

func FriendShipFromModel(model *models.Friendship, userId string) *FriendShip {
	friend := model.R.Friend1
	if model.R.Friend1.ID == userId {
		friend = model.R.Friend2
	}
	return &FriendShip{
		Friend:        UserFromModel(friend),
		FriendRequest: FriendRequestFromModel(model.R.FriendRequest, userId),
	}
}

type FriendShipsResponse struct {
	FriendShips []*FriendShip `json:"friendShips"`
}

func FriendShipsFromModelSlice(mFriends models.FriendshipSlice, userId string) []*FriendShip {
	friends := make([]*FriendShip, len(mFriends))
	for i, model := range mFriends {
		friends[i] = FriendShipFromModel(model, userId)
	}
	return friends
}

func FriendShipsResponseFromModel(mFriends models.FriendshipSlice, userId string) *FriendShipsResponse {
	return &FriendShipsResponse{
		FriendShips: FriendShipsFromModelSlice(mFriends, userId),
	}
}
