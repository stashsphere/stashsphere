import { useContext, useEffect, useMemo, useState } from 'react';
import {
  FriendRequestNotification,
  StashSphereNotification,
  ListSharedNotification,
  UnknownNotification,
  ThingSharedNotification,
  FriendRequest,
  FriendRequestReactionNotification,
  ThingsAddedToListNotification,
} from '../api/resources';
import { AxiosContext } from '../context/axios';
import { acknowledgeNotification } from '../api/notification';
import { PrimaryButton } from './shared';
import { Icon } from './shared';
import { UserNameAndUserId } from './shared/user';
import { getFriendRequests } from '../api/friend';
import { AuthContext } from '../context/auth';

const FriendRequestNotificationComponent = ({
  notification,
}: {
  notification: FriendRequestNotification;
}) => {
  const fontColor = notification.acknowledged ? 'text-display-light' : 'text-display';
  return (
    <a href="/friends">
      <div className="flex flex-row items-center gap-1">
        <UserNameAndUserId
          userId={notification.content.senderId}
          textColor={fontColor}
          imageBorderColor="border-display"
        />
        <div className={fontColor}>wants to be your friend.</div>
      </div>
    </a>
  );
};

const FriendRequestReactionNotificationComponent = ({
  notification,
}: {
  notification: FriendRequestReactionNotification;
}) => {
  const fontColor = notification.acknowledged ? 'text-display-light' : 'text-display';

  const authContext = useContext(AuthContext);
  const axiosInstance = useContext(AxiosContext);
  const [friendRequest, setFriendRequest] = useState<FriendRequest | undefined>(undefined);

  useEffect(() => {
    if (!axiosInstance) {
      return;
    }
    getFriendRequests(axiosInstance).then((v) =>
      setFriendRequest(
        v.received.concat(v.sent).find((r) => r.id === notification.content.requestId)
      )
    );
  }, [axiosInstance, notification]);

  if (!friendRequest || !authContext) {
    return <div>Loading</div>;
  }

  if (friendRequest.receiver.id === authContext.profile?.id) {
    return (
      <div className="flex flex-row items-center gap-1">
        <UserNameAndUserId
          userId={friendRequest.receiver.id}
          textColor={fontColor}
          imageBorderColor="border-display"
        />
        <div className={fontColor}>
          You have {notification.content.accepted ? 'accepted' : 'rejected'} the friend request.
        </div>
      </div>
    );
  } else {
    return (
      <div className="flex flex-row items-center gap-1">
        <UserNameAndUserId
          userId={friendRequest.receiver.id}
          textColor={fontColor}
          imageBorderColor="border-display"
        />
        <div className={fontColor}>
          has {notification.content.accepted ? 'accepted' : 'rejected'} your friend request.
        </div>
      </div>
    );
  }
};

const ListSharedNotificationComponent = ({
  notification,
}: {
  notification: ListSharedNotification;
}) => {
  const fontColor = notification.acknowledged ? 'text-display-light' : 'text-display';

  return (
    <div className="flex flex-row items-center gap-1">
      <UserNameAndUserId
        userId={notification.content.sharerId}
        textColor={fontColor}
        imageBorderColor="border-display"
      />
      <div className={fontColor}>
        shared a{' '}
        <a className="text-accent" href={`/lists/${notification.content.listId}`}>
          list
        </a>{' '}
        with you.
      </div>
    </div>
  );
};

const ThingSharedNotificationComponent = ({
  notification,
}: {
  notification: ThingSharedNotification;
}) => {
  const fontColor = notification.acknowledged ? 'text-display-light' : 'text-display';

  return (
    <div className="flex flex-row items-center gap-1">
      <UserNameAndUserId
        userId={notification.content.sharerId}
        textColor={fontColor}
        imageBorderColor="border-display"
      />
      <div className={fontColor}>
        shared a{' '}
        <a className="text-accent" href={`/things/${notification.content.thingId}`}>
          thing
        </a>{' '}
        with you.
      </div>
    </div>
  );
};

const ThingsAddedToListNotificationComponent = ({
  notification,
}: {
  notification: ThingsAddedToListNotification;
}) => {
  const fontColor = notification.acknowledged ? 'text-display-light' : 'text-display';

  return (
    <div className="flex flex-row items-center gap-1">
      <UserNameAndUserId
        userId={notification.content.addedById}
        textColor={fontColor}
        imageBorderColor="border-display"
      />
      <div className={fontColor}>
        added things to a{' '}
        <a className="text-accent" href={`/lists/${notification.content.listId}`}>
          list you have access to.
        </a>{' '}
      </div>
    </div>
  );
};

const UnknownNotificationComponent = ({ notification }: { notification: UnknownNotification }) => {
  return (
    <span className="overflow-auto text-display">
      Unknown notification: {JSON.stringify(notification)}
    </span>
  );
};

export const NotificationItem = ({
  notification,
  onAcknowledge,
}: {
  notification: StashSphereNotification;
  onAcknowledge: () => void;
}) => {
  const axiosInstance = useContext(AxiosContext);

  const acknowledge = () => {
    if (axiosInstance === null) {
      return;
    }
    acknowledgeNotification(axiosInstance, notification.id).then(() => {
      onAcknowledge();
    });
  };

  const body = useMemo(() => {
    console.log(notification);
    switch (notification.contentType) {
      case 'FRIEND_REQUEST':
        return <FriendRequestNotificationComponent notification={notification} />;
      case 'LIST_SHARED':
        return <ListSharedNotificationComponent notification={notification} />;
      case 'THING_SHARED':
        return <ThingSharedNotificationComponent notification={notification} />;
      case 'FRIEND_REQUEST_REACTION':
        return <FriendRequestReactionNotificationComponent notification={notification} />;
      case 'THINGS_ADDED_TO_LIST':
        return <ThingsAddedToListNotificationComponent notification={notification} />;
      default:
        return <UnknownNotificationComponent notification={notification} />;
    }
    return null;
  }, [notification]);

  return (
    <div className="flex flex-row justify-between border border-primary m-2 p-2">
      {body}
      {!notification.acknowledged && (
        <PrimaryButton
          className="w-14 h-14 flex-none"
          onClick={() => {
            acknowledge();
          }}
        >
          <Icon icon="mdi--check" />
        </PrimaryButton>
      )}
    </div>
  );
};
