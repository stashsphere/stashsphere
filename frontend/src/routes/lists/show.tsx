import { useParams } from 'react-router';
import { ListDetails } from '../../components/list_details';
import { deleteList, getList } from '../../api/lists';
import { useCallback, useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { List } from '../../api/resources';
import { useNavigate } from 'react-router';

export const ShowList = () => {
  const [list, setList] = useState<null | List>(null);
  const axiosInstance = useContext(AxiosContext);
  const { listId } = useParams();
  const navigate = useNavigate();

  useEffect(() => {
    if (!axiosInstance) {
      return;
    }
    if (!listId) {
      return;
    }
    getList(axiosInstance, listId).then(setList);
  }, [listId, axiosInstance]);

  const onDelete = useCallback(async () => {
    if (list === null) {
      return;
    }
    if (axiosInstance === null) {
      return;
    }
    await deleteList(axiosInstance, list.id);
    navigate(`/lists`);
  }, [axiosInstance, navigate, list]);

  if (listId === undefined) {
    return <p>invalid id</p>;
  } else if (list === null) {
    return <h1>Loading</h1>;
  } else {
    return <ListDetails list={list} onDelete={onDelete} />;
  }
};
