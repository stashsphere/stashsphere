import { useNavigate, useParams } from 'react-router';
import { ShareEditor } from '../../components/share_editor';
import { useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { getThing } from '../../api/things';
import { Thing, User } from '../../api/resources';
import { AuthContext } from '../../context/auth';
import { shareObject } from '../../api/share';
import { getAllUsers } from '../../api/user';

export const ShareThing = () => {
  const { thingId } = useParams();
  const navigate = useNavigate();

  const [thing, setThing] = useState<null | Thing>(null);
  const [mutateKey, setMutateKey] = useState(0);
  const axiosInstance = useContext(AxiosContext);
  const authContext = useContext(AuthContext);
  const profile = authContext.profile;

  const [users, setUsers] = useState<User[]>([]);

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
    getAllUsers(axiosInstance).then(setUsers);
  }, [axiosInstance]);

  if (thing === null || profile === null) {
    return <h1>Loading</h1>;
  }

  const onShare = async (targetUser: User) => {
    if (!axiosInstance) {
      return;
    }
    console.log('Sharing Thing to', targetUser);
    const share = await shareObject(axiosInstance, {
      objectId: thing.id,
      targetUserId: targetUser.id,
    });
    console.log('Share result', share);
    navigate(`/things/${thingId}`);
  };

  return (
    <ShareEditor
      type={'thing'}
      thing={thing}
      users={users}
      userProfile={profile}
      onSubmit={onShare}
      onMutate={() => setMutateKey(mutateKey + 1)}
    />
  );
};
