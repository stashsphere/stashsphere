import { Axios } from 'axios';
import { PagedNotifications } from './resources';

export const fetchNotifications = async (
  axios: Axios,
  onlyAcknowledged: boolean,
  currentPage: number
) => {
  const response = await axios.get(
    `/notifications?page=${currentPage}&onlyAcknowledged=${onlyAcknowledged}`
  );

  const notifications = response.data as PagedNotifications;
  return notifications;
};

export const acknowledgeNotification = async (axios: Axios, id: string) => {
  return axios.patch(`/notifications/${id}`);
};
