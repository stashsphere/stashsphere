import { ThingInfo } from './shared';
import { FormEvent, useContext, useEffect, useMemo, useState } from 'react';
import { List, Profile, Share, SharingState, Thing, User, UserProfile } from '../api/resources';
import { ListInfo } from './list_info';
import { UserList } from './user_list';
import { Icon } from './shared';
import { DangerButton, PrimaryButton, SecondaryButton } from './shared';
import { AxiosContext } from '../context/axios';
import { deleteShare } from '../api/share';
import { UserNameAndProfile } from './shared/user';

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
  users: UserProfile[];
  // the profile of the currently logged in user
  userProfile: Profile;
  onSubmit(targetUser: User): void;
  onChangeSharingState(newState: SharingState): void;
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
  const [targetUser, setTargetUser] = useState<UserProfile | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [sharingState, setSharingState] = useState<SharingState>('private');
  const [initialSharingState, setInitialSharingState] = useState<SharingState>('private');

  const searchableUsers = useMemo(() => {
    return props.users.filter((user) => user.id !== props.userProfile.id);
  }, [props.users, props.userProfile]);

  const selectableUsers = useMemo(() => {
    if (searchTerm === '') return [];

    return searchableUsers.filter((user) =>
      user.name.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [searchableUsers, searchTerm]);

  const ObjectComponent = useMemo(() => {
    switch (props.type) {
      case 'thing': {
        return <ThingInfo thing={props.thing}></ThingInfo>;
      }
      case 'list': {
        return <ListInfo list={props.list}></ListInfo>;
      }
    }
  }, [props]);

  useEffect(() => {
    switch (props.type) {
      case 'thing':
        setSharingState(props.thing.sharingState || 'private');
        setInitialSharingState(props.thing.sharingState || 'private');
        break;
      case 'list':
        setSharingState(props.list.sharingState || 'private');
        setInitialSharingState(props.list.sharingState || 'private');
        break;
    }
  }, [props]);

  const onSubmit = (event: FormEvent) => {
    event.preventDefault();
    if (targetUser === null) return;
    props.onSubmit(targetUser);
  };

  const onUpdateSharingState = () => {
    props.onChangeSharingState(sharingState);
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
          <h2 className="text-xl text-accent">General Setting</h2>
          <select
            className="text-display"
            value={sharingState}
            onChange={(e) => setSharingState(e.target.value as SharingState)}
          >
            <option value="private">Private</option>
            <option value="friends">Friends</option>
            <option value="friends-of-friends">Friends of Friends</option>
          </select>
          <PrimaryButton
            disabled={sharingState === initialSharingState}
            onClick={onUpdateSharingState}
          >
            Update
          </PrimaryButton>
          <h2 className="text-xl text-accent">Individual Shares</h2>
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
              <Icon icon="mdi--search" />
            </span>

            <input
              onChange={(e) => setSearchTerm(e.target.value)}
              value={searchTerm}
              type="text"
              placeholder="Search for names"
              className="block w-full py-2.5 text-gray-700 placeholder-gray-400/70 bg-white border border-gray-200 rounded-lg pl-11 pr-5 rtl:pr-11 rtl:pl-5 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-600 focus:border-blue-400 dark:focus:border-blue-300 focus:ring-blue-300 focus:outline-hidden focus:ring-3 focus:ring-opacity-40"
            ></input>
          </div>
          <UserList
            users={selectableUsers}
            hintText="Share to this user"
            onClick={(userId) =>
              setTargetUser(selectableUsers.find((s) => s.id === userId) || null)
            }
          />
        </>
      ) : (
        <>
          <div className="flex items-center p-3 mt-2 text-sm text-gray-600 dark:text-gray-300 border border-green-500 mb-2">
            <UserNameAndProfile
              profile={targetUser}
              textColor="text-display"
              imageBorderColor="border-display"
            />
            <div className="grow"></div>
            <div className="w-16 h-16" onClick={() => setTargetUser(null)}>
              <Icon icon="mdi--close" />
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
