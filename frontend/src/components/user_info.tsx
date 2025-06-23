import { User } from '../api/resources';

export const UserInfo = ({ user }: { user: User }) => {
  return (
    <div>
      <h1 className="text-2xl text-accent">{user.name}</h1>
    </div>
  );
};
