import { Axios } from 'axios';

export const login = async (axios: Axios, email: string, password: string) => {
  await axios.post(
    '/user/login',
    {
      email: email,
      password: password,
    },
    {
      headers: {
        'Content-Type': 'application/json',
      },
    }
  );
  return null;
};

export const refreshTokens = async (axios: Axios) => {
  await axios.post('/user/refresh', null, {
    headers: {
      'Content-Type': 'application/json',
    },
  });
  return null;
};

export const logout = async (axios: Axios) => {
  await axios.delete('/user/logout');
  return null;
};
