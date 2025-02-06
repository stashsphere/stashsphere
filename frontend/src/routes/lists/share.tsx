import { useNavigate, useParams } from "react-router-dom";
import { ShareEditor } from "../../components/share_editor";
import { useContext, useEffect, useState } from "react";
import { AxiosContext } from "../../context/axios";
import { getList } from "../../api/lists";
import { Profile, List } from "../../api/resources";
import { AuthContext } from "../../context/auth";
import { getAllProfiles } from "../../api/profile";
import { shareObject } from "../../api/share";

export const ShareList = () => {
    const { listId } = useParams();
    const navigate = useNavigate();

    const [list, setList] = useState<null | List>(null);
    const axiosInstance = useContext(AxiosContext);
    const authContext = useContext(AuthContext);
    const profile = authContext.profile;

    const [profiles, setProfiles] = useState<Profile[]>([]);

    useEffect(() => {
        if (!axiosInstance || listId === undefined) {
            return;
        }
        getList(axiosInstance, listId).then(setList);
    }, [axiosInstance, listId]);

    useEffect(() => {
        if (!axiosInstance) {
            return;
        }
        getAllProfiles(axiosInstance).then(setProfiles);
    }, [axiosInstance]);

    if (list === null || profile === null) {
        return <h1>Loading</h1>;
    }

    const onShare = async (targetUserProfile: Profile) => {
        if (!axiosInstance) {
            return;
        }
        console.log("Sharing List to", targetUserProfile);
        const share = await shareObject(axiosInstance, {
            objectId: list.id,
            targetUserId: targetUserProfile.id
        });
        console.log("Share result", share);
        navigate(`/lists/${listId}`)
    };

    return (
        <ShareEditor type={"list"} list={list} profiles={profiles} userProfile={profile} onSubmit={onShare} />
    )
}