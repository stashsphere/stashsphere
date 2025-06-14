import { Axios } from 'axios';
import { List, PagedLists, SharingState } from './resources';

export const getLists = async (axios: Axios, currentPage: number) => {
  const response = await axios.get(`/lists?page=${currentPage}`, {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (response.status != 200) {
    throw `Got error ${response}`;
  }

  const lists = response.data as PagedLists;
  return lists;
};

export const getList = async (axios: Axios, id: string) => {
  const response = await axios.get(`/lists/${id}`, {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (response.status != 200) {
    throw `Got error ${response}`;
  }

  const list = response.data as List;
  return list;
};

export interface CreateListParams {
  name: string;
  thingsIds: string[];
  sharingState: SharingState;
}

export const createList = async (axios: Axios, params: CreateListParams) => {
  const response = await axios.post('/lists', params, {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  const list = response.data as List;
  return list;
};

export interface UpdateListParams {
  name: string;
  thingIds: string[];
  sharingState: SharingState;
}

export const updateList = async (axios: Axios, id: string, params: UpdateListParams) => {
  const response = await axios.patch('/lists/' + id, params, {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  const thing = response.data as List;
  return thing;
};

export const updateListParamsFromList = (list: List): UpdateListParams => {
  const params = {
    name: list.name,
    thingIds: list.things.map((t) => t.id),
    sharingState: list.sharingState || 'private',
  };
  return params;
};
