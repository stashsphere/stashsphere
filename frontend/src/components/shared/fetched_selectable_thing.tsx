import { useContext, useEffect, useState } from 'react';
import { Thing } from '../../api/resources';
import { getThing } from '../../api/things';
import { AxiosContext } from '../../context/axios';
import { SelectableThing } from './selectable_thing';

type FetchedSelectableThingProps = {
  thingId: string;
  selected: boolean;
  onSelect: (thingId: string, selected: boolean) => void;
};

export const FetchedSelectableThing = ({
  thingId,
  selected,
  onSelect,
}: FetchedSelectableThingProps) => {
  const axiosInstance = useContext(AxiosContext);
  const [thing, setThing] = useState<Thing | null>(null);

  useEffect(() => {
    if (!axiosInstance) return;
    getThing(axiosInstance, thingId).then(setThing);
  }, [axiosInstance, thingId]);

  if (!thing) {
    return null;
  }

  return <SelectableThing thing={thing} selected={selected} onSelect={onSelect} />;
};
