import { useParams } from 'react-router';
import { ListDetails } from '../../components/list_details';

export const ShowList = () => {
  const { listId } = useParams();
  if (listId === undefined) {
    return <p>invalid id</p>;
  } else {
    return <ListDetails id={listId} />;
  }
};
