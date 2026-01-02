import { Axios } from 'axios';
import { PagedThings, SharingState, Thing, ThingsSummary } from './resources';

export const getThings = async (
  axios: Axios,
  currentPage: number,
  perPage: number,
  ownerIds: string[],
  searchTerm: string | undefined
) => {
  const response = await axios.get(`/things`, {
    headers: {
      'Content-Type': 'application/json',
    },
    params: {
      page: currentPage,
      filterOwnerId: ownerIds,
      perPage: perPage,
      searchTerm: searchTerm,
    },
    paramsSerializer: {
      indexes: null,
    },
  });

  if (response.status != 200) {
    throw `Got error ${response}`;
  }

  const things = response.data as PagedThings;
  return things;
};

export const getThingsSummary = async (axios: Axios) => {
  const response = await axios.get(`/things/summary`, {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (response.status != 200) {
    throw `Got error ${response}`;
  }

  const summary = response.data as ThingsSummary;
  return summary;
};

export const getThing = async (axios: Axios, id: string) => {
  const response = await axios.get(`/things/${id}`, {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (response.status != 200) {
    throw `Got error ${response}`;
  }
  return response.data as Thing;
};

export type CreatePropertyStringParam = {
  name: string;
  value: string;
  type: 'string';
};
export type CreatePropertyFloatParam = {
  name: string;
  value: number;
  type: 'float';
  unit?: string;
};
export type CreatePropertyDatetimeParam = { name: string; value: string; type: 'datetime' };
export type CreatePropertyParam =
  | CreatePropertyDatetimeParam
  | CreatePropertyStringParam
  | CreatePropertyFloatParam;

export interface CreateThingParams {
  name: string;
  privateNote: string;
  description: string;
  imagesIds: string[];
  properties: CreatePropertyParam[];
  quantity: number;
  quantityUnit: string;
  sharingState: SharingState;
}

export const createThing = async (axios: Axios, params: CreateThingParams): Promise<Thing> => {
  const response = await axios.post('/things', params, {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  const thing = response.data as Thing;
  return thing;
};

export type UpdateThingParams = CreateThingParams;

export const updateThing = async (axios: Axios, id: string, params: UpdateThingParams) => {
  const response = await axios.patch('/things/' + id, params, {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  const thing = response.data as Thing;
  return thing;
};

export const updateThingParamsFromThing = (thing: Thing): UpdateThingParams => {
  const params = {
    name: thing.name,
    privateNote: thing.privateNote || '',
    description: thing.description,
    imagesIds: thing.images.map((x) => x.id),
    properties: thing.properties,
    quantity: thing.quantity,
    quantityUnit: thing.quantityUnit,
    sharingState: thing.sharingState || 'private',
  };
  return params;
};

export const deleteThing = async (axios: Axios, id: string) => {
  return await axios.delete(`/things/` + id);
};
