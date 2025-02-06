import { useContext, useEffect, useMemo } from "react";
import { AxiosContext } from "../context/axios";
import { useNavigate } from "react-router-dom";
import { AuthContext } from "../context/auth";

export const Logout = () => {
  const navigate = useNavigate();
  const axiosInstance = useContext(AxiosContext);
  const authContext = useContext(AuthContext);

  const logout = useMemo(() => async () => {
    if (axiosInstance === null) {
      return;
    }
    try {
      const response = await axiosInstance.delete("/user/logout");
      console.log(response.data);
    } catch (error) {
      console.error(error);
    }
  }, [axiosInstance]);
  
  useEffect(() => {
    if (authContext.loggedIn) {
      logout();
    } else {
      navigate("/user/login");
    }
  }, [navigate, logout, authContext.loggedIn]);

  return <></>;
};
