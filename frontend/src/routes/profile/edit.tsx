import { useContext } from "react";
import { useNavigate } from "react-router-dom";
import { AuthContext } from "../../context/auth";
import { AxiosContext } from "../../context/axios";
import { EditableProfile, ProfileEditor } from "../../components/profile_editor";
import { patchProfile } from "../../api/profile";
import { YellowButton } from "../../components/button";

export const EditProfile = () => {
    const axiosInstance = useContext(AxiosContext);
    const navigate = useNavigate();
    const authContext = useContext(AuthContext);
    const profile = authContext.profile;

    const update = async (data: EditableProfile) => {
        if (!axiosInstance) {
            return;
        }
        await patchProfile(axiosInstance, { ...data });
        console.log("Updated profile");
        authContext.invalidateProfile();
        navigate("/user/profile");
    }

    if (!profile) {
        return <div>No Profile</div>;
    }

    return <ProfileEditor profile={profile} onUpdateProfile={update}>
        <YellowButton type="submit">Update Profile</YellowButton>
    </ProfileEditor>
}