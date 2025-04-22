import { useParams } from 'react-router';
import { ThingDetails } from '../../components/thing_detail';

export const ShowThing = () => {
  const { thingId } = useParams();

  if (thingId === undefined) {
    return <p>invalid id</p>;
  } else {
    return <ThingDetails id={thingId} />;
  }
};
