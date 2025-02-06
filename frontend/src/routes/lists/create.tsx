import { useContext, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { AxiosContext } from "../../context/axios";
import { ListEditor, ListEditorData } from "../../components/list_editor";
import { createList } from "../../api/lists";
import { PagedThings } from "../../api/resources";
import { getThings } from "../../api/things";
import { Pages } from "../../components/pages";

export const CreateList = () => {
  const axiosInstance = useContext(AxiosContext);
  const navigate = useNavigate();
  
  const [selectableThings, setSelectableThings] = useState<PagedThings | undefined>(undefined)
  const [currentPage, setCurrentPage] = useState(0);
  
  const create = async (data: ListEditorData) => {
    if (!axiosInstance) {
        return;
    }
    const list = await createList(axiosInstance, {
      name: data.name,
      thing_ids: data.selectedThingIDs,
    });
    console.log("Created", list);
    navigate(`/lists/${list.id}`);
  }
  
  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getThings(axiosInstance, currentPage)
      .then(setSelectableThings)
      .catch((reason) => {
        console.log(reason);
      });
  }, [axiosInstance, currentPage]);

  return (
    <ListEditor onChange={create} selectableThings={selectableThings?.things || []}>
      <Pages
        currentPage={currentPage}
        onPageChange={(n) => setCurrentPage(n)}
        pages={selectableThings?.totalPageCount || 0}
      />
      <button
        type="submit"
        className="bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 focus:outline-none focus:ring focus:border-blue-300"
      >
        Create
      </button>
    </ListEditor>
  );
};
