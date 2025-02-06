import { useContext } from "react";
import { AuthContext } from "../../context/auth";
import { ProfileDetails } from "../../components/profile_details";

export const ShowProfile = () => {
  const authContext = useContext(AuthContext);
  const profile = authContext.profile;
  
  if (!profile) {
    return <div>No Profile!</div>;
  }
  return <ProfileDetails profile={profile} />
}