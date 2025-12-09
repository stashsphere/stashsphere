package notifications

const (
	NotifyFriendRequestSent     = "FRIEND_REQUEST"
	NotifyThingShared           = "THING_SHARED"
	NotifyListShared            = "LIST_SHARED"
	NotifyFriendRequestReaction = "FRIEND_REQUEST_REACTION"
	NotifyThingsAddedToList     = "THINGS_ADDED_TO_LIST"
)

type StashsphereNotification interface {
	ContentType() string
}

type FriendRequest struct {
	RequestId string `json:"requestId"`
	SenderId  string `json:"senderId"`
}

func (n FriendRequest) ContentType() string {
	return NotifyFriendRequestSent

}

type FriendRequestReaction struct {
	RequestId string `json:"requestId"`
	Accepted  bool   `json:"accepted"`
}

func (n FriendRequestReaction) ContentType() string {
	return NotifyFriendRequestReaction
}

type ThingShared struct {
	ThingId  string `json:"thingId"`
	SharerId string `json:"sharerId"`
}

func (n ThingShared) ContentType() string {
	return NotifyThingShared
}

type ListShared struct {
	ListId   string `json:"listId"`
	SharerId string `json:"sharerId"`
}

func (n ListShared) ContentType() string {
	return NotifyListShared
}

type ThingsAddedToList struct {
	ListId string `json:"listId"`
}

func (n ThingsAddedToList) ContentType() string {
	return NotifyThingsAddedToList
}
