import { useContext, useEffect, useMemo, useState } from "react";
import { Thing } from "../api/resources";
import { getThing } from "../api/things";
import { AxiosContext } from "../context/axios";
import { ImageGallery } from "./image_gallery";
import PropertyList from "./property_list";
import { DangerButton, SecondaryButton } from "./button";

interface ThingDetailsProps {
  id: string;
}

const ThingActions = ({ thing }: { thing: Thing }) => {
  const colClass = useMemo(() => {
    let actionCount = 0;

    if (thing.actions.canEdit) {
      actionCount++;
    }
    if (thing.actions.canShare) {
      actionCount++;
    }
    if (thing.actions.canDelete) {
      actionCount++;
    }
    return `grid-cols-${actionCount}`
  }, [thing])

  return <div className={`grid gap-x-2 ${colClass}`}>
    {thing.actions.canEdit &&
      <a href={`/things/${thing.id}/edit`}>
        <SecondaryButton className="w-full">Edit</SecondaryButton>
      </a>}
    {thing.actions.canShare &&
      <a href={`/things/${thing.id}/share`}>
        <SecondaryButton className="w-full">Share</SecondaryButton>
      </a>}
    {thing.actions.canDelete && <a href="#">
      <DangerButton className="w-full">Delete</DangerButton>
    </a>}
  </div>
}

export const ThingDetails = (props: ThingDetailsProps) => {
  const [thing, setThing] = useState<null | Thing>(null);
  const axiosInstance = useContext(AxiosContext);

  useEffect(() => {
    if (!axiosInstance) {
      return;
    }
    getThing(axiosInstance, props.id).then(setThing);
  }, [axiosInstance, props.id]);

  if (thing === null) {
    return <h1>Loading</h1>;
  } else {
    return (
      <>
        <div className="flex flex-row justify-between mb-4">
          <h1 className="text-2xl text-accent">{thing.name}</h1>
          <ThingActions thing={thing} />
        </div>
        <h2 className="text-secondary text-xl">Quantity</h2>
        <p className="text-display text-l">{thing.quantity} {thing.quantityUnit}</p>
        <h2 className="text-secondary text-xl">Description</h2>
        <div className="text-display">
          {thing.description}
        </div>
        {thing.privateNote !== null && <>
          <h2 className="text-secondary text-xl">Private Note</h2>
          {thing.privateNote.length > 0 &&
            <div className="bg-warning text-display rounded p-2">
              {thing.privateNote}
            </div>}</>}
        <PropertyList properties={thing.properties} keyWidth="14rem"/>
        <div className="mt-4">
          <ImageGallery images={thing.images} />
        </div>
        <div>
          <h2 className="text-secondary text-xl font-bold">Lists</h2>
          {thing.lists.length === 0 ? <p className="text-display">Not in any lists</p> : null}
          {thing.lists.map((list) => (
            <ul key={list.id}>
              <li className="text-display">
                <a href={`/lists/${list.id}`}>{list.name}</a>
              </li>
            </ul>
          ))}
        </div>
      </>
    );
  }
};
