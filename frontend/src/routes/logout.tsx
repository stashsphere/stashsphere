import { useContext, useEffect, useMemo } from 'react';
import { AxiosContext } from '../context/axios';
import { useNavigate } from 'react-router';
import { AuthContext } from '../context/auth';

export const Logout = () => {
  const navigate = useNavigate();
  const axiosInstance = useContext(AxiosContext);
  const authContext = useContext(AuthContext);

  const logout = useMemo(
    () => async () => {
      if (axiosInstance === null) {
        return;
      }
      try {
        await axiosInstance.delete('/user/logout');
      } catch (error) {
        console.error(error);
      }
    },
    [axiosInstance]
  );

  useEffect(() => {
    if (authContext.loggedIn) {
      logout();
    } else {
      navigate('/user/login');
    }
  }, [navigate, logout, authContext.loggedIn]);

  return <></>;
};
