import { useContext } from "react";
import { Header } from "./header";
import { AuthContext } from "../context/auth";
import { Outlet } from "react-router-dom";

export const Layout = () => {
  const authContext = useContext(AuthContext);
  return <>
    <Header userName={authContext.profile !== null ? authContext.profile.name : null} />
    <div className="bg-content md:p-2">
      <Outlet />
    </div>
  </>;
};
