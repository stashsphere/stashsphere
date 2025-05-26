import { useContext, useEffect, useMemo, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { PagedImages, Image } from '../../api/resources';
import { getImages } from '../../api/image';
import { ImageGrid } from './edit_image_grid';
import { Pages } from '../pages';

export interface ImageBrowserGridProps {
  onSelected: (image: Image[]) => void;
}

export const ImageBrowserGrid = ({ onSelected }: ImageBrowserGridProps) => {
  const axiosInstance = useContext(AxiosContext);
  const [currentPage, setCurrentPage] = useState(0);
  const [images, setImages] = useState<PagedImages | undefined>(undefined);

  const [selectedImages, setSelectedImages] = useState<Image[]>([]);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getImages(axiosInstance, currentPage, 24)
      .then(setImages)
      .catch((reason) => {
        console.log(reason);
      });
  }, [axiosInstance, currentPage]);

  const selectedImageIds = useMemo(() => {
    return selectedImages.map((e) => e.id);
  }, [selectedImages]);

  useEffect(() => {
    onSelected(selectedImages);
  }, [onSelected, selectedImages]);

  const onSelect = (id: string, selected: boolean) => {
    const selectedImage = images?.images.find((e) => e.id === id);
    if (!selectedImage) {
      return;
    }
    const newSelectedImages = selected
      ? [...selectedImages, selectedImage]
      : selectedImages.filter((e) => e.id !== id);
    setSelectedImages(newSelectedImages);
  };

  if (!images) {
    return <p>Loading...</p>;
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex-1 overflow-y-auto overflow-x-hidden">
        <ImageGrid images={images.images} selectedImageIds={selectedImageIds} onSelect={onSelect} />
      </div>
      {images.totalCount > 0 && (
        <div className="mt-3 flex-shrink-0">
          <Pages
            currentPage={currentPage}
            onPageChange={(n) => setCurrentPage(n)}
            pages={images.totalPageCount}
          />
        </div>
      )}
    </div>
  );
};
