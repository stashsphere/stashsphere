package resources

import "github.com/stashsphere/backend/operations"

type AccessReason struct {
	Type                 string  `json:"type"`
	ShareOwnerId         *string `json:"shareOwnerId,omitempty"`
	FriendId             *string `json:"friendId,omitempty"`
	FriendOfFriendId     *string `json:"friendOfFriendId,omitempty"`
	ListId               *string `json:"listId,omitempty"`
}

func AccessReasonFromOperations(reason operations.AccessReasonInformation) *AccessReason {
	if reason == nil {
		return nil
	}

	switch r := reason.(type) {
	case operations.AccessReasonOwner:
		return &AccessReason{
			Type: "owner",
		}
	case operations.AccessReasonSharedDirectly:
		return &AccessReason{
			Type:         "shared-directly",
			ShareOwnerId: &r.ShareOwnerId,
		}
	case operations.AccessReasonFriend:
		return &AccessReason{
			Type:     "friend",
			FriendId: &r.FriendId,
		}
	case operations.AccessReasonFriendOfFriend:
		return &AccessReason{
			Type:             "friend-of-friend",
			FriendId:         &r.FriendId,
			FriendOfFriendId: &r.FriendOfFriendId,
		}
	case operations.AccessReasonListSharedDirectly:
		return &AccessReason{
			Type:         "list-shared-directly",
			ShareOwnerId: &r.ShareOwnerId,
			ListId:       &r.ListId,
		}
	case operations.AccessReasonListFriend:
		return &AccessReason{
			Type:     "list-friend",
			FriendId: &r.FriendId,
			ListId:   &r.ListId,
		}
	case operations.AccessReasonListFriendOfFriend:
		return &AccessReason{
			Type:             "list-friend-of-friend",
			FriendId:         &r.FriendId,
			FriendOfFriendId: &r.FriendOfFriendId,
			ListId:           &r.ListId,
		}
	default:
		return nil
	}
}
