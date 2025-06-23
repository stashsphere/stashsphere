import { useParams } from 'react-router';
import { UserInfo } from '../../components/user_info';
import { useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { getUser } from '../../api/user';
import { User } from '../../api/resources';

export const ShowUser = () => {
  const { userId } = useParams();

  const [user, setUser] = useState<null | User>(null);
  const axiosInstance = useContext(AxiosContext);

  useEffect(() => {
    if (!axiosInstance || !userId) {
      return;
    }
    getUser(axiosInstance, userId).then(setUser);
  }, [axiosInstance, userId]);

  if (user === null) {
    return <p>Loading...</p>;
  } else {
    return <UserInfo user={user} />;
  }
};
