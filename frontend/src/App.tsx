import { Layout } from './components/layout';
import { Config, getConfig } from './config/config';
import { AuthContext } from './context/auth';
import { Routes, Route, Navigate } from 'react-router';
import { AxiosContext } from './context/axios';
import axios from 'axios';
import { Login } from './routes/login';
import { Register } from './routes/register';
import { Logout } from './routes/logout';
import { useCookies } from 'react-cookie';
import { Things } from './routes/things/list';
import { CreateThing } from './routes/things/create';
import { ShowThing } from './routes/things/show';
import { ConfigContext } from './context/config';
import { EditThing } from './routes/things/edit';
import { RequireAuth } from './routes/private';
import { Lists } from './routes/lists/list';
import { CreateList } from './routes/lists/create';
import { ShowList } from './routes/lists/show';
import { Fragment, useCallback, useEffect, useMemo, useState } from 'react';
import { Search } from './routes/search';
import { getProfile } from './api/profile';
import { Profile } from './api/resources';
import { ShowProfile } from './routes/profile/show';
import { EditProfile } from './routes/profile/edit';
import { EditList } from './routes/lists/edit';
import { ShareThing } from './routes/things/share';
import { ShareList } from './routes/lists/share';
import { Images } from './routes/images/list';
import { SearchContext } from './context/search';
import { ShowFriends } from './routes/friends';
import { ShowNotifications } from './routes/notifications';
import { ShowUser } from './routes/users/show';
import { jwtDecode } from 'jwt-decode';
import { refreshTokens } from './api/auth';
import { ShowCart } from './routes/cart';
import { useCart } from './hooks/useCart';
import { CartContext } from './context/cart';
import React from 'react';

export const App = () => {
  const [config, setConfig] = useState<Config | null>(null);
  const [cookies] = useCookies(['stashsphere-info', 'stashsphere-refresh-info']);
  const infoCookie = cookies['stashsphere-info'] as string | undefined;
  const refreshInfoCookie = cookies['stashsphere-refresh-info'] as string | undefined;
  const [profileKey, setProfileKey] = useState(0);
  const [profile, setProfile] = useState<Profile | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const searchContextValue = useMemo(
    () => ({ searchTerm, setSearchTerm }),
    [searchTerm, setSearchTerm]
  );

  const loggedIn = infoCookie !== undefined || refreshInfoCookie !== undefined;

  useEffect(() => {
    getConfig()
      .then(setConfig)
      .catch((e) => console.error(e));
  }, []);

  const axiosInstance = useMemo(() => {
    if (config === null) {
      return null;
    }
    return axios.create({
      baseURL: config.apiHost + '/api',
      headers: {
        'Content-Type': 'application/json',
      },
      withCredentials: true,
    });
  }, [config]);

  useEffect(() => {
    if (infoCookie && axiosInstance) {
      getProfile(axiosInstance).then(setProfile);
    } else {
      setProfile(null);
    }
  }, [infoCookie, axiosInstance, profileKey]);

  useEffect(() => {
    const callback = () => {
      if (infoCookie && axiosInstance) {
        const decoded = jwtDecode(infoCookie);
        const now = new Date().getTime();
        const secondsToExpiry = (decoded.exp || 0) - now / 1000;
        // if the token expires in the next 15 minutes
        if (secondsToExpiry < 15 * 60) {
          console.log('Access Token expires in %d. Starting refresh process', secondsToExpiry);
          refreshTokens(axiosInstance);
        }
      }
      if (!infoCookie && refreshInfoCookie && axiosInstance) {
        refreshTokens(axiosInstance).then(() => {
          window.location.reload();
        });
      }
    };
    const id = setInterval(callback, 10000);
    callback();
    return () => clearInterval(id);
  }, [infoCookie, axiosInstance, refreshInfoCookie]);

  const invalidateProfile = useCallback(() => {
    setProfileKey((prev) => prev + 1);
  }, []);

  const authContextValue = useMemo(() => {
    return {
      profile,
      loggedIn,
      invalidateProfile,
    };
  }, [invalidateProfile, loggedIn, profile]);

  const [cart, addToCart, removeFromCart, clearCart, cartByUser] = useCart(axiosInstance, loggedIn);

  const ExternalWrapper = useMemo(() => {
    if (config?.strict === true) {
      console.warn('React StrictMode enabled.');
      return React.StrictMode;
    }
    return Fragment;
  }, [config]);

  if (config === null) {
    return 'Fetching config. Please wait.';
  }

  return (
    <ExternalWrapper>
      <ConfigContext.Provider value={config}>
        <AxiosContext.Provider value={axiosInstance}>
          <AuthContext.Provider value={authContextValue}>
            <SearchContext.Provider value={searchContextValue}>
              <CartContext.Provider
                value={{
                  cart,
                  addToCart,
                  removeFromCart,
                  clearCart,
                  cartByUser,
                }}
              >
                <Routes>
                  <Route path="/" element={<Layout />}>
                    <Route path="/" element={<Navigate to="/things" />} />
                    <Route path="/user/login" element={<Login />} />
                    <Route path="/user/logout" element={<Logout />} />
                    <Route path="/user/register" element={<Register />} />
                    <Route
                      path="/user/profile"
                      element={
                        <RequireAuth>
                          <ShowProfile />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/user/profile/edit"
                      element={
                        <RequireAuth>
                          <EditProfile />
                        </RequireAuth>
                      }
                    />

                    <Route
                      path="/things"
                      element={
                        <RequireAuth>
                          <Things />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/things/create"
                      element={
                        <RequireAuth>
                          <CreateThing />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/things/:thingId"
                      element={
                        <RequireAuth>
                          <ShowThing />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/things/:thingId/edit"
                      element={
                        <RequireAuth>
                          <EditThing />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/things/:thingId/share"
                      element={
                        <RequireAuth>
                          <ShareThing />
                        </RequireAuth>
                      }
                    />

                    <Route
                      path="/lists"
                      element={
                        <RequireAuth>
                          <Lists />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/lists/create"
                      element={
                        <RequireAuth>
                          <CreateList />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/lists/:listId"
                      element={
                        <RequireAuth>
                          <ShowList />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/lists/:listId/edit"
                      element={
                        <RequireAuth>
                          <EditList />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/lists/:listId/share"
                      element={
                        <RequireAuth>
                          <ShareList />
                        </RequireAuth>
                      }
                    />

                    <Route
                      path="/friends"
                      element={
                        <RequireAuth>
                          <ShowFriends />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/images"
                      element={
                        <RequireAuth>
                          <Images />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/search"
                      element={
                        <RequireAuth>
                          <Search />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/notifications"
                      element={
                        <RequireAuth>
                          <ShowNotifications />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/cart"
                      element={
                        <RequireAuth>
                          <ShowCart />
                        </RequireAuth>
                      }
                    />
                    <Route
                      path="/users/:userId"
                      element={
                        <RequireAuth>
                          <ShowUser />
                        </RequireAuth>
                      }
                    />
                  </Route>
                </Routes>
              </CartContext.Provider>
            </SearchContext.Provider>
          </AuthContext.Provider>
        </AxiosContext.Provider>
      </ConfigContext.Provider>
    </ExternalWrapper>
  );
};
