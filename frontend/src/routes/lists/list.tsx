import { useCallback, useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { getLists } from '../../api/lists';
import { PagedLists, ThingsSummary } from '../../api/resources';
import { Pages } from '../../components/pages';
import { PrimaryButton, Select } from '../../components/shared';
import { ListInfo } from '../../components/list_info';
import { OrderParam } from '../../api/common';
import { getThingsSummary } from '../../api/things';
import { UserNameAndUserId } from '../../components/shared/user';

export const Lists = () => {
  const axiosInstance = useContext(AxiosContext);
  const [lists, setLists] = useState<PagedLists | undefined>(undefined);
  const [currentPage, setCurrentPage] = useState(0);
  // can be reused for this purpose as things are just part of lists so the
  // owner will be identical, however might show users without lists
  const [summary, setSummary] = useState<ThingsSummary | undefined>(undefined);
  const [selectedOwners, setSelectedOwners] = useState<string[] | undefined>(undefined);
  const [order, setOrder] = useState<OrderParam>(OrderParam.AccessReasonDescending);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getThingsSummary(axiosInstance).then(setSummary);
  }, [axiosInstance]);

  useEffect(() => {
    if (summary === undefined) {
      setSelectedOwners(undefined);
    } else {
      setSelectedOwners(summary.ownerIds);
    }
  }, [summary]);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    if (selectedOwners === undefined) {
      return;
    }

    getLists(axiosInstance, currentPage, 20, selectedOwners, true, [order])
      .then(setLists)
      .catch((reason) => {
        console.error(reason);
      });
  }, [axiosInstance, currentPage, order, selectedOwners]);

  const toggleOwnerId = useCallback(
    (id: string) => {
      if (!summary) {
        return;
      }
      if (!selectedOwners) {
        return;
      }
      if (!selectedOwners.includes(id)) {
        setSelectedOwners([...selectedOwners, id]);
      } else if (selectedOwners.length === summary.ownerIds.length) {
        setSelectedOwners([id]);
      } else {
        const temp = [...selectedOwners].filter((v) => v !== id);
        if (temp.length > 0) {
          setSelectedOwners(temp);
        } else {
          setSelectedOwners(summary.ownerIds);
        }
      }
    },
    [selectedOwners, summary]
  );

  if (!lists) {
    return <p>Loading...</p>;
  }

  return (
    <>
      <div className="flex flex-row justify-between">
        <div className="flex flex-row border-primary gap-1 md:gap-2 select-none flex-wrap">
          {summary &&
            summary.ownerIds.sort().map((ownerId) => (
              <div
                className={
                  (selectedOwners && selectedOwners.includes(ownerId) ? '' : 'brightness-30 ') +
                  'bg-secondary p-1 rounded flex-none'
                }
                onClick={() => toggleOwnerId(ownerId)}
                key={ownerId}
              >
                <UserNameAndUserId
                  key={ownerId}
                  userId={ownerId}
                  textColor="text-primary"
                  imageBorderColor="border-display"
                />
              </div>
            ))}
        </div>
        <div className="flex flex-row gap-2 items-center flex-none">
          <Select value={order} onChange={(e) => setOrder(e.target.value as OrderParam)}>
            <option value={OrderParam.AccessReasonDescending}>Access Reason ↓</option>
            <option value={OrderParam.AccessReasonAscending}>Access Reason ↑</option>
            <option value={OrderParam.CreatedAtDescending}>Newest</option>
            <option value={OrderParam.CreatedAtAscending}>Oldest</option>
          </Select>
          <a href="/things/create">
            <PrimaryButton>Add Thing</PrimaryButton>
          </a>
        </div>
      </div>
      {lists.totalCount === 0 ? <p className="mt-3 text-display">No lists yet</p> : null}
      <div className="flex flex-row gap-4 mt-4 flex-wrap justify-center">
        {lists.lists.map((list) => (
          <ListInfo list={list} key={list.id} />
        ))}
      </div>
      {lists.lists.length > 0 && (
        <Pages
          currentPage={currentPage}
          onPageChange={(n) => setCurrentPage(n)}
          pages={lists.totalPageCount}
        />
      )}
    </>
  );
};
