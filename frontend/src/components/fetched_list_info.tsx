import { useContext, useEffect, useState } from 'react';
import { List } from '../api/resources';
import { getList } from '../api/lists';
import { AxiosContext } from '../context/axios';
import { ListInfo } from './list_info';

type FetchedListInfoProps = {
  listId: string;
  compact?: boolean;
};

export const FetchedListInfo = ({ listId, compact }: FetchedListInfoProps) => {
  const axiosInstance = useContext(AxiosContext);
  const [list, setList] = useState<List | null>(null);

  useEffect(() => {
    if (!axiosInstance) return;
    getList(axiosInstance, listId).then(setList);
  }, [axiosInstance, listId]);

  if (!list) {
    return null;
  }

  return <ListInfo list={list} compact={compact} />;
};
