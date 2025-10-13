import { useCallback, useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { AuthContext } from '../../context/auth';
import { PagedNotifications } from '../../api/resources';
import { fetchNotifications } from '../../api/notification';
import { Pages } from '../../components/pages';
import { NotificationItem } from '../../components/notification';

export const ShowNotifications = () => {
  const axiosInstance = useContext(AxiosContext);
  const authContext = useContext(AuthContext);

  const [currentPage, setCurrentPage] = useState(0);
  const [notifications, setNotifications] = useState<PagedNotifications | undefined>(undefined);
  const [mutateKey, setMutateKey] = useState(0);

  const loggedIn = authContext.loggedIn;

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    if (!loggedIn) {
      return;
    }
    fetchNotifications(axiosInstance, false, currentPage)
      .then(setNotifications)
      .catch((reason) => {
        console.log(reason);
      });
  }, [axiosInstance, loggedIn, currentPage, mutateKey]);

  const mutate = useCallback(() => {
    setMutateKey((prev) => prev + 1);
  }, []);

  if (!notifications) {
    return <p>Loading...</p>;
  }
  return (
    <>
      {notifications.totalCount === 0
        ? 'No notifications'
        : notifications.notifications.map((v) => (
            <NotificationItem onAcknowledge={() => mutate()} notification={v} key={v.id} />
          ))}
      {notifications.notifications.length > 0 && (
        <Pages
          currentPage={currentPage}
          onPageChange={(n) => setCurrentPage(n)}
          pages={notifications.totalPageCount}
        />
      )}
    </>
  );
};
