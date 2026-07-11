import { useCallback, useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { PublicShareIndexEntry } from '../../api/resources';
import { deletePublicShare, getPublicShares, urlForPublicShare } from '../../api/public_share';
import { Headline, Icon } from '../../components/shared';
import { DangerButton, SecondaryButton } from '../../components/shared';
import { FetchedThingInfo } from '../../components/fetched_thing_info';
import { FetchedListInfo } from '../../components/fetched_list_info';

type PublicShareEntryProps = {
  entry: PublicShareIndexEntry;
  onDelete: () => void;
};

const PublicShareEntry = ({ entry, onDelete }: PublicShareEntryProps) => {
  const axiosInstance = useContext(AxiosContext);
  const [wantDelete, setWantDelete] = useState(false);

  const url = urlForPublicShare(entry.id);

  const onDeleteClick = () => {
    if (axiosInstance === null) {
      return;
    }
    deletePublicShare(axiosInstance, entry.id).then(() => {
      onDelete();
    });
  };

  return (
    <div className="flex flex-col gap-2 border border-secondary rounded-md p-2">
      {entry.type === 'thing' ? (
        <FetchedThingInfo thingId={entry.objectId} hideCart />
      ) : (
        <FetchedListInfo listId={entry.objectId} />
      )}
      <div className="text-display text-sm">
        created {new Date(entry.createdAt).toLocaleString()}
      </div>
      <a href={url} className="text-display text-sm underline truncate w-80" title={url}>
        {url}
      </a>
      <div className="flex flex-row gap-2 items-center">
        <SecondaryButton
          className="py-0 px-1"
          onClick={() => navigator.clipboard.writeText(url)}
          title="Copy public link"
        >
          <Icon icon="mdi--content-copy" />
        </SecondaryButton>
        {!wantDelete ? (
          <SecondaryButton className="py-0 px-1" onClick={() => setWantDelete(true)}>
            <Icon icon="mdi--trash" />
          </SecondaryButton>
        ) : (
          <>
            <div className="text-display">Revoke this link?</div>
            <DangerButton className="py-0 px-1" onClick={() => onDeleteClick()}>
              Yes
            </DangerButton>
            <SecondaryButton className="py-0 px-1" onClick={() => setWantDelete(false)}>
              No
            </SecondaryButton>
          </>
        )}
      </div>
    </div>
  );
};

export const PublicShares = () => {
  const axiosInstance = useContext(AxiosContext);
  const [entries, setEntries] = useState<PublicShareIndexEntry[] | null>(null);
  const [mutateKey, setMutateKey] = useState(0);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getPublicShares(axiosInstance)
      .then(setEntries)
      .catch((reason) => {
        console.log(reason);
      });
  }, [axiosInstance, mutateKey]);

  const mutate = useCallback(() => {
    setMutateKey((prev) => prev + 1);
  }, []);

  if (entries === null) {
    return <p>Loading...</p>;
  }
  return (
    <>
      <Headline type="h1">Public Links</Headline>
      {entries.length === 0 ? (
        <p className="text-display">No public links yet</p>
      ) : (
        <div className="flex flex-row gap-4 mt-4 flex-wrap">
          {entries.map((entry) => (
            <PublicShareEntry key={entry.id} entry={entry} onDelete={() => mutate()} />
          ))}
        </div>
      )}
    </>
  );
};
