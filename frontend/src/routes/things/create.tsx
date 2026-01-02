import { useContext, useEffect, useState } from 'react';
import { createImage, modifyImage } from '../../api/image';
import { ThingEditor, ThingEditorData } from '../../components/thing_editor';
import { AxiosContext } from '../../context/axios';
import { createThing } from '../../api/things';
import { useNavigate } from 'react-router';
import { PrimaryButton } from '../../components/shared';
import { getLists, updateList, updateListParamsFromList } from '../../api/lists';
import { AuthContext } from '../../context/auth';
import { List } from '../../api/resources';

export const CreateThing = () => {
  const authContext = useContext(AuthContext);
  const axiosInstance = useContext(AxiosContext);
  const navigate = useNavigate();
  const [lists, setLists] = useState<List[]>([]);

  const [editedData, setEditedData] = useState<null | ThingEditorData>(null);

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

  const create = async () => {
    if (!axiosInstance) {
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

    const createdThing = await createThing(axiosInstance, params);
    console.log('Created', createdThing);

    // TODO move to backend transaction:
    for (const listId of editedData.listIds) {
      const list = lists.find((l) => l.id === listId);
      if (list) {
        const listParams = updateListParamsFromList(list);
        listParams.thingIds = [...listParams.thingIds, createdThing.id];
        await updateList(axiosInstance, listId, listParams);
      }
    }
    navigate(`/things/${createdThing.id}`);
  };

  return (
    <ThingEditor onChange={setEditedData} lists={lists}>
      <PrimaryButton onClick={() => create()}>Create</PrimaryButton>
    </ThingEditor>
  );
};
