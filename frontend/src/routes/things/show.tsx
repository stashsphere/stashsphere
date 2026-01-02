import { useNavigate, useParams } from 'react-router';
import { ThingDetails } from '../../components/thing_details';
import { useCallback, useContext, useEffect, useState } from 'react';
import { Thing } from '../../api/resources';
import { getThing, deleteThing } from '../../api/things';
import { AxiosContext } from '../../context/axios';

export const ShowThing = () => {
  const { thingId } = useParams();
  const [thing, setThing] = useState<null | Thing>(null);
  const axiosInstance = useContext(AxiosContext);
  const navigate = useNavigate();

  useEffect(() => {
    if (!axiosInstance) {
      return;
    }
    if (!thingId) {
      return;
    }
    getThing(axiosInstance, thingId).then(setThing);
  }, [axiosInstance, thingId]);

  const onDelete = useCallback(async () => {
    if (thing === null) {
      return;
    }
    if (axiosInstance === null) {
      return;
    }
    await deleteThing(axiosInstance, thing.id);
    navigate(`/things`);
  }, [axiosInstance, navigate, thing]);

  if (thingId === undefined) {
    return <p>invalid id</p>;
  } else if (thing === null) {
    return <h1>Loading</h1>;
  } else {
    return <ThingDetails thing={thing} onDelete={onDelete} />;
  }
};
