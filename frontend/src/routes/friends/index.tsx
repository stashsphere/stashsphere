import { useContext, useEffect, useMemo, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import {
  deleteFriendShip,
  getFriendRequests,
  getFriendShips,
  reactToFriendRequest,
  sendFriendRequest,
} from '../../api/friend';
import { FriendRequest, FriendShip, User, UserProfile } from '../../api/resources';
import { AuthContext } from '../../context/auth';
import { UserList } from '../../components/user_list';
import { Headline, PrimaryButton, SecondaryButton } from '../../components/shared';
import { getAllUsers } from '../../api/user';
import { UserNameAndUserId } from '../../components/shared/user';

const FriendShipEntry = ({
  friendShip,
  onDelete,
}: {
  friendShip: FriendShip;
  onDelete: () => void;
}) => {
  const [unfriend, setUnfriend] = useState(false);
  const axiosInstance = useContext(AxiosContext);

  const onUnfriend = () => {
    if (!axiosInstance) {
      return;
    }
    deleteFriendShip(axiosInstance, friendShip.friend.id).then(onDelete);
  };

  return (
    <>
      <div className="flex flex-row gap-4 items-center justify-between">
        <a href={`/users/${friendShip.friend.id}`}>
          <UserNameAndUserId
            userId={friendShip.friend.id}
            imageBorderColor="border-display"
            textColor="text-display"
          />
        </a>
        {!unfriend && <SecondaryButton onClick={() => setUnfriend(true)}>Unfriend</SecondaryButton>}
        {unfriend && (
          <div className="grid grid-cols-2 gap-2 max-w-sm">
            <PrimaryButton onClick={onUnfriend}>Unfriend {friendShip.friend.name}</PrimaryButton>
            <SecondaryButton onClick={() => setUnfriend(false)}>Cancel</SecondaryButton>
          </div>
        )}
      </div>
    </>
  );
};

const SentFriendRequestEntry = ({ request }: { request: FriendRequest }) => {
  const state = useMemo(() => {
    switch (request.state) {
      case 'accepted':
        return <div className="text-green-500">Accepted</div>;
      case 'pending':
        return <div className="text-yellow-500">Pending</div>;
      case 'rejected':
        return <div className="text-red-800">Rejected</div>;
    }
  }, [request]);

  return (
    <>
      <div className="flex flex-row items-center gap-2 justify-between">
        <UserNameAndUserId
          userId={request.receiver.id}
          textColor="text-display"
          imageBorderColor="border-display"
        />
        {state}
      </div>
    </>
  );
};

const ReceivedFriendRequestEntry = ({
  request,
  onReacted,
}: {
  request: FriendRequest;
  onReacted: () => void;
}) => {
  const axiosInstance = useContext(AxiosContext);

  const reactFriendRequest = (accept: boolean) => {
    if (axiosInstance === null) {
      return;
    }
    reactToFriendRequest(axiosInstance, request.id, accept).then(() => {
      onReacted();
    });
  };

  return (
    <>
      <div className="flex flex-row gap-2">
        <UserNameAndUserId
          userId={request.sender.id}
          textColor="text-display"
          imageBorderColor="border-display"
        />
        <div className="grid grid-cols-2 gap-2 max-w-sm">
          <PrimaryButton onClick={() => reactFriendRequest(true)}>Accept</PrimaryButton>
          <SecondaryButton onClick={() => reactFriendRequest(false)}>Reject</SecondaryButton>
        </div>
      </div>
    </>
  );
};

const SendFriendRequestComponent = ({
  user,
  onSuccess,
  onCancel,
}: {
  user: User;
  onSuccess: () => void;
  onCancel: () => void;
}) => {
  const axiosInstance = useContext(AxiosContext);

  const onSend = () => {
    if (axiosInstance === null) {
      return;
    }
    sendFriendRequest(axiosInstance, user.id)
      .then(() => {
        onSuccess();
      })
      .catch((e) => {
        console.log(e);
        onCancel();
      });
  };

  return (
    <div className="flex flex-col">
      <span className="text-display">Send friend request to {user.name}?</span>
      <div className="grid grid-cols-2 gap-2 max-w-sm">
        <PrimaryButton onClick={onSend}>Ok</PrimaryButton>
        <SecondaryButton onClick={onCancel}>Cancel</SecondaryButton>
      </div>
    </div>
  );
};

export const ShowFriends = () => {
  const axiosInstance = useContext(AxiosContext);
  const authContext = useContext(AuthContext);
  const [friendShips, setFriendShips] = useState<FriendShip[]>([]);
  const [sentFriendRequests, setSentFriendRequests] = useState<FriendRequest[]>([]);
  const [receivedFriendRequests, setReceivedFriendRequests] = useState<FriendRequest[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [users, setUsers] = useState<UserProfile[]>([]);
  const [targetUser, setTargetUser] = useState<UserProfile | null>(null);
  const [mutateKey, setMutateKey] = useState(0);

  const userProfile = authContext.profile;

  useEffect(() => {
    if (!axiosInstance) {
      return;
    }
    getAllUsers(axiosInstance).then(setUsers);
  }, [axiosInstance]);

  const searchableUsers = useMemo(() => {
    if (userProfile === null) {
      return [];
    }
    return users.filter((user) => user.id !== userProfile.id);
  }, [users, userProfile]);

  const selectableUsers = useMemo(() => {
    if (searchTerm === '') return [];
    // TODO filter out existing friends
    return searchableUsers.filter((profile) =>
      profile.name.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [searchableUsers, searchTerm]);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getFriendRequests(axiosInstance).then((value) => {
      setReceivedFriendRequests(value.received);
      setSentFriendRequests(value.sent);
    });
  }, [axiosInstance, mutateKey]);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getFriendShips(axiosInstance).then((value) => {
      setFriendShips(value.friendShips);
    });
  }, [axiosInstance, mutateKey]);

  const updateState = () => {
    setMutateKey((prev) => prev + 1);
  };

  return (
    <>
      <div className="flex flex-col max-w-3xl">
        <div>
          <Headline type="h2">Friend Requests</Headline>
          <input
            onChange={(e) => setSearchTerm(e.target.value)}
            value={searchTerm}
            type="text"
            placeholder="Search for names and email addresses"
            className="mt-1 p-2 text-display focus:outline-none border border-gray-300 rounded-sm w-full"
          />
          {targetUser === null ? (
            <UserList
              users={selectableUsers}
              hintText="Send friend request"
              onClick={(id) => setTargetUser(selectableUsers.find((u) => u.id === id) || null)}
            />
          ) : (
            <SendFriendRequestComponent
              user={targetUser}
              onCancel={() => setTargetUser(null)}
              onSuccess={() => {
                setTargetUser(null);
                setSearchTerm('');
                updateState();
              }}
            />
          )}
          <div className="flex flex-col gap-2 mt-2">
            {receivedFriendRequests
              .filter((r) => r.state === 'pending')
              .map((r) => (
                <ReceivedFriendRequestEntry request={r} onReacted={updateState} key={r.id} />
              ))}
            {sentFriendRequests.map((r) => (
              <SentFriendRequestEntry request={r} key={r.id} />
            ))}
          </div>
        </div>
        <div className="mt-4">
          <Headline type="h2">Friends</Headline>
          <div className="flex flex-col gap-2">
            {friendShips.length > 0 &&
              friendShips.map((friendShip) => (
                <FriendShipEntry friendShip={friendShip} onDelete={updateState} />
              ))}
            {friendShips.length == 0 && (
              <p className="text-display text-sm">No friends added yet.</p>
            )}
          </div>
        </div>
      </div>
    </>
  );
};
