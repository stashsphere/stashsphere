import { useContext, useEffect, useMemo, useState } from 'react';
import { UserProfile } from '../../api/resources';
import { Icon } from './icon';
import { ImageComponent } from './image';
import { getUser } from '../../api/user';
import { AxiosContext } from '../../context/axios';

export type UserNameAndProfileProps = {
  imageBorderColor?: string;
  textColor?: string;
};

export const UserNameAndProfile = ({
  profile,
  imageBorderColor,
  textColor,
}: { profile: UserProfile } & UserNameAndProfileProps) => {
  const profilePicture = useMemo(() => {
    if (profile.image === null) {
      return <Icon icon="mdi--user" className="mr-2" />;
    } else {
      return (
        <ImageComponent
          image={profile.image}
          className={`w-10 h-10 rounded-full border-2 ${imageBorderColor}`}
          defaultWidth={128}
        />
      );
    }
  }, [profile, imageBorderColor]);

  return (
    <div className={`flex flex-row items-center gap-2 ${textColor}`}>
      {profilePicture}
      {profile.name}
    </div>
  );
};

export const UserNameAndUserId = ({
  userId,
  ...rest
}: { userId: string } & UserNameAndProfileProps) => {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const axiosInstance = useContext(AxiosContext);

  useEffect(() => {
    if (!axiosInstance) {
      return;
    }
    getUser(axiosInstance, userId).then((p) => setProfile(p));
  }, [userId, axiosInstance]);

  if (!profile) {
    return <div>Loading</div>;
  }
  return <UserNameAndProfile profile={profile} {...rest} />;
};
