import { useContext, useEffect, useState } from 'react';
import { List } from '../../api/resources';
import { AxiosContext } from '../../context/axios';
import { getList } from '../../api/lists';
import { Headline, Icon } from '../shared';
import { ThingInfo } from '../shared';
import { SharingStateComponent } from '../shared/sharing_state';

interface ListDetailsProps {
  id: string;
}

const ListActions = ({ list }: { list: List }) => {
  return (
    <div className="flex flex-row gap-3 justify-between items-center">
      <div className="flex flex-row justify-start gap-3">
        {list.sharingState !== null && (
          <div className="flex flex-row text-display">
            {list.sharingState != 'private' && 'Visible to'}
            <SharingStateComponent state={list.sharingState} />
          </div>
        )}
        {list.actions.canShare && (
          <>
            <div className="text-display">
              Directly shared with
              <span className="rounded-sm bg-secondary-201 text-onprimary mx-1 px-1">
                {list.shares.length}
              </span>
            </div>
            <a href={`/lists/${list.id}/share`}>
              <Icon
                icon="mdi--offer"
                size="medium"
                className="text-onneutral"
                tooltip="Share this thing"
              />
            </a>
          </>
        )}
      </div>
      <div className="flex flex-row gap-3 justify-end">
        {list.actions.canEdit && (
          <a href={`/lists/${list.id}/edit`}>
            <Icon
              icon="mdi--pencil"
              size="medium"
              className="text-onneutral"
              tooltip="Edit this thing"
            />
          </a>
        )}
        {list.actions.canDelete && (
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

export const ListDetails = (props: ListDetailsProps) => {
  const [list, setList] = useState<null | List>(null);
  const axiosInstance = useContext(AxiosContext);

  useEffect(() => {
    if (!axiosInstance) {
      return;
    }
    getList(axiosInstance, props.id).then(setList);
  }, [axiosInstance, props.id]);

  if (list === null) {
    return <h1>Loading</h1>;
  } else {
    return (
      <>
        <div className="flex flex-row justify-between mb-4">
          <h1 className="text-2xl text-accent">{list.name}</h1>
        </div>
        <ListActions list={list} />
        <div>
          <Headline type="h2">Owner</Headline>
          <p className="text-display text-l">
            <Icon icon="mdi--user" />
            {list.owner.name}
          </p>
        </div>
        <div className="flex flex-row gap-4 mt-4 flex-wrap justify-center">
          {list.things.map((thing) => (
            <ThingInfo thing={thing} key={thing.id} />
          ))}
        </div>
      </>
    );
  }
};
