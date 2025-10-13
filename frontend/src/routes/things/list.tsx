import { useCallback, useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { PagedThings, ThingsSummary } from '../../api/resources';
import { getThings, getThingsSummary } from '../../api/things';
import { Pages } from '../../components/pages';
import { ThingInfo } from '../../components/shared';
import { PrimaryButton } from '../../components/shared';
import { UserNameAndUserId } from '../../components/shared/user';

export const Things = () => {
  const axiosInstance = useContext(AxiosContext);
  const [things, setThings] = useState<PagedThings | undefined>(undefined);
  const [summary, setSummary] = useState<ThingsSummary | undefined>(undefined);
  const [selectedOwners, setSelectedOwners] = useState<string[] | undefined>(undefined);

  const [currentPage, setCurrentPage] = useState(0);

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
    getThings(axiosInstance, currentPage, 10, selectedOwners)
      .then(setThings)
      .catch((reason) => {
        console.log(reason);
      });
  }, [axiosInstance, currentPage, selectedOwners]);

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

  if (!things) {
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
        <div className="flex-none">
          <a href="/things/create">
            <PrimaryButton>Add Thing</PrimaryButton>
          </a>
        </div>
      </div>
      <div className="flex flex-row gap-4 mt-4 flex-wrap justify-center">
        {things.things.map((thing) => (
          <ThingInfo thing={thing} key={thing.id} />
        ))}
      </div>
      {things.things.length > 0 && (
        <Pages
          currentPage={currentPage}
          onPageChange={(n) => setCurrentPage(n)}
          pages={things.totalPageCount}
        />
      )}
    </>
  );
};
