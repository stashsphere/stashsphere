package notifications

const (
	NotifyFriendRequest = "FRIEND_REQUEST"
)

type StashsphereNotification interface {
	ContentType() string
}

type FriendRequest struct {
	RequestId string `json:"requestId"`
	SenderId  string `json:"senderId"`
}

func (n FriendRequest) ContentType() string {
	return NotifyFriendRequest
}
