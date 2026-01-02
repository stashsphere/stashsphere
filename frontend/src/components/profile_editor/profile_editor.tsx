import { ChangeEvent, ReactNode, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { Profile, ReducedImage, Image } from '../../api/resources';
import { ImageBrowserGrid } from '../thing_editor/image_browser_grid';
import { urlForImage } from '../../api/image';
import { Headline, Icon, Modal, PrimaryButton, SecondaryButton } from '../shared';
import { ConfigContext } from '../../context/config';

export type EditableProfile = {
  name: string;
  fullName: string;
  information: string;
  image: ProfileImage | null;
};

type Props = {
  profile: Profile;
  onUpdateProfile: (profile: EditableProfile) => void;
  children?: ReactNode;
};

export type ProfileFileImage = {
  type: 'file';
  file: File;
  rotation: number;
};
export type ProfileUrlImage = {
  type: 'url';
  image: ReducedImage;
  rotation: number;
};
export type ProfileImage = ProfileFileImage | ProfileUrlImage;

export const ProfileEditor = ({ children, profile, onUpdateProfile }: Props) => {
  const [name, setName] = useState('');
  const [fullName, setFullName] = useState('');
  const [information, setInformation] = useState('');

  const [image, setImage] = useState<ProfileImage | null>(null);
  const [showImageBrowser, setShowImageBrowser] = useState(false);
  const [imageBrowserImages, setImageBrowserImages] = useState<Image[]>([]);
  const config = useContext(ConfigContext);

  useEffect(() => {
    setName(profile.name);
    setFullName(profile.fullName);
    setInformation(profile.information);
    if (profile.image) {
      setImage({
        type: 'url',
        image: profile.image,
        rotation: 0,
      });
    }
  }, [profile]);

  useEffect(() => {
    const data = {
      name,
      fullName,
      information,
      image,
    };
    onUpdateProfile(data);
  }, [fullName, image, information, name, onUpdateProfile]);

  const inputRef = useRef<HTMLInputElement>(null);

  const onFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files![0];
    if (!file) {
      return;
    }
    setImage({ type: 'file', file, rotation: 0 });
  };

  const selectImage = () => {
    if (imageBrowserImages.length === 0) {
      return;
    }
    const browserImage = imageBrowserImages[0];

    setImage({ type: 'url', image: browserImage as ReducedImage, rotation: 0 } as ProfileUrlImage);
  };

  const imageUrl = useMemo(
    () => (image: ReducedImage) => {
      return urlForImage(config, image.hash, 512);
    },
    [config]
  );

  const previewUrl = useMemo(() => {
    if (!image) {
      return null;
    }
    if (image.type === 'file') {
      return URL.createObjectURL(image.file);
    } else {
      return imageUrl(image.image);
    }
  }, [image, imageUrl]);

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
  const rotateLeft = () => {
    if (!image) {
      return;
    }
    image.rotation = clampRotation(image.rotation + 90);
    setImage({ ...image });
  };

  const rotateRight = () => {
    if (!image) {
      return;
    }
    image.rotation = clampRotation(image.rotation + 90);
    setImage({ ...image });
  };

  const removeFile = () => {
    setImage(null);
  };

  return (
    <div>
      <div className="mb-4">
        <label htmlFor="name" className="block text-primary text-sm font-medium">
          Name
        </label>
        <input
          type="text"
          id="name"
          name="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="mt-1 p-2 border border-gray-300 rounded-sm text-display"
        />
      </div>
      <div className="mb-4">
        <label htmlFor="realName" className="block text-primary text-sm font-medium">
          Full Name
        </label>
        <input
          type="text"
          id="fullName"
          name="fullName"
          value={fullName}
          onChange={(e) => setFullName(e.target.value)}
          className="mt-1 p-2 border border-gray-300 rounded-sm text-display"
        />
      </div>
      <div className="mb-4">
        <label htmlFor="realName" className="block text-primary text-sm font-medium">
          Information
        </label>
        <textarea
          id="information"
          name="information"
          value={information}
          rows={4}
          onChange={(e) => setInformation(e.target.value)}
          className="mt-1 p-2 border border-gray-300 rounded-sm text-display w-full"
        />
      </div>
      <Headline type="h2">Image</Headline>
      {image && previewUrl && (
        <div className="relative group w-64 h-64">
          <div className="aspect-square relative overflow-hidden rounded-sm border border-gray-300">
            <img
              className={`w-full h-full object-cover -rotate-${image.rotation}`}
              src={previewUrl}
              alt="Preview"
            />
            <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex items-center justify-center gap-1">
              <button
                onClick={() => rotateLeft()}
                className="p-1 bg-white/20 rounded-full hover:bg-white/30 transition-colors"
              >
                <Icon icon="mdi--rotate-left" className="text-white text-xs" />
              </button>
              <button
                onClick={() => removeFile()}
                className="p-1 bg-red-500/80 rounded-full hover:bg-red-500 transition-colors"
              >
                <Icon icon="mdi--trash" className="text-white text-xs" />
              </button>
              <button
                onClick={() => rotateRight()}
                className="p-1 bg-white/20 rounded-full hover:bg-white/30 transition-colors"
              >
                <Icon icon="mdi--rotate-right" className="text-white text-xs" />
              </button>
            </div>
          </div>
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
        <PrimaryButton onClick={() => setShowImageBrowser(true)}>Select from Gallery</PrimaryButton>
        <PrimaryButton onClick={() => inputRef.current?.click()}>Upload from Device</PrimaryButton>
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
                selectImage();
                setShowImageBrowser(false);
              }}
            >
              Add Selected Images
            </PrimaryButton>
          </div>
        }
      >
        <ImageBrowserGrid
          onSelected={setImageBrowserImages}
          multiple={false}
          onlyUnassigned={true}
        />
      </Modal>

      {children}
    </div>
  );
};
