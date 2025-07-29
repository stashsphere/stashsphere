import { useContext, useEffect, useMemo, useState } from 'react';
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
    if (axiosInstance === null) {
      return;
    }
    getThings(axiosInstance, currentPage, [])
      .then(setSelectableThingsPages)
      .catch((reason) => {
        console.log(reason);
      });
  }, [axiosInstance, currentPage]);

  const edit = async (data: ListEditorData) => {
    if (!axiosInstance || !listId) {
      return;
    }
    const params = {
      name: data.name,
      thingIds: data.selectedThingIDs,
      sharingState: data.sharingState,
    };
    const list = await updateList(axiosInstance, listId, params);
    navigate(`/lists/${list.id}`);
  };

  const data = {
    name: list?.name || '',
    selectedThingIDs: list?.things.map((thing) => thing.id) || [],
    sharingState: list?.sharingState || 'private',
  };

  // TODO: Move to backend
  const selectableThings = useMemo(() => {
    if (selectableThingsPages === undefined) {
      return [];
    }
    return selectableThingsPages.things.filter((t) => t.owner.id === authCtx.profile?.id);
  }, [authCtx.profile?.id, selectableThingsPages]);

  return (
    <ListEditor onChange={edit} list={data} selectableThings={selectableThings}>
      <Pages
        currentPage={currentPage}
        onPageChange={(n) => setCurrentPage(n)}
        pages={selectableThingsPages?.totalPageCount || 0}
      />

      <div className="flex gap-4">
        <YellowButton type="submit">Save</YellowButton>
        <GrayButton>Abort</GrayButton>
      </div>
    </ListEditor>
  );
};
