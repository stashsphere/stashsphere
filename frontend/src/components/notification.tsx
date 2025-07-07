import { useContext, useEffect, useMemo, useState } from 'react';
import {
  FriendRequestNotification,
  StashsphereNotification,
  ListSharedNotification,
  UnknownNotification,
  ThingSharedNotification,
} from '../api/resources';
import { AxiosContext } from '../context/axios';
import { getUser } from '../api/user';
import { acknowledgeNotification } from '../api/notification';
import { PrimaryButton } from './shared';
import { Icon } from './shared';

const FriendRequestNotificationComponent = ({
  notification,
}: {
  notification: FriendRequestNotification;
}) => {
  const axiosInstance = useContext(AxiosContext);
  const [requester, setRequester] = useState('');

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getUser(axiosInstance, notification.content.senderId).then((v) => {
      setRequester(v.name);
    });
  }, [axiosInstance, notification]);

  const fontColor = notification.acknowledged ? 'text-display-light' : 'text-display';

  return <span className={fontColor}>{requester} wants to be your friend.</span>;
};

const ListSharedNotificationComponent = ({
  notification,
}: {
  notification: ListSharedNotification;
}) => {
  const axiosInstance = useContext(AxiosContext);
  const [sharer, setSharer] = useState('');

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getUser(axiosInstance, notification.content.sharerId).then((v) => {
      setSharer(v.name);
    });
  }, [axiosInstance, notification]);

  const fontColor = notification.acknowledged ? 'text-display-light' : 'text-display';

  return (
    <span className={fontColor}>
      {sharer} shared a{' '}
      <a className="text-accent" href={`/lists/${notification.content.listId}`}>
        list
      </a>{' '}
      with you.
    </span>
  );
};

const ThingSharedNotificationComponent = ({
  notification,
}: {
  notification: ThingSharedNotification;
}) => {
  const axiosInstance = useContext(AxiosContext);
  const [sharer, setSharer] = useState('');

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getUser(axiosInstance, notification.content.sharerId).then((v) => {
      setSharer(v.name);
    });
  }, [axiosInstance, notification]);

  const fontColor = notification.acknowledged ? 'text-display-light' : 'text-display';

  return (
    <span className={fontColor}>
      {sharer} shared a{' '}
      <a className="text-accent" href={`/things/${notification.content.thingId}`}>
        thing
      </a>{' '}
      with you.
    </span>
  );
};

const UnknownNotificationComponent = ({ notification }: { notification: UnknownNotification }) => {
  return (
    <span className="overflow-auto">Unknown notification: {JSON.stringify(notification)}</span>
  );
};

export const NotificationItem = ({
  notification,
  onAcknowledge,
}: {
  notification: StashsphereNotification;
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
    switch (notification.contentType) {
      case 'FRIEND_REQUEST':
        return <FriendRequestNotificationComponent notification={notification} />;
      case 'LIST_SHARED':
        return <ListSharedNotificationComponent notification={notification} />;
      case 'THING_SHARED':
        return <ThingSharedNotificationComponent notification={notification} />;
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
