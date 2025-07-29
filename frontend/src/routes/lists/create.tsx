import { useContext, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router';
import { AxiosContext } from '../../context/axios';
import { ListEditor, ListEditorData } from '../../components/list_editor';
import { createList } from '../../api/lists';
import { PagedThings } from '../../api/resources';
import { getThings } from '../../api/things';
import { Pages } from '../../components/pages';
import { AuthContext } from '../../context/auth';
import { PrimaryButton } from '../../components/shared';

export const CreateList = () => {
  const authCtx = useContext(AuthContext);
  const axiosInstance = useContext(AxiosContext);
  const navigate = useNavigate();

  const [selectableThingsPages, setSelectableThingsPages] = useState<PagedThings | undefined>(
    undefined
  );
  const [currentPage, setCurrentPage] = useState(0);

  const create = async (data: ListEditorData) => {
    if (!axiosInstance) {
      return;
    }
    const list = await createList(axiosInstance, {
      name: data.name,
      thingIds: data.selectedThingIDs,
      sharingState: data.sharingState,
    });
    console.log('Created', list);
    navigate(`/lists/${list.id}`);
  };

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

  // TODO: Move to backend
  const selectableThings = useMemo(() => {
    if (selectableThingsPages === undefined) {
      return [];
    }
    return selectableThingsPages.things.filter((t) => t.owner.id === authCtx.profile?.id);
  }, [authCtx.profile?.id, selectableThingsPages]);

  return (
    <ListEditor onChange={create} selectableThings={selectableThings}>
      <Pages
        currentPage={currentPage}
        onPageChange={(n) => setCurrentPage(n)}
        pages={selectableThingsPages?.totalPageCount || 0}
      />
      <PrimaryButton type="submit">Create</PrimaryButton>
    </ListEditor>
  );
};
