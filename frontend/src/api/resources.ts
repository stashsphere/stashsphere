export interface User {
  id: string;
  name: string;
}

export interface UserProfile {
  id: string;
  name: string;
  fullName: string;
  information: string;
  image: ReducedImage | null;
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

export type PropertyBoolean = {
  type: 'boolean';
  name: string;
  value: boolean;
  unit: undefined;
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

export type Property = PropertyString | PropertyFloat | PropertyDatetime | PropertyBoolean;
export type SharingState = 'private' | 'friends' | 'friends-of-friends';

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
  publicShares: PublicShareInfo[];
  sharingState: SharingState | null;
}

export interface ReducedThing {
  id: string;
  name: string;
  description: string;
  privateNote: string | null;
  createdAt: Date;
  owner: User;
}

export interface ThingsSummary {
  ownerIds: string[];
  totalCount: number;
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
  publicShares: PublicShareInfo[];
  sharingState: SharingState;
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
  fullName: string;
  information: string;
  image: ReducedImage | null;
  purgeAt: Date | null;
  emailVerified?: boolean;
  externalAuth: boolean;
  defaultSharingState: SharingState;
}

export interface Share {
  id: string;
  owner: User;
  targetUser: User;
}

export interface PublicShareInfo {
  id: string;
  createdAt: Date;
}

export interface PublicShareIndexEntry {
  id: string;
  createdAt: Date;
  type: 'thing' | 'list';
  objectId: string;
  objectName: string;
}

export interface PublicThing {
  id: string;
  name: string;
  description: string;
  createdAt: Date;
  ownerName: string;
  images: ReducedImage[];
  properties: Property[];
  quantity: number;
  quantityUnit: string;
}

export interface PublicList {
  id: string;
  name: string;
  createdAt: Date;
  ownerName: string;
  things: PublicThing[];
}

export type PublicShare =
  | {
      token: string;
      type: 'thing';
      thing: PublicThing;
    }
  | {
      token: string;
      type: 'list';
      list: PublicList;
    };

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

export type BaseNotification = {
  id: string;
  createdAt: Date;
  acknowledged: boolean;
};

export type FriendRequestNotification = BaseNotification & {
  contentType: 'FRIEND_REQUEST';
  content: {
    senderId: string;
    requestId: string;
  };
};

export type FriendRequestReactionNotification = BaseNotification & {
  contentType: 'FRIEND_REQUEST_REACTION';
  content: {
    accepted: boolean;
    requestId: string;
  };
};

export type ListSharedNotification = BaseNotification & {
  contentType: 'LIST_SHARED';
  content: {
    sharerId: string;
    listId: string;
  };
};

export type ThingSharedNotification = BaseNotification & {
  contentType: 'THING_SHARED';
  content: {
    sharerId: string;
    thingId: string;
  };
};

export type ThingsAddedToListNotification = BaseNotification & {
  contentType: 'THINGS_ADDED_TO_LIST';
  content: {
    listId: string;
    addedById: string;
  };
};

export type UnknownNotification = BaseNotification & {
  contentType: string;
  content: unknown;
};

export type StashSphereNotification =
  | FriendRequestNotification
  | ListSharedNotification
  | ThingSharedNotification
  | FriendRequestReactionNotification
  | ThingsAddedToListNotification;

export interface PagedNotifications extends Paged {
  notifications: StashSphereNotification[];
}

export type CartEntry = {
  thingId: string;
  ownerId: string;
  createdAt: Date;
};

export type Cart = {
  entries: CartEntry[];
};

export type PropertyNameSuggestion = {
  name: string;
  type: string;
  units: string[];
};

export type AutoCompleteResult = {
  completionType: 'name' | 'value';
  values: string[];
  suggestions: PropertyNameSuggestion[];
};

export type OIDCProviderInfo = {
  name: string;
  displayName: string;
};

export type InstanceInfo = {
  inviteRequired: boolean;
  oidcProviders: OIDCProviderInfo[];
};

export type SchemaEntry = {
  schema: {
    $schema: string;
    $id?: string;
    title?: string;
    description?: string;
    additionalProperties: boolean;
    properties: {
      name: { const: string };
      value: Record<string, unknown>;
      unit?: Record<string, unknown>;
    };
    required: string[];
    type: string;
    'x-aliases'?: string[];
    'x-icons'?: string[];
  };
  icons?: string[];
};

export type SchemaCollection = {
  schemas: Record<string, SchemaEntry>;
  aliases: Record<string, string>;
};
