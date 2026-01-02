import { Axios } from 'axios';
import { AutoCompleteResult, SearchResult } from './resources';

export const getSearchResult = async (axios: Axios, query: string): Promise<SearchResult> => {
  const response = await axios.get(`/search`, {
    headers: {
      'Content-Type': 'application/json',
    },
    params: {
      query: query,
    },
  });

  if (response.status != 200) {
    throw `Got error ${response}`;
  }

  return response.data as SearchResult;
};

export const getAutoComplete = async (
  axios: Axios,
  name: string,
  value: string | null
): Promise<AutoCompleteResult> => {
  const response = await axios.get(`/search/property_auto_complete`, {
    headers: {
      'Content-Type': 'application/json',
    },
    params: {
      name,
      value,
    },
  });

  if (response.status != 200) {
    throw `Got error ${response}`;
  }

  return response.data as AutoCompleteResult;
};
