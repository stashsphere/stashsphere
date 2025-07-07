import { useParams } from 'react-router';
import { UserInfo } from '../../components/user_info';
import { useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { getUser } from '../../api/user';
import { UserProfile } from '../../api/resources';

export const ShowUser = () => {
  const { userId } = useParams();

  const [profile, setProfile] = useState<null | UserProfile>(null);
  const axiosInstance = useContext(AxiosContext);

  useEffect(() => {
    if (!axiosInstance || !userId) {
      return;
    }
    getUser(axiosInstance, userId).then(setProfile);
  }, [axiosInstance, userId]);

  if (profile === null) {
    return <p>Loading...</p>;
  } else {
    return <UserInfo profile={profile} />;
  }
};
