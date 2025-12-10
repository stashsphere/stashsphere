import { Axios } from 'axios';
import { InstanceInfo } from './resources';

export const getInstanceInfo = async (axios: Axios) => {
  const response = await axios.get('/info', {
    headers: {
      'Content-Type': 'application/json',
    },
  });
  return response.data as InstanceInfo;
};
