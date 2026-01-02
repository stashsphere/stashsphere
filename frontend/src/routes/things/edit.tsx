import { useContext, useEffect, useMemo, useState } from 'react';
import { ThingEditor, ThingEditorData, ThingImage } from '../../components/thing_editor';
import { AxiosContext } from '../../context/axios';
import { useNavigate, useParams } from 'react-router';
import { getThing, updateThing } from '../../api/things';
import { List, Thing } from '../../api/resources';
import { createImage, modifyImage } from '../../api/image';
import { GrayButton, PrimaryButton } from '../../components/shared';
import { getLists, updateList, updateListParamsFromList } from '../../api/lists';
import { AuthContext } from '../../context/auth';

export const EditThing = () => {
  const authContext = useContext(AuthContext);
  const [thing, setThing] = useState<null | Thing>(null);
  const [lists, setLists] = useState<List[]>([]);
  const axiosInstance = useContext(AxiosContext);
  const navigate = useNavigate();
  const { thingId } = useParams();
  const [editedData, setEditedData] = useState<null | ThingEditorData>(null);

  useEffect(() => {
    if (!axiosInstance || thingId == undefined) {
      return;
    }
    getThing(axiosInstance, thingId).then(setThing);
  }, [axiosInstance, thingId]);

  useEffect(() => {
    if (!axiosInstance) {
      return;
    }
    if (!authContext.profile) {
      return;
    }
    getLists(axiosInstance, 0, 0, [authContext.profile.id], false).then((lists) =>
      setLists(lists.lists)
    );
  }, [authContext.profile, axiosInstance]);

  const edit = async () => {
    if (!axiosInstance || !thingId) {
      return;
    }
    if (!editedData) {
      return;
    }
    const images = [];
    for (const file of editedData.images) {
      if (file.type === 'url') {
        images.push({ id: file.image.id, rotation: file.rotation });
      } else {
        const image = await createImage(axiosInstance, file.file);
        images.push({ id: image.id, rotation: file.rotation });
      }
    }

    for (const image of images) {
      if (image.rotation !== 0) {
        await modifyImage(axiosInstance, image.id, image.rotation);
      }
    }

    const params = {
      name: editedData.name,
      privateNote: editedData.privateNote,
      description: editedData.description,
      imagesIds: images.map((x) => x.id),
      properties: editedData.properties,
      quantity: editedData.quantity,
      quantityUnit: editedData.quantityUnit,
      sharingState: editedData.sharingState,
    };
    const updatedThing = await updateThing(axiosInstance, thingId, params);

    // TODO move to backend transaction:
    const originalListIds = new Set(thing?.lists.map((l) => l.id) || []);
    const newListIds = new Set(editedData.listIds);

    const listsToAdd = editedData.listIds.filter((id) => !originalListIds.has(id));
    const listsToRemove = [...originalListIds].filter((id) => !newListIds.has(id));

    for (const listId of listsToAdd) {
      const list = lists.find((l) => l.id === listId);
      if (list) {
        const listParams = updateListParamsFromList(list);
        listParams.thingIds = [...listParams.thingIds, thingId];
        await updateList(axiosInstance, listId, listParams);
      }
    }

    for (const listId of listsToRemove) {
      const list = lists.find((l) => l.id === listId);
      if (list) {
        const listParams = updateListParamsFromList(list);
        listParams.thingIds = listParams.thingIds.filter((id) => id !== thingId);
        await updateList(axiosInstance, listId, listParams);
      }
    }

    navigate(`/things/${updatedThing.id}`);
  };

  const data = useMemo(() => {
    return {
      name: thing?.name || '',
      images:
        thing?.images.map((x) => {
          return { type: 'url', image: x, rotation: 0 } as ThingImage;
        }) || [],
      properties: thing?.properties || [],
      privateNote: thing?.privateNote || '',
      description: thing?.description || '',
      quantity: thing?.quantity || 0,
      quantityUnit: thing?.quantityUnit || '',
      sharingState: thing?.sharingState || 'private',
      listIds: thing?.lists.map((e) => e.id) || [],
    };
  }, [thing]);

  return (
    <ThingEditor onChange={setEditedData} thing={data} lists={lists}>
      <div className="flex gap-4">
        <PrimaryButton onClick={() => edit()}>Save</PrimaryButton>
        <GrayButton>Abort</GrayButton>
      </div>
    </ThingEditor>
  );
};
