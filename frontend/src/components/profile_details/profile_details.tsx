import { Profile } from '../../api/resources';
import { Icon, ImageComponent, PrimaryButton } from '../shared';
import { Labeled } from '../shared';

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
        <p className="text-accent">private</p>
        <Labeled label="ID">{profile.id}</Labeled>
        <Labeled label="E-Mail">{profile.email}</Labeled>
        <p className="text-accent">visible to other users</p>
        <div className="flex flex-row gap-4">
          <div>
            {profile.image ? (
              <div className="flex w-80 h-80 items-center justify-center bg-brand-900 p-2 rounded-md">
                <ImageComponent
                  image={profile.image}
                  defaultWidth={512}
                  className="w-full h-full amb-4 rounded-sm object-contain"
                  alt="Main"
                />
              </div>
            ) : (
              <div className="flex flex-col text-center">
                <Icon size="256px" icon="mdi--user"></Icon>
                <p className="text-display">No profile picture set</p>
              </div>
            )}
          </div>
          <div>
            <Labeled label="Name">{profile.name}</Labeled>
            <Labeled label="Full Name">{profile.fullName}</Labeled>
            <Labeled label="Information">{profile.information}</Labeled>
          </div>
        </div>
      </div>
    </div>
  );
};
