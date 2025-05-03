import { FormEvent, useContext, useMemo, useState } from 'react';
import { List, Profile, Share, Thing, User } from '../api/resources';
import ThingInfo from './thing_info';
import { ListInfo } from './list_info';
import { UserList } from './user_list';
import { Icon } from './icon';
import { DangerButton, PrimaryButton, SecondaryButton } from './button';
import { AxiosContext } from '../context/axios';
import { deleteShare } from '../api/share';

type ShareDeleterProps = {
  share: Share;
  onDelete: () => void;
};

const ShareDeleter = ({ share, onDelete }: ShareDeleterProps) => {
  const axiosInstance = useContext(AxiosContext);
  const [wantDelete, setWantDelete] = useState(false);

  const onDeleteClick = () => {
    if (axiosInstance === null) {
      return;
    }
    deleteShare(axiosInstance, share.id).then(() => {
      onDelete();
    });
  };

  if (!wantDelete) {
    return (
      <div className="flex flex-row gap-4 my-2 justify-between">
        <div className="text-display">{share.targetUser.name}</div>
        <SecondaryButton className="py-0 px-1" onClick={() => setWantDelete(true)}>
          <Icon icon="mdi--trash" />
        </SecondaryButton>
      </div>
    );
  } else {
    return (
      <div className="flex flex-row gap-4 my-2 justify-between">
        <div className="text-display">Unshare for {share.targetUser.name}</div>
        <DangerButton className="py-0 px-1" onClick={() => onDeleteClick()}>
          Yes
        </DangerButton>
        <SecondaryButton className="py-0 px-1" onClick={() => setWantDelete(false)}>
          No
        </SecondaryButton>
      </div>
    );
  }
};

type ShareEditorProps = {
  users: User[];
  // the profile of the currently logged in user
  userProfile: Profile;
  onSubmit(targetUser: User): void;
  onMutate(): void;
} & (
  | {
      type: 'thing';
      thing: Thing;
    }
  | {
      type: 'list';
      list: List;
    }
);

export const ShareEditor = (props: ShareEditorProps) => {
  const [targetUser, setTargetUser] = useState<User | null>(null);
  const [searchTerm, setSearchTerm] = useState('');

  const searchableUsers = useMemo(() => {
    return props.users.filter((user) => user.id !== props.userProfile.id);
  }, [props.users, props.userProfile]);

  const selectableUsers = useMemo(() => {
    if (searchTerm === '') return [];

    return searchableUsers.filter((user) =>
      user.name.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [searchableUsers, searchTerm]);

  const ObjectComponent = (() => {
    switch (props.type) {
      case 'thing': {
        return <ThingInfo thing={props.thing}></ThingInfo>;
      }
      case 'list': {
        return <ListInfo list={props.list}></ListInfo>;
      }
    }
  })();
  const onSubmit = (event: FormEvent) => {
    event.preventDefault();
    if (targetUser === null) return;
    props.onSubmit(targetUser);
  };

  const existingShares = (() => {
    switch (props.type) {
      case 'thing':
        return props.thing.shares;
      case 'list':
        return props.list.shares;
    }
  })();

  return (
    <div>
      <div className="grid grid-cols-2">
        <div className="p-2">
          <h2 className="text-xl text-accent">
            Share {props.type === 'thing' ? 'Thing' : 'List'} to a Friend
          </h2>
          {ObjectComponent}
        </div>
        <div className="p-2">
          <h2 className="text-xl text-accent">Shared to</h2>
          <ul>
            {existingShares.map((x) => (
              <ShareDeleter key={x.id} share={x} onDelete={() => props.onMutate()} />
            ))}
          </ul>
        </div>
      </div>

      {targetUser === null ? (
        <>
          <div className="relative flex items-center my-2">
            <span className="absolute ml-2 w-8 h-8">
              <Icon icon="mdi--search" height={'100%'} />
            </span>

            <input
              onChange={(e) => setSearchTerm(e.target.value)}
              value={searchTerm}
              type="text"
              placeholder="Search for names"
              className="block w-full py-2.5 text-gray-700 placeholder-gray-400/70 bg-white border border-gray-200 rounded-lg pl-11 pr-5 rtl:pr-11 rtl:pl-5 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-600 focus:border-blue-400 dark:focus:border-blue-300 focus:ring-blue-300 focus:outline-hidden focus:ring-3 focus:ring-opacity-40"
            ></input>
          </div>
          <UserList users={selectableUsers} hintText="Share to this user" onClick={setTargetUser} />
        </>
      ) : (
        <>
          <div className="flex items-center p-3 mt-2 text-sm text-gray-600 dark:text-gray-300 border border-green-500 mb-2">
            <div className="w-16 h-16">
              <Icon icon="mdi--image-off-outline" height={'100%'} width={'75%'} />
            </div>
            <div className="mx-1">
              <h1 className="text-sm font-semibold text-gray-700 dark:text-gray-200">
                {targetUser.name}
              </h1>
            </div>
            <div className="grow"></div>
            <div className="w-16 h-16" onClick={() => setTargetUser(null)}>
              <Icon icon="mdi--close" height={'100%'} width={'75%'} />
            </div>
          </div>
          <form onSubmit={onSubmit}>
            <PrimaryButton type="submit">Share to {targetUser.name}</PrimaryButton>
          </form>
        </>
      )}
    </div>
  );
};
