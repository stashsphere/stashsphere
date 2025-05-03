import { useContext, useEffect, useState } from 'react';
import { Header } from './header';
import { AuthContext } from '../context/auth';
import { Outlet } from 'react-router';
import { AxiosContext } from '../context/axios';
import { fetchNotifications } from '../api/notification';

export const Layout = () => {
  const authContext = useContext(AuthContext);
  const axiosInstance = useContext(AxiosContext);

  const [hasUnacknowledgedNotifications, setHasUnacknowledgedNotifications] = useState(false);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    if (!authContext.loggedIn) {
      return;
    }
    fetchNotifications(axiosInstance, true, 0).then((v) => {
      if (v.totalCount > 0) {
        setHasUnacknowledgedNotifications(true);
      } else {
        setHasUnacknowledgedNotifications(false);
      }
    });
  }, [axiosInstance, authContext]);

  return (
    <>
      <Header
        userName={authContext.profile !== null ? authContext.profile.name : null}
        hasUnacknowledgedNotifications={hasUnacknowledgedNotifications}
      />
      <div className="bg-content md:p-2">
        <Outlet />
      </div>
    </>
  );
};
