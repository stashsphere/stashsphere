import { useNavigate, useParams } from 'react-router';
import { ShareEditor } from '../../components/share_editor';
import { useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { getList } from '../../api/lists';
import { Profile, List, User } from '../../api/resources';
import { AuthContext } from '../../context/auth';
import { shareObject } from '../../api/share';
import { getAllUsers } from '../../api/user';

export const ShareList = () => {
  const { listId } = useParams();
  const navigate = useNavigate();

  const [list, setList] = useState<null | List>(null);
  const [mutateKey, setMutateKey] = useState(0);
  const axiosInstance = useContext(AxiosContext);
  const authContext = useContext(AuthContext);
  const profile = authContext.profile;

  const [users, setUsers] = useState<User[]>([]);

  useEffect(() => {
    if (!axiosInstance || listId === undefined) {
      return;
    }
    getList(axiosInstance, listId).then(setList);
  }, [axiosInstance, listId, mutateKey]);

  useEffect(() => {
    if (!axiosInstance) {
      return;
    }
    getAllUsers(axiosInstance).then(setUsers);
  }, [axiosInstance]);

  if (list === null || profile === null) {
    return <h1>Loading</h1>;
  }

  const onShare = async (targetUser: Profile) => {
    if (!axiosInstance) {
      return;
    }
    console.log('Sharing List to', targetUser);
    const share = await shareObject(axiosInstance, {
      objectId: list.id,
      targetUserId: targetUser.id,
    });
    console.log('Share result', share);
    navigate(`/lists/${listId}`);
  };

  return (
    <ShareEditor
      type={'list'}
      list={list}
      users={users}
      userProfile={profile}
      onSubmit={onShare}
      onMutate={() => setMutateKey(mutateKey + 1)}
    />
  );
};
