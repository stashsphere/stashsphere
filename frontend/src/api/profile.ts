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

export type ProfileUpdateParams = {
  name: string;
  fullName: string;
  information: string;
  imageId: string | null;
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

export type UpdatePasswordParams = {
  oldPassword: string;
  newPassword: string;
};

export const updatePassword = async (axios: Axios, params: UpdatePasswordParams) => {
  await axios.patch('/user/password', params, {
    headers: {
      'Content-Type': 'application/json',
    },
  });
};

export const scheduleDeletion = async (axios: Axios, password: string) => {
  const response = await axios.post(
    '/user/deletion',
    { password },
    {
      headers: {
        'Content-Type': 'application/json',
      },
    }
  );

  return response.data as Profile;
};

export const cancelDeletion = async (axios: Axios) => {
  const response = await axios.delete('/user/deletion', {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  return response.data as Profile;
};

export const requestEmailVerification = async (axios: Axios) => {
  await axios.post(
    '/user/email-verification/request',
    {},
    {
      headers: {
        'Content-Type': 'application/json',
      },
    }
  );
};

export const verifyEmail = async (axios: Axios, code: string) => {
  await axios.post(
    '/user/email-verification/verify',
    { code },
    {
      headers: {
        'Content-Type': 'application/json',
      },
    }
  );
};
