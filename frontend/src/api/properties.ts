import { Axios } from 'axios';
import { SchemaCollection } from './resources';

export const getSchemaCollection = async (axios: Axios) => {
  const response = await axios.get('/properties/schemas', {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (response.status !== 200) {
    throw `Got error ${response}`;
  }
  return response.data as SchemaCollection;
};
