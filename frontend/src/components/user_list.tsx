import { UserProfile } from '../api/resources';
import { UserNameAndProfile } from './shared/user';

type UserListProps = {
  users: UserProfile[];
  onClick?: (userId: string) => void;
  hintText?: string;
};

type UserListItemProps = {
  user: UserProfile;
  hintText?: string;
};

const UserListItem = ({ user, hintText }: UserListItemProps) => {
  return (
    <div className="flex items-center gap-2 p-3 -mt-2 text-sm text-gray-600 transition-colors duration-300 transform dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 dark:hover:text-white">
      <UserNameAndProfile
        profile={user}
        textColor="text-display"
        imageBorderColor="border-display"
      />
      {hintText}
    </div>
  );
};

export const UserList = (props: UserListProps) => {
  return (
    <ul className="border border-secondary">
      {props.users.map((user, index) => (
        <li onClick={() => props.onClick && props.onClick(user.id)} key={user.id}>
          <UserListItem key={index} user={user} hintText={props.hintText} />
        </li>
      ))}
    </ul>
  );
};
