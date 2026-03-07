import { useContext, useState } from 'react';
import { useNavigate } from 'react-router';
import { AxiosContext } from '../../context/axios';
import { AuthContext } from '../../context/auth';
import { ListEditor, ListEditorData } from '../../components/list_editor';
import { createList } from '../../api/lists';
import { PrimaryButton } from '../../components/shared';

export const CreateList = () => {
  const axiosInstance = useContext(AxiosContext);
  const authContext = useContext(AuthContext);
  const navigate = useNavigate();

  const [editedData, setEditedData] = useState<null | ListEditorData>(null);

  const create = async () => {
    if (!axiosInstance || !editedData) {
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
    <ListEditor
      onChange={setEditedData}
      defaultSharingState={authContext.profile?.defaultSharingState}
    >
      <PrimaryButton onClick={create}>Create</PrimaryButton>
    </ListEditor>
  );
};
