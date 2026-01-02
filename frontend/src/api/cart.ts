import { Axios } from 'axios';
import { Cart } from './resources';

export const getCart = async (axios: Axios) => {
  const response = await axios.get(`/cart`, {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (response.status != 200) {
    throw `Got error ${response}`;
  }

  const cart = response.data as Cart;
  return cart;
};

export const putCart = async (axios: Axios, thingIds: string[]) => {
  const response = await axios.put(
    `/cart`,
    {
      thingIds,
    },
    {
      headers: {
        'Content-Type': 'application/json',
      },
    }
  );

  if (response.status != 200) {
    throw `Got error ${response}`;
  }

  const cart = response.data as Cart;
  return cart;
};
