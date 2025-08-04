import { useContext, useState } from 'react';
import { createImage, modifyImage } from '../../api/image';
import { ThingEditor, ThingEditorData } from '../../components/thing_editor';
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
