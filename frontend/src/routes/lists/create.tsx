import { useContext, useEffect, useState } from 'react';
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

  return (
    <div>
      <ListEditor
        list={editedData}
        onChange={setEditedData}
        selectableThings={selectableThingsPages?.things || []}
      >
        <PrimaryButton onClick={create}>Create</PrimaryButton>
      </ListEditor>
      <Pages
        currentPage={currentPage}
        onPageChange={(n) => setCurrentPage(n)}
        pages={selectableThingsPages?.totalPageCount || 0}
      />
    </div>
  );
};
