import { useState } from 'react';
import { List } from '../../api/resources';
import { Headline, Icon, PrimaryButton, SecondaryButton } from '../shared';
import { ThingInfo } from '../shared';
import { SharingStateComponent } from '../shared/sharing_state';
import { UserNameAndUserId } from '../shared/user';

interface ListDetailsProps {
  list: List;
  onDelete: () => void;
}

const ListActions = ({ list, onDeleteClick }: { list: List; onDeleteClick: () => void }) => {
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
              <span className="rounded-sm bg-secondary-200 text-onprimary mx-1 px-1">
                {list.shares.length}
              </span>
              {list.shares.length === 1 ? 'user' : 'users'}
            </div>
          </>
        )}
      </div>
      <div className="flex flex-row gap-3 justify-end">
        {list.actions.canShare && (
          <a href={`/lists/${list.id}/share`}>
            <Icon
              icon="mdi--share-variant"
              size="medium"
              className="text-onneutral"
              tooltip="Share this thing"
            />
          </a>
        )}
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
          <div onClick={() => onDeleteClick()}>
            <Icon
              icon="mdi--trash-can"
              size="medium"
              className="text-danger"
              tooltip="Delete this List"
            />
          </div>
        )}
      </div>
    </div>
  );
};

export const ListDetails = ({ list, onDelete }: ListDetailsProps) => {
  const [showDeleteDialog, setDeleteDialog] = useState(false);

  return (
    <>
      <div className="flex flex-row justify-between mb-4">
        <h1 className="text-2xl text-accent">{list.name}</h1>
      </div>
      <ListActions list={list} onDeleteClick={() => setDeleteDialog(!showDeleteDialog)} />
      {showDeleteDialog && (
        <div>
          <Headline type="h2">Delete List?</Headline>
          <div className="grid grid-cols-2 gap-2 max-w-sm">
            <PrimaryButton onClick={onDelete}>Delete</PrimaryButton>
            <SecondaryButton onClick={() => setDeleteDialog(false)}>Cancel</SecondaryButton>
          </div>
        </div>
      )}
      <div>
        <Headline type="h2">Owner</Headline>
        <UserNameAndUserId
          userId={list.owner.id}
          imageBorderColor="border-display"
          textColor="text-display"
        />
      </div>
      <div className="flex flex-row gap-4 mt-4 flex-wrap justify-center">
        {list.things.map((thing) => (
          <ThingInfo thing={thing} key={thing.id} />
        ))}
      </div>
    </>
  );
};
