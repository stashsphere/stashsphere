import { Axios } from 'axios';
import { User } from './resources';

export const getAllUsers = async (axios: Axios) => {
  const response = await axios.get('/users', {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (response.status !== 200) {
    throw `Go error ${response}`;
  }
  return response.data as User[];
};

export const getUser = async (axios: Axios, userId: string) => {
  const response = await axios.get(`/users/${userId}`, {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (response.status !== 200) {
    throw `Go error ${response}`;
  }
  return response.data as User;
};
