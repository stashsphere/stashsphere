import { useContext, useEffect, useMemo, useState } from 'react';
import { FriendRequestNotification, StashsphereNotification } from '../api/resources';
import { AxiosContext } from '../context/axios';
import { getUser } from '../api/user';
import { acknowledgeNotification } from '../api/notification';
import { PrimaryButton } from './button';
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

export const NotificationItem = ({ notification }: { notification: StashsphereNotification }) => {
  const axiosInstance = useContext(AxiosContext);

  const onAcknowledge = () => {
    if (axiosInstance === null) {
      return;
    }
    acknowledgeNotification(axiosInstance, notification.id);
  };

  const body = useMemo(() => {
    switch (notification.contentType) {
      case 'FRIEND_REQUEST':
        return <FriendRequestNotificationComponent notification={notification} />;
    }
    return null;
  }, [notification]);

  return (
    <div className="flex flex-row justify-between border border-primary m-2 p-2">
      {body}
      {!notification.acknowledged && (
        <PrimaryButton
          onClick={() => {
            onAcknowledge();
          }}
        >
          <Icon icon="mdi--check" />
        </PrimaryButton>
      )}
    </div>
  );
};
