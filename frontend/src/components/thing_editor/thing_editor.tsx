import { ChangeEvent, ReactNode, useContext, useEffect, useMemo, useRef, useState } from 'react';
import PropertyEditor from './property_editor';
import { Property, ReducedImage, Image, SharingState } from '../../api/resources';
import { ConfigContext } from '../../context/config';
import { PrimaryButton, SecondaryButton } from '../shared';
import { Icon, Headline, Modal } from '../shared';
import { ImageBrowserGrid } from './image_browser_grid';
import QuantityEditor from './quantity_editor';
import { urlForImage } from '../../api/image';

export type ThingEditorData = {
  name: string;
  images: ThingImage[];
  description: string;
  privateNote: string;
  properties: Property[];
  quantity: number;
  quantityUnit: string;
  sharingState: SharingState;
};

type ThingEditorProps = {
  children?: ReactNode;
  thing?: ThingEditorData;
  onChange: (thing: ThingEditorData) => void;
};

export type ThingFileImage = {
  type: 'file';
  file: File;
  rotation: number;
};
export type ThingUrlImage = {
  type: 'url';
  image: ReducedImage;
  rotation: number;
};
export type ThingImage = ThingUrlImage | ThingFileImage;

export const ThingEditor = ({ children, thing, onChange }: ThingEditorProps) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [privateNote, setPrivateNote] = useState('');
  const [images, setImages] = useState<ThingImage[]>([]);
  const [properties, setProperties] = useState<Property[]>([]);
  const [showImageBrowser, setShowImageBrowser] = useState(false);
  const [imageBrowserImages, setImageBrowserImages] = useState<Image[]>([]);
  const [quantity, setQuantity] = useState(0);
  const [quantityUnit, setQuantityUnit] = useState('');
  const [sharingState, setSharingState] = useState<SharingState>('private');

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
    setSharingState(thing.sharingState);
  }, [thing]);

  useEffect(() => {
    const data = {
      name,
      images,
      properties,
      description,
      privateNote,
      quantity,
      quantityUnit,
      sharingState,
    };
    onChange(data);
  }, [
    onChange,
    name,
    images,
    properties,
    description,
    privateNote,
    quantity,
    quantityUnit,
    sharingState,
  ]);

  const imageUrl = useMemo(
    () => (image: ReducedImage) => {
      return urlForImage(config, image.hash, 512);
    },
    [config]
  );

  const previewUrls = useMemo(() => {
    const urls = [];

    for (const file of images) {
      if (file.type === 'file') {
        urls.push(URL.createObjectURL(file.file));
      } else {
        urls.push(imageUrl(file.image));
      }
    }
    return urls;
  }, [images, imageUrl]);

  const addFile = (file: File) => {
    if (!file) {
      return;
    }
    const entry: ThingFileImage = { type: 'file', file, rotation: 0 };
    setImages([...images, entry]);
  };

  const selectImages = () => {
    const reducedImages = imageBrowserImages.map((image) => {
      return {
        type: 'url',
        image: image as ReducedImage,
        rotation: 0,
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
  };

  const clampRotation = (value: number) => {
    const x = value % 360;
    if (x < 0) {
      return 360 + x;
    } else {
      return x;
    }
  };

  // this will need these tailwind classes
  // leave it so tailwind picks it up
  // -rotate-90 -rotate-180 -rotate-270
  const rotateLeft = (idx: number) => {
    const rotatedImage = images[idx];
    rotatedImage.rotation = clampRotation(rotatedImage.rotation + 90);
    images[idx] = rotatedImage;
    setImages([...images]);
  };

  const rotateRight = (idx: number) => {
    const rotatedImage = images[idx];
    rotatedImage.rotation = clampRotation(rotatedImage.rotation - 90);
    images[idx] = rotatedImage;
    setImages([...images]);
  };

  const moveLeft = (idx: number) => {
    if (idx == 0) {
      return;
    }
    if (idx > images.length) {
      return;
    }
    const tmp = images[idx - 1];
    images[idx - 1] = images[idx];
    images[idx] = tmp;
    setImages([...images]);
  };

  const moveRight = (idx: number) => {
    if (idx < 0) {
      return;
    }
    if (idx + 1 >= images.length) {
      return;
    }
    const tmp = images[idx + 1];
    images[idx + 1] = images[idx];
    images[idx] = tmp;
    setImages([...images]);
  };

  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <div className="flex flex-col gap-8">
      <Headline type="h1">{name || 'New Thing'}</Headline>

      <div className="flex flex-col lg:flex-row gap-6">
        <div className="flex-1 min-w-0">
          <div className="mb-6">
            <Headline type="h2">Images</Headline>
            {images.length > 0 ? (
              <div className="grid grid-cols-3 gap-2 mb-4">
                {previewUrls.map((url, idx) => (
                  <div key={url} className="relative group">
                    <div className="aspect-square relative overflow-hidden rounded-sm border border-gray-300">
                      <img
                        className={`w-full h-full object-cover -rotate-${images[idx].rotation}`}
                        src={url}
                        alt="Preview"
                      />
                      <div className="absolute bg-black/50 flex items-center justify-center gap-1 h-10 bottom-0 w-full">
                        <button
                          onClick={() => moveLeft(idx)}
                          className="p-1 bg-white/20 rounded-full hover:bg-white/30 transition-colors flex"
                        >
                          <Icon
                            icon="mdi--arrow-left-circle-outline"
                            className="text-white text-xs"
                          />
                        </button>
                        <button
                          onClick={() => rotateLeft(idx)}
                          className="p-1 bg-white/20 rounded-full hover:bg-white/30 transition-colors flex"
                        >
                          <Icon icon="mdi--rotate-left" className="text-white text-xs" />
                        </button>
                        <button
                          onClick={() => removeFile(idx)}
                          className="p-1 bg-red-500/80 rounded-full hover:bg-red-500 transition-colors flex"
                        >
                          <Icon icon="mdi--trash" className="text-white text-xs" />
                        </button>
                        <button
                          onClick={() => rotateRight(idx)}
                          className="p-1 bg-white/20 rounded-full hover:bg-white/30 transition-colors flex"
                        >
                          <Icon icon="mdi--rotate-right" className="text-white text-xs" />
                        </button>
                        <button
                          onClick={() => moveRight(idx)}
                          className="p-1 bg-white/20 rounded-full hover:bg-white/30 transition-colors flex"
                        >
                          <Icon
                            icon="mdi--arrow-right-circle-outline"
                            className="text-white text-xs"
                          />
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center mb-4">
                <p className="text-gray-500">No images added yet</p>
              </div>
            )}

            <div className="flex gap-4">
              <input
                ref={inputRef}
                type="file"
                accept="image/*"
                onChange={onFileChange}
                multiple
                hidden
              />
              <PrimaryButton onClick={() => setShowImageBrowser(true)}>
                Select from Gallery
              </PrimaryButton>
              <PrimaryButton onClick={() => inputRef.current?.click()}>
                Upload from Device
              </PrimaryButton>
            </div>
          </div>
        </div>

        <div className="flex flex-col flex-1 min-w-0 gap-6">
          <div>
            <label htmlFor="name" className="block text-primary text-sm font-medium mb-1">
              Name
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full p-2 text-display border border-gray-300 rounded-sm"
            />
          </div>

          <div>
            <Headline type="h2">Quantity</Headline>
            <QuantityEditor
              quantity={quantity}
              unit={quantityUnit}
              onChange={(q, u) => {
                setQuantity(q);
                setQuantityUnit(u);
              }}
            />
          </div>

          <PropertyEditor properties={properties} onUpdateProperties={setProperties} />

          <div>
            <Headline type="h2">Description</Headline>
            <textarea
              id="description"
              name="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full p-2 text-display border border-gray-300 rounded-sm"
              rows={4}
              placeholder="Describe this thing..."
            />
          </div>

          <div>
            <Headline type="h2">Private Note</Headline>
            <textarea
              id="privateNote"
              name="privateNote"
              value={privateNote}
              onChange={(e) => setPrivateNote(e.target.value)}
              className="w-full p-2 text-display border border-gray-300 rounded-sm bg-warning/10"
              rows={3}
              placeholder="Private notes (visible to you only)..."
            />
          </div>

          {children}
        </div>
      </div>

      <Modal
        isOpen={showImageBrowser}
        onClose={() => setShowImageBrowser(false)}
        title="Select Images from Gallery"
        size="full"
        footer={
          <div className="flex gap-4 justify-end">
            <SecondaryButton onClick={() => setShowImageBrowser(false)}>Cancel</SecondaryButton>
            <PrimaryButton
              onClick={() => {
                selectImages();
                setShowImageBrowser(false);
              }}
            >
              Add Selected Images
            </PrimaryButton>
          </div>
        }
      >
        <ImageBrowserGrid onSelected={setImageBrowserImages} multiple={true} />
      </Modal>
    </div>
  );
};
