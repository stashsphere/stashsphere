import "./App.css";
import { Layout } from "./components/layout";
import { Config, getConfig } from "./config/config";
import { AuthContext } from "./context/auth";
import { Routes, Route, Navigate } from "react-router-dom";
import { AxiosContext } from "./context/axios";
import axios from "axios";
import { Login } from "./routes/login";
import { Register } from "./routes/register";
import { Logout } from "./routes/logout";
import { useCookies } from "react-cookie";
import { Things } from "./routes/things/list";
import { CreateThing } from "./routes/things/create";
import { ShowThing } from "./routes/things/show";
import { ConfigContext } from "./context/config";
import { EditThing } from "./routes/things/edit";
import { RequireAuth } from "./routes/private";
import { Lists } from "./routes/lists/list";
import { CreateList } from "./routes/lists/create";
import { ShowList } from "./routes/lists/show";
import { useEffect, useMemo, useState } from "react";
import { Search } from "./routes/search";
import { getProfile } from "./api/profile";
import { Profile } from "./api/resources";
import { ShowProfile } from "./routes/profile/show";
import { EditProfile } from "./routes/profile/edit";
import { EditList } from "./routes/lists/edit";
import { ShareThing } from "./routes/things/share";
import { ShareList } from "./routes/lists/share";
import { Images } from "./routes/images/list";
import { SearchContext } from "./context/search";

export const App = () => {
  const [config, setConfig] = useState<Config | null>(null);
  const [cookies] = useCookies(["stashsphere-info"]);
  const infoCookie = cookies["stashsphere-info"];
  const [profileKey, setProfileKey] = useState(0);
  const [profile, setProfile] = useState<Profile | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const searchContextValue = useMemo(
    () => ({ searchTerm, setSearchTerm }),
    [searchTerm, setSearchTerm ],
  );

  useEffect(() => {
    getConfig().then(setConfig).catch(e => console.error(e))
  }, [])

  const axiosInstance = useMemo(() => {
    if (config === null) {
      return null;
    }
    return axios.create({
      baseURL: config.apiHost + "/api",
      headers: {
        "Content-Type": "application/json",
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
  
  if (config === null) {
    return "Fetching config. Please wait."
  }

  return (
    <ConfigContext.Provider value={config}>
      <AxiosContext.Provider value={axiosInstance}>
        <AuthContext.Provider value={{
          profile, loggedIn: infoCookie !== undefined, invalidateProfile: () => {
            setProfileKey(profileKey + 1);
          }
        }}>
        <SearchContext.Provider value={searchContextValue}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="/" element={<Navigate to="/things" />} />
              <Route path="/user/login" element={<Login />} />
              <Route path="/user/logout" element={<Logout />} />
              <Route path="/user/register" element={<Register />} />
              <Route path="/user/profile" element={<RequireAuth><ShowProfile /></RequireAuth>} />
              <Route path="/user/profile/edit" element={<RequireAuth><EditProfile /></RequireAuth>} />

              <Route path="/things" element={<RequireAuth><Things /></RequireAuth>} />
              <Route path="/things/create" element={<RequireAuth><CreateThing /></RequireAuth>} />
              <Route path="/things/:thingId" element={<RequireAuth><ShowThing /></RequireAuth>} />
              <Route path="/things/:thingId/edit" element={<RequireAuth><EditThing /></RequireAuth>} />
              <Route path="/things/:thingId/share" element={<RequireAuth><ShareThing /></RequireAuth>} />

              <Route path="/lists" element={<RequireAuth><Lists /></RequireAuth>} />
              <Route path="/lists/create" element={<RequireAuth><CreateList /></RequireAuth>} />
              <Route path="/lists/:listId" element={<RequireAuth><ShowList /></RequireAuth>} />
              <Route path="/lists/:listId/edit" element={<RequireAuth><EditList /></RequireAuth>} />
              <Route path="/lists/:listId/share" element={<RequireAuth><ShareList /></RequireAuth>} />

              <Route path="/images" element={<RequireAuth><Images /></RequireAuth>} />

              <Route path="/search" element={<RequireAuth><Search /></RequireAuth>} />
            </Route>
          </Routes>
          </SearchContext.Provider>
        </AuthContext.Provider>
      </AxiosContext.Provider>
    </ConfigContext.Provider>
  );
};
