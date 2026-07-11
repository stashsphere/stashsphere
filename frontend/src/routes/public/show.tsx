import { useContext, useEffect, useState } from 'react';
import { useParams } from 'react-router';
import { PublicShare } from '../../api/resources';
import { getPublicShare } from '../../api/public_share';
import { AxiosContext } from '../../context/axios';
import { PublicShareTokenContext } from '../../context/public_share';
import { PublicThingDetails } from '../../components/public/public_thing_details';
import { PublicListDetails } from '../../components/public/public_list_details';

export const ShowPublicShare = () => {
  const { token } = useParams();
  const [share, setShare] = useState<PublicShare | null>(null);
  const [notFound, setNotFound] = useState(false);
  const axiosInstance = useContext(AxiosContext);

  useEffect(() => {
    if (!axiosInstance) {
      return;
    }
    if (!token) {
      return;
    }
    getPublicShare(axiosInstance, token)
      .then(setShare)
      .catch(() => {
        setNotFound(true);
      });
  }, [axiosInstance, token]);

  if (token === undefined || notFound) {
    return <p className="text-display">This link is invalid or has been revoked.</p>;
  }
  if (share === null) {
    return <h1>Loading</h1>;
  }
  return (
    <PublicShareTokenContext.Provider value={token}>
      {share.type === 'thing' ? (
        <PublicThingDetails thing={share.thing} />
      ) : (
        <PublicListDetails list={share.list} />
      )}
    </PublicShareTokenContext.Provider>
  );
};
