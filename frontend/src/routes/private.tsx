import { useContext } from 'react';
import { AuthContext } from '../context/auth';
import { Navigate, useLocation } from 'react-router';

export const RequireAuth = ({ children }: { children: React.ReactNode }) => {
  const { loggedIn } = useContext(AuthContext);
  const location = useLocation();

  return loggedIn === true ? (
    children
  ) : (
    <Navigate to="/user/login" replace state={{ path: location.pathname }} />
  );
};
