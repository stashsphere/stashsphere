import { useContext, useEffect, useState } from 'react';
import { Thing } from '../../api/resources';
import { getThing } from '../../api/things';
import { AxiosContext } from '../../context/axios';
import { ThingImages } from './thing_images';
import { PropertyList } from './property_list';
import { Headline } from '../shared';
import { Icon } from '../shared';
import { SharingStateComponent } from '../shared/sharing_state';
import { UserNameAndUserId } from '../shared/user';

interface ThingDetailsProps {
  id: string;
}

const ThingActions = ({ thing }: { thing: Thing }) => {
  return (
    <div className="flex flex-row gap-2 justify-between items-center">
      <div className="flex flex-row justify-start gap-2">
        {thing.sharingState !== null && (
          <div className="flex flex-row text-display">
            {thing.sharingState != 'private' && 'Visible to'}
            <SharingStateComponent state={thing.sharingState} />
          </div>
        )}
        {thing.actions.canShare && (
          <>
            <div className="text-display">
              Directly shared with
              <span className="rounded-sm bg-secondary-200 text-onprimary mx-1 px-1">
                {thing.shares.length}
              </span>
              {thing.shares.length === 1 ? 'user' : 'users'}
            </div>
          </>
        )}
      </div>
      <div className="flex flex-row gap-2 justify-end">
        {thing.actions.canShare && (
          <a href={`/things/${thing.id}/share`}>
            <Icon
              icon="mdi--offer"
              size="medium"
              className="text-onneutral"
              tooltip="Share this thing"
            />
          </a>
        )}
        {thing.actions.canEdit && (
          <a href={`/things/${thing.id}/edit`}>
            <Icon
              icon="mdi--pencil"
              size="medium"
              className="text-onneutral"
              tooltip="Edit this thing"
            />
          </a>
        )}
        {thing.actions.canDelete && (
          <a href="#">
            <Icon
              icon="mdi--trash-can"
              size="medium"
              className="text-danger"
              tooltip="Delete this thing"
            />
          </a>
        )}
      </div>
    </div>
  );
};

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
      <div className="flex flex-col gap-8">
        <Headline type="h1">{thing.name}</Headline>
        <div className="flex flex-col md:flex-row gap-6">
          <div className="flex-1">
            <ThingImages images={thing.images} />
          </div>
          <div className="flex flex-col flex-1 gap-6">
            <ThingActions thing={thing} />
            <div>
              <Headline type="h2">Owner</Headline>

              <a href={`/users/${thing.owner.id}`}>
                <UserNameAndUserId
                  userId={thing.owner.id}
                  textColor="text-display"
                  imageBorderColor="border-display"
                />
              </a>
            </div>
            <div>
              <Headline type="h2">Quantity</Headline>
              <p className="text-display text-l">
                {thing.quantity} {thing.quantityUnit}
              </p>
            </div>
            <PropertyList properties={thing.properties} keyWidth="14rem" />
            <div>
              <Headline type="h2">Description</Headline>
              <div className="text-display">{thing.description}</div>
            </div>
            <div>
              <Headline type="h2">Lists</Headline>
              {thing.lists.length === 0 ? <p className="text-display">Not in any lists</p> : null}
              {thing.lists.map((list) => (
                <ul key={list.id}>
                  <li className="text-display">
                    <a href={`/lists/${list.id}`}>{list.name}</a>
                  </li>
                </ul>
              ))}
            </div>
            {thing.privateNote !== null && (
              <div>
                <Headline type="h2">Private Note</Headline>
                {thing.privateNote.length > 0 && (
                  <div className="bg-warning text-display rounded-sm p-2">{thing.privateNote}</div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }
};
