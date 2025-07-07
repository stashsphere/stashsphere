import { UserProfile } from '../api/resources';
import { Headline, Icon, ImageComponent } from './shared';

export const UserInfo = ({ profile }: { profile: UserProfile }) => {
  return (
    <div>
      {profile.image ? (
        <ImageComponent
          image={profile.image}
          defaultWidth={256}
          className="w-[256px] h-[256px] mb-4 rounded-sm object-contain"
          alt={`Profile picture of ${profile.name}`}
        />
      ) : (
        <Icon size="256px" icon={'mdi--user'}></Icon>
      )}

      <Headline type="h1">{profile.fullName}</Headline>
      <Headline type="h2">{profile.name}</Headline>

      <p className="text-display">{profile.information}</p>
    </div>
  );
};
