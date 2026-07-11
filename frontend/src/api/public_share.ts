import { Axios } from 'axios';
import { PublicShare, PublicShareIndexEntry, PublicShareInfo } from './resources';

export const createPublicShare = async (axios: Axios, objectId: string) => {
  const response = await axios.post(
    '/public-shares',
    { objectId },
    {
      headers: {
        'Content-Type': 'application/json',
      },
    }
  );

  return response.data as PublicShareInfo;
};

export const getPublicShare = async (axios: Axios, token: string) => {
  const response = await axios.get(`/public-shares/${token}`);
  return response.data as PublicShare;
};

export const getPublicShares = async (axios: Axios) => {
  const response = await axios.get('/public-shares');
  return response.data as PublicShareIndexEntry[];
};

export const deletePublicShare = async (axios: Axios, token: string) => {
  await axios.delete(`/public-shares/${token}`);
};

export const urlForPublicShare = (token: string) => {
  return `${window.location.origin}/public/${token}`;
};
