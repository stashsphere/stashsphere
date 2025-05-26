import { useContext, useState } from 'react';
import { createImage } from '../../api/image';
import { ThingEditor, ThingEditorData } from '../../components/thing_editor/thing_editor';
import { AxiosContext } from '../../context/axios';
import { createThing } from '../../api/things';
import { useNavigate } from 'react-router';
import { PrimaryButton } from '../../components/shared';

export const CreateThing = () => {
  const axiosInstance = useContext(AxiosContext);
  const navigate = useNavigate();

  const [editedData, setEditedData] = useState<null | ThingEditorData>(null);

  const create = async () => {
    if (!axiosInstance) {
      return;
    }
    if (!editedData) {
      return;
    }
    const images_ids = [];

    for (const file of editedData.images) {
      if (file.type === 'url') {
        images_ids.push(file.image.id);
      } else {
        const image = await createImage(axiosInstance, file.file);
        images_ids.push(image.id);
      }
    }

    const params = {
      name: editedData.name,
      privateNote: editedData.privateNote,
      description: editedData.description,
      imagesIds: images_ids,
      properties: editedData.properties,
      quantity: editedData.quantity,
      quantityUnit: editedData.quantityUnit,
    };

    const thing = await createThing(axiosInstance, params);
    console.log('Created', thing);
    navigate(`/things/${thing.id}`);
  };

  return (
    <ThingEditor onChange={setEditedData}>
      <PrimaryButton onClick={() => create()}>Create</PrimaryButton>
    </ThingEditor>
  );
};
