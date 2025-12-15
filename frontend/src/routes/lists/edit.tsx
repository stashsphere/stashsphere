import { useContext, useEffect, useState } from 'react';
import { List, PagedThings } from '../../api/resources';
import { useNavigate, useParams } from 'react-router';
import { AxiosContext } from '../../context/axios';
import { getList, updateList } from '../../api/lists';
import { ListEditor, ListEditorData } from '../../components/list_editor';
import { GrayButton, YellowButton } from '../../components/shared';
import { Pages } from '../../components/pages';
import { getThings } from '../../api/things';
import { AuthContext } from '../../context/auth';

export const EditList = () => {
  const [list, setList] = useState<null | List>(null);
  const [editedData, setEditedData] = useState<ListEditorData>({
    name: '',
    selectedThingIDs: [],
    sharingState: 'private',
  });

  const axiosInstance = useContext(AxiosContext);
  const navigate = useNavigate();
  const { listId } = useParams();
  const authCtx = useContext(AuthContext);

  const [selectableThingsPages, setSelectableThingsPages] = useState<PagedThings | undefined>(
    undefined
  );
  const [currentPage, setCurrentPage] = useState(0);

  useEffect(() => {
    if (!axiosInstance || listId == undefined) {
      return;
    }
    getList(axiosInstance, listId).then(setList);
  }, [axiosInstance, listId]);

  useEffect(() => {
    if (list === null) {
      return;
    }
    setEditedData({
      name: list.name,
      selectedThingIDs: list.things.map((thing) => thing.id),
      sharingState: list.sharingState,
    });
  }, [list]);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    if (authCtx.profile === null) {
      return;
    }
    getThings(axiosInstance, currentPage, 10, [authCtx.profile.id])
      .then(setSelectableThingsPages)
      .catch((reason) => {
        console.log(reason);
      });
  }, [authCtx.profile, axiosInstance, currentPage]);

  const edit = async () => {
    if (!axiosInstance || !listId) {
      return;
    }
    const params = {
      name: editedData.name,
      thingIds: editedData.selectedThingIDs,
      sharingState: editedData.sharingState,
    };
    const list = await updateList(axiosInstance, listId, params);
    navigate(`/lists/${list.id}`);
  };

  return (
    <div>
      <ListEditor
        onChange={setEditedData}
        list={editedData}
        selectableThings={selectableThingsPages?.things || []}
      >
        <div className="flex gap-4">
          <YellowButton onClick={edit}>Save</YellowButton>
          <GrayButton>Abort</GrayButton>
        </div>
      </ListEditor>
      <Pages
        currentPage={currentPage}
        onPageChange={(n) => setCurrentPage(n)}
        pages={selectableThingsPages?.totalPageCount || 0}
      />
    </div>
  );
};
