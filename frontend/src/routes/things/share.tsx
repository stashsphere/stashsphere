import { useNavigate, useParams } from "react-router-dom";
import { ShareEditor } from "../../components/share_editor";
import { useContext, useEffect, useState } from "react";
import { AxiosContext } from "../../context/axios";
import { getThing } from "../../api/things";
import { Profile, Thing } from "../../api/resources";
import { AuthContext } from "../../context/auth";
import { getAllProfiles } from "../../api/profile";
import { shareObject } from "../../api/share";

export const ShareThing = () => {
    const { thingId } = useParams();
    const navigate = useNavigate();

    const [thing, setThing] = useState<null | Thing>(null);
    const [mutateKey, setMutateKey] = useState(0);
    const axiosInstance = useContext(AxiosContext);
    const authContext = useContext(AuthContext);
    const profile = authContext.profile;

    const [profiles, setProfiles] = useState<Profile[]>([]);

    useEffect(() => {
        if (!axiosInstance || thingId === undefined) {
            return;
        }
        getThing(axiosInstance, thingId).then(setThing);
    }, [axiosInstance, thingId, mutateKey]);

    useEffect(() => {
        if (!axiosInstance) {
            return;
        }
        getAllProfiles(axiosInstance).then(setProfiles);
    }, [axiosInstance]);

    if (thing === null || profile === null) {
        return <h1>Loading</h1>;
    }

    const onShare = async (targetUserProfile: Profile) => {
        if (!axiosInstance) {
            return;
        }
        console.log("Sharing Thing to", targetUserProfile);
        const share = await shareObject(axiosInstance, {
            objectId: thing.id,
            targetUserId: targetUserProfile.id
        });
        console.log("Share result", share);
        navigate(`/things/${thingId}`)
    };

    return (
        <ShareEditor type={"thing"} thing={thing} profiles={profiles} userProfile={profile} onSubmit={onShare} onMutate={() => setMutateKey(mutateKey+1)}/>
    )
}