import { useContext, useEffect, useMemo, useState } from 'react';
import {
  ThingEditor,
  ThingEditorData,
  ThingImage,
} from '../../components/thing_editor/thing_editor';
import { AxiosContext } from '../../context/axios';
import { useNavigate, useParams } from 'react-router';
import { getThing, updateThing } from '../../api/things';
import { Thing } from '../../api/resources';
import { createImage, modifyImage } from '../../api/image';
import { GrayButton, PrimaryButton } from '../../components/shared';

export const EditThing = () => {
  const [thing, setThing] = useState<null | Thing>(null);
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
    };
    const thing = await updateThing(axiosInstance, thingId, params);
    console.log('Updated', thing);
    navigate(`/things/${thing.id}`);
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
    };
  }, [thing]);

  return (
    <ThingEditor onChange={setEditedData} thing={data}>
      <div className="flex gap-4">
        <PrimaryButton onClick={() => edit()}>Save</PrimaryButton>
        <GrayButton>Abort</GrayButton>
      </div>
    </ThingEditor>
  );
};
