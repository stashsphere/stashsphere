import { User } from '../api/resources';
import { Icon } from './shared';

type UserListProps = {
  users: User[];
  onClick?: (user: User) => void;
  hintText?: string;
};

type UserListItemProps = {
  user: User;
  hintText?: string;
};

const UserListItem = (props: UserListItemProps) => {
  return (
    <div className="flex items-center p-3 -mt-2 text-sm text-gray-600 transition-colors duration-300 transform dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 dark:hover:text-white">
      <div className="w-16 h-16">
        <Icon icon="mdi--image-off-outline" />
      </div>
      <div className="mx-1">
        <h1 className="text-sm font-semibold text-gray-700 dark:text-gray-200">
          {props.user.name}
        </h1>
      </div>
      {props.hintText}
    </div>
  );
};

export const UserList = (props: UserListProps) => {
  return (
    <ul className="border border-gray-300">
      {props.users.map((user, index) => (
        <li onClick={() => props.onClick && props.onClick(user)} key={user.id}>
          <UserListItem key={index} user={user} hintText={props.hintText} />
        </li>
      ))}
    </ul>
  );
};
