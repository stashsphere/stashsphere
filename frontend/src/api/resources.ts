export interface User {
  id: string;
  name: string;
}

export interface ReducedList {
  id: string;
  name: string;
}

export type PropertyString = {
  type: 'string';
  name: string;
  value: string;
  unit: undefined;
};

export type PropertyFloat = {
  type: 'float';
  name: string;
  value: number;
  unit: string;
};

export type PropertyDatetime = {
  type: 'datetime';
  name: string;
  value: string;
  unit: undefined;
};

export type ThingActions = {
  canEdit: boolean;
  canShare: boolean;
  canDelete: boolean;
};

export type Property = PropertyString | PropertyFloat | PropertyDatetime;

export interface Thing {
  id: string;
  name: string;
  createdAt: Date;
  owner: User;
  lists: ReducedList[];
  images: ReducedImage[];
  properties: Property[];
  description: string;
  privateNote: string | null;
  actions: ThingActions;
  quantity: number;
  quantityUnit: string;
  shares: Share[];
}

export interface ReducedThing {
  id: string;
  name: string;
  description: string;
  privateNote: string | null;
  createdAt: Date;
  owner: User;
}

export interface Paged {
  perPage: number;
  page: number;
  totalPageCount: number;
  totalCount: number;
}

export interface ReducedImage {
  id: string;
  name: string;
  createdAt: Date;
  hash: string;
  owner: User;
}

export type ImageActions = {
  canDelete: boolean;
};

export interface Image {
  id: string;
  name: string;
  createdAt: Date;
  owner: User;
  hash: string;
  things: ReducedThing[];
  actions: ImageActions;
}

export interface PagedThings extends Paged {
  things: Thing[];
}

export interface PagedImages extends Paged {
  images: Image[];
}

export interface List {
  id: string;
  name: string;
  createdAt: Date;
  owner: User;
  things: Thing[];
  actions: ThingActions;
  shares: Share[];
}

export interface PagedLists extends Paged {
  lists: List[];
}

export interface SearchResult {
  things: Thing[];
  lists: List[];
}

export interface Profile {
  id: string;
  name: string;
  email: string;
}

export interface Share {
  id: string;
  owner: User;
  targetUser: User;
}

export interface FriendRequest {
  id: string;
  sender: User;
  receiver: User;
  createdAt: Date;
  state: 'pending' | 'accepted' | 'rejected';
}

export interface FriendRequestResponse {
  received: FriendRequest[];
  sent: FriendRequest[];
}

export interface FriendShip {
  friend: User;
  friendRequest: FriendRequest;
}

export interface FriendShips {
  friendShips: FriendShip[];
}

export interface BaseNotification {
  id: string;
  createdAt: Date;
  acknowledged: boolean;
}

export interface FriendRequestNotification extends BaseNotification {
  contentType: 'FRIEND_REQUEST';
  content: {
    senderId: string;
    requestId: string;
  };
}

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface StashsphereNotification extends FriendRequestNotification {}

export interface PagedNotifications extends Paged {
  notifications: StashsphereNotification[];
}
