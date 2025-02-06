export interface Owner {
  id: string;
  name: string;
}

export interface ReducedList {
  id: string;
  name: string;
}

export type PropertyString = {
  type: "string";
  name: string;
  value: string;
  unit: undefined;
}

export type PropertyFloat = {
  type: "float";
  name: string;
  value: number;
  unit: string;
}

export type PropertyDatetime = {
  type: "datetime";
  name: string;
  value: string;
  unit: undefined;
}

export type ThingActions = {
  canEdit: boolean;
  canShare: boolean;
  canDelete: boolean;
}

export type Property = PropertyString | PropertyFloat | PropertyDatetime;

export interface Thing {
  id: string;
  name: string;
  createdAt: Date;
  owner: Owner;
  lists: ReducedList[];
  images: ReducedImage[];
  properties: Property[];
  description: string;
  privateNote: string | null;
  actions: ThingActions;
  quantity: number;
  quantityUnit: string;
}

export interface ReducedThing {
  id: string;
  name: string;
  description: string;
  privateNote: string | null;
  createdAt: Date;
  owner: Owner;
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
  owner: Owner
}

export type ImageActions = {
  canDelete: boolean;
}

export interface Image {
  id: string;
  name: string;
  createdAt: Date;
  owner: Owner
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
  owner: Owner;
  things: Thing[];
  actions: ThingActions;
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
  owner: Owner;
}