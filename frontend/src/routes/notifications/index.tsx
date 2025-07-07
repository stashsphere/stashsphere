import { useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { AuthContext } from '../../context/auth';
import { PagedNotifications } from '../../api/resources';
import { fetchNotifications } from '../../api/notification';
import { Pages } from '../../components/pages';
import { NotificationItem } from '../../components/notification';

export const ShowNotifications = () => {
  const axiosInstance = useContext(AxiosContext);
  const authContext = useContext(AuthContext);

  const [notifications, setNotifications] = useState<PagedNotifications | undefined>(undefined);
  const [currentPage, setCurrentPage] = useState(0);
  const [mutateKey, setMutateKey] = useState(0);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    if (!authContext.loggedIn) {
      return;
    }
    fetchNotifications(axiosInstance, false, currentPage)
      .then(setNotifications)
      .catch((reason) => {
        console.log(reason);
      });
  }, [axiosInstance, authContext, currentPage, mutateKey]);

  const mutate = () => {
    setMutateKey(mutateKey + 1);
  };

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
