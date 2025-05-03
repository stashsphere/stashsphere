import { Profile } from '../api/resources';
import { PrimaryButton } from './button';
import { Labeled } from './labeled';

type ProfileProps = {
  profile: Profile;
};

export const ProfileDetails = ({ profile }: ProfileProps) => {
  return (
    <div>
      <div className="flex flex-row justify-between">
        <h1 className="text-2xl text-secondary">My Profile</h1>
        <div className="flex flex-row items-end justify-end">
          <a href={'/user/profile/edit'}>
            <PrimaryButton>Edit</PrimaryButton>
          </a>
        </div>
      </div>
      <div className="flex flex-col gap-2 mt-2 text-display">
        <Labeled label="Name">{profile.name}</Labeled>
        <Labeled label="E-Mail">{profile.email}</Labeled>
        <Labeled label="ID">{profile.id}</Labeled>
      </div>
    </div>
  );
};
