import { useContext, useEffect, useMemo, useState } from 'react';
import { List } from '../../api/resources';
import { useNavigate, useParams } from 'react-router';
import { AxiosContext } from '../../context/axios';
import { getList, updateList } from '../../api/lists';
import { ListEditor, ListEditorData } from '../../components/list_editor';
import { GrayButton, YellowButton } from '../../components/shared';

export const EditList = () => {
  const [list, setList] = useState<null | List>(null);
  const [editedData, setEditedData] = useState<null | ListEditorData>(null);

  const axiosInstance = useContext(AxiosContext);
  const navigate = useNavigate();
  const { listId } = useParams();

  useEffect(() => {
    if (!axiosInstance || listId == undefined) {
      return;
    }
    getList(axiosInstance, listId).then(setList);
  }, [axiosInstance, listId]);

  const data = useMemo(() => {
    if (!list) return undefined;
    return {
      name: list.name,
      selectedThingIDs: list.things.map((thing) => thing.id),
      sharingState: list.sharingState,
    };
  }, [list]);

  const edit = async () => {
    if (!axiosInstance || !listId || !editedData) {
      return;
    }
    await updateList(axiosInstance, listId, {
      name: editedData.name,
      thingIds: editedData.selectedThingIDs,
      sharingState: editedData.sharingState,
    });
    navigate(`/lists/${listId}`);
  };

  return (
    <ListEditor onChange={setEditedData} list={data}>
      <div className="flex gap-4">
        <YellowButton onClick={edit}>Save</YellowButton>
        <GrayButton onClick={() => navigate(`/lists/${listId}`)}>Abort</GrayButton>
      </div>
    </ListEditor>
  );
};
