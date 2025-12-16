import { useContext, useState } from 'react';
import { useNavigate } from 'react-router';
import { AxiosContext } from '../../context/axios';
import { ListEditor, ListEditorData } from '../../components/list_editor';
import { createList } from '../../api/lists';
import { PrimaryButton } from '../../components/shared';

export const CreateList = () => {
  const axiosInstance = useContext(AxiosContext);
  const navigate = useNavigate();

  const [editedData, setEditedData] = useState<ListEditorData>({
    name: '',
    selectedThingIDs: [],
    sharingState: 'private',
  });

  const create = async () => {
    if (!axiosInstance) {
      return;
    }
    const list = await createList(axiosInstance, {
      name: editedData.name,
      thingIds: editedData.selectedThingIDs,
      sharingState: editedData.sharingState,
    });
    console.log('Created', list);
    navigate(`/lists/${list.id}`);
  };

  return (
    <ListEditor list={editedData} onChange={setEditedData}>
      <PrimaryButton onClick={create}>Create</PrimaryButton>
    </ListEditor>
  );
};
