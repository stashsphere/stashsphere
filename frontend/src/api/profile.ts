import { Axios } from 'axios';
import { Profile } from './resources';

export const getProfile = async (axios: Axios) => {
  const response = await axios.get('/user/profile', {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (response.status !== 200) {
    throw `Got error ${response}`;
  }
  return response.data as Profile;
};

type ProfileUpdateParams = {
  name: string;
};

export const patchProfile = async (axios: Axios, params: ProfileUpdateParams) => {
  const response = await axios.patch('/user/profile', params, {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  const profile = response.data as Profile;
  return profile;
};
