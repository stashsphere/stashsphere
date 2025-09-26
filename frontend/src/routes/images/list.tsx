import { useContext, useEffect, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { PagedImages } from '../../api/resources';
import { deleteImage, getImages } from '../../api/image';
import { Pages } from '../../components/pages';
import { ImageList } from '../../components/image_list';
import { PrimaryButton } from '../../components/shared';
import { ImageUploader } from '../../components/image_uploader';
import { Toggle } from '../../components/shared/toggle';

export const Images = () => {
  const axiosInstance = useContext(AxiosContext);
  const [images, setImages] = useState<PagedImages | undefined>(undefined);
  const [showUploader, setShowUploader] = useState(false);
  const [currentPage, setCurrentPage] = useState(0);
  const [uploadedKey, setUploadedKey] = useState(0);
  const [onlyUnassigned, setOnlyUnassigned] = useState(true);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getImages(axiosInstance, currentPage, 18, onlyUnassigned)
      .then(setImages)
      .catch((reason) => {
        console.log(reason);
      });
  }, [axiosInstance, currentPage, onlyUnassigned, uploadedKey]);

  const onUpload = () => {
    setUploadedKey((key) => key + 1);
    setShowUploader(false);
  };

  const onDelete = (id: string) => {
    if (axiosInstance === null) {
      return;
    }
    const asyncDo = async () => {
      await deleteImage(axiosInstance, id);
      setUploadedKey((key) => key + 1);
    };
    asyncDo();
  };

  if (!images) {
    return <p>Loading...</p>;
  }
  return (
    <>
      <div className="flex mb-2 gap-4">
        {showUploader ? (
          <PrimaryButton onClick={() => setShowUploader(false)}>Hide browser</PrimaryButton>
        ) : (
          <PrimaryButton onClick={() => setShowUploader(true)}>Upload more images</PrimaryButton>
        )}
        <Toggle value={onlyUnassigned} onChange={(v) => setOnlyUnassigned(v)}>
          <span className="ms-3 text-sm font-medium text-display">Show only unassigned images</span>
        </Toggle>
        <label className="inline-flex items-center cursor-pointer">
          <input
            type="checkbox"
            checked={onlyUnassigned}
            className="sr-only peer"
            onChange={(e) => setOnlyUnassigned(e.target.checked)}
          />
          <div className="relative w-11 h-6 bg-secondary-500 rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-primary after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-primary after:border-primary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-secondary-600"></div>
          <span className="ms-3 text-sm font-medium text-display">Show only unassigned images</span>
        </label>
      </div>
      {showUploader ? (
        <div className="my-2">
          <ImageUploader onUpload={onUpload} />
        </div>
      ) : null}
      <ImageList images={images.images} onDelete={onDelete} />
      {images.totalCount === 0 ? <p className="mt-3 text-display">No images yet</p> : null}
      {images.totalCount > 0 && (
        <Pages
          currentPage={currentPage}
          onPageChange={(n) => setCurrentPage(n)}
          pages={images.totalPageCount}
        />
      )}
    </>
  );
};
