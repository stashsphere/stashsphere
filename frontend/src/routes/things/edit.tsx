import { useContext, useEffect, useMemo, useState } from "react";
import {
  ThingEditor,
  ThingEditorData,
  ThingImage,
} from "../../components/thing_editor";
import { AxiosContext } from "../../context/axios";
import { useNavigate, useParams } from "react-router-dom";
import { getThing, updateThing } from "../../api/things";
import { Thing } from "../../api/resources";
import { createImage } from "../../api/image";
import { GrayButton, PrimaryButton } from "../../components/button";

export const EditThing = () => {
  const [thing, setThing] = useState<null | Thing>(null);
  const axiosInstance = useContext(AxiosContext);
  const navigate = useNavigate();
  const { thingId } = useParams();
  const [editedData, setEditedData] = useState<null | ThingEditorData>(null)

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
    const images_ids = [];
    for (const file of editedData.images) {
      if (file.type === "url") {
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
      quantityUnit: editedData.quantityUnit
    };
    const thing = await updateThing(axiosInstance, thingId, params);
    console.log("Updated", thing);
    navigate(`/things/${thing.id}`);
  };

  const data = useMemo(() => {
    return {
      name: thing?.name || "",
      images:
        thing?.images.map((x) => {
          return { type: "url", image: x } as ThingImage;
        }) || [],
      properties: thing?.properties || [],
      privateNote: thing?.privateNote || "",
      description: thing?.description || "",
      quantity: thing?.quantity || 0,
      quantityUnit: thing?.quantityUnit || ""
    };
  }, [thing]);

  return (
    <ThingEditor onChange={setEditedData} thing={data}>
      <div className="flex gap-4">
        <PrimaryButton onClick={() => edit()}>
          Save
        </PrimaryButton>
        <GrayButton>Abort</GrayButton>
      </div>
    </ThingEditor>
  );
};
