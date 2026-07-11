import { useContext, useEffect, useState } from 'react';
import { Thing } from '../api/resources';
import { getThing } from '../api/things';
import { AxiosContext } from '../context/axios';
import { ThingInfo } from './shared';

type FetchedThingInfoProps = {
  thingId: string;
  hideCart?: boolean;
};

export const FetchedThingInfo = ({ thingId, hideCart }: FetchedThingInfoProps) => {
  const axiosInstance = useContext(AxiosContext);
  const [thing, setThing] = useState<Thing | null>(null);

  useEffect(() => {
    if (!axiosInstance) return;
    getThing(axiosInstance, thingId).then(setThing);
  }, [axiosInstance, thingId]);

  if (!thing) {
    return null;
  }

  return <ThingInfo thing={thing} hideCart={hideCart} />;
};
