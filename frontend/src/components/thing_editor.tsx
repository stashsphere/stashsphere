import {
  ChangeEvent,
  ReactNode,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import PropertyEditor from "./property_editor";
import { Property, ReducedImage, Image } from "../api/resources";
import { ConfigContext } from "../context/config";
import { DangerButton, PrimaryButton } from "./button";
import { Icon } from "./icon";
import { ImageBrowser } from "./image_browser";
import QuantityEditor from "./quantity_editor";

export type ThingEditorData = {
  name: string;
  images: ThingImage[];
  description: string;
  privateNote: string;
  properties: Property[];
  quantity: number;
  quantityUnit: string;
};

type ThingEditorProps = {
  children?: ReactNode;
  thing?: ThingEditorData;
  onChange: (thing: ThingEditorData) => void;
};

export type ThingFileImage = {
  type: "file";
  file: File;
};
export type ThingUrlImage = {
  type: "url";
  image: ReducedImage;
};
export type ThingImage = ThingUrlImage | ThingFileImage;

export const ThingEditor = ({
  children,
  thing,
  onChange,
}: ThingEditorProps) => {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [privateNote, setPrivateNote] = useState("");
  const [images, setImages] = useState<ThingImage[]>([]);
  const [properties, setProperties] = useState<Property[]>([]);
  const [showImageBrowser, setShowImageBrowser] = useState(false);
  const [imageBrowserImages, setImageBrowserImages] = useState<Image[]>([]);
  const [quantity, setQuantity] = useState(0);
  const [quantityUnit, setQuantityUnit] = useState("");

  const config = useContext(ConfigContext);

  useEffect(() => {
    if (!thing) {
      return;
    }
    setName(thing.name);
    setPrivateNote(thing.privateNote);
    setDescription(thing.description);
    setImages(thing.images);
    setProperties(thing.properties);
    setQuantity(thing.quantity);
    setQuantityUnit(thing.quantityUnit);
  }, [thing]);

  useEffect(() => {
    const data = {
      name,
      images,
      properties,
      description,
      privateNote,
      quantity,
      quantityUnit
    };
    onChange(data);
  }, [onChange, name, images, properties, description, privateNote, quantity, quantityUnit])

  const urlForImage = useMemo(
    () => (image: ReducedImage) => {
      return `${config.apiHost}/api/images/${image.id}`;
    },
    [config]
  );

  const previewUrls = useMemo(() => {
    const urls = [];

    for (const file of images) {
      if (file.type === "file") {
        urls.push(URL.createObjectURL(file.file));
      } else {
        urls.push(urlForImage(file.image));
      }
    }
    return urls;
  }, [images, urlForImage]);

  const addFile = (file: File) => {
    if (!file) {
      return;
    }
    const entry: ThingFileImage = { type: "file", file };
    setImages([...images, entry]);
  };

  const selectImages = () => {
    const reducedImages = imageBrowserImages.map((image) => {
      return {
        type: "url",
        image: image as ReducedImage,
      } as ThingUrlImage;
    });
    setImages([...images, ...reducedImages]);
  };

  const removeFile = (idx: number) => {
    const newFiles = images.filter((_, i) => i !== idx);
    setImages(newFiles);
  };

  const onFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files![0];
    addFile(file);
  }

  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <div>
      <div className="mb-4">
        <label
          htmlFor="name"
          className="block text-primary text-sm font-medium"
        >
          Name
        </label>
        <input
          type="text"
          id="name"
          name="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="mt-1 p-2 text-display border border-gray-300 rounded"
        />
      </div>

      <div className="mb-4">
        <label
          htmlFor="description"
          className="block text-primary text-sm font-medium"
        >
          Description (visible to others)
        </label>
        <textarea
          id="description"
          name="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          className="mt-1 p-2 text-display border border-gray-300 rounded w-1/2"
        />
      </div>

      <div className="mb-4">
        <label
          htmlFor="privateNote"
          className="block text-warning text-sm font-medium"
        >
          Private Note (visible to you only)
        </label>
        <textarea
          id="privateNote"
          name="privateNote"
          value={privateNote}
          onChange={(e) => setPrivateNote(e.target.value)}
          className="mt-1 p-2 text-display border border-gray-300 rounded w-1/2"
        />
      </div>

      <h2 className="text-xl font-bold mb-4 text-secondary">Quantity</h2>
      <div className="mb-4">
        <QuantityEditor
          quantity={quantity}
          unit={quantityUnit}
          onChange={(q, u) => {
            setQuantity(q);
            setQuantityUnit(u);
          }}
        />
      </div>

      <h2 className="text-xl font-bold mb-4 text-secondary">Images</h2>
      <div className="mb-4">
        <div className="flex flex-wrap gap-4">
          {previewUrls.map((url, idx) => (
            <div key={url}>
              <div className="flex items-center gap-4 mb-2 flex-col">
                <img className="w-96" src={url} alt="Preview" />
                <DangerButton onClick={() => removeFile(idx)}>
                  <Icon icon="mdi--trash" />
                  Remove
                </DangerButton>
              </div>
            </div>
          ))}
        </div>
        {showImageBrowser &&
          <>
            <ImageBrowser onSelected={setImageBrowserImages} />
            <PrimaryButton onClick={() => {
              selectImages();
              setShowImageBrowser(false);
            }}>Add selected images</PrimaryButton>
          </>}
        {!showImageBrowser &&
          <div className="flex gap-4">
            <input ref={inputRef} type="file" accept="image/*" onChange={onFileChange} multiple hidden />
            <PrimaryButton onClick={() => setShowImageBrowser(true)}>Select from Images</PrimaryButton>
            <PrimaryButton onClick={() => inputRef.current?.click()}>Choose from Filesystem</PrimaryButton>
          </div>}
      </div>

      <PropertyEditor
        properties={properties}
        onUpdateProperties={setProperties}
      />

      {!showImageBrowser && children}
    </div>
  );
};
