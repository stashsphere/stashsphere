import { useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { getLists } from '../../api/lists';
import { PagedLists } from '../../api/resources';
import { Pages } from '../../components/pages';
import { PrimaryButton } from '../../components/button';
import { ListInfo } from '../../components/list_info';

export const Lists = () => {
  const axiosInstance = useContext(AxiosContext);
  const [lists, setLists] = useState<PagedLists | undefined>(undefined);
  const [currentPage, setCurrentPage] = useState(0);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getLists(axiosInstance, currentPage)
      .then(setLists)
      .catch((reason) => {
        console.error(reason);
      });
  }, [axiosInstance, currentPage]);

  if (!lists) {
    return <p>Loading...</p>;
  }

  return (
    <>
      <div className="flex flex-row flex-row-reverse">
        <a href="/lists/create">
          <PrimaryButton>Add List</PrimaryButton>
        </a>
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
