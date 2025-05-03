import { useState } from 'react';
import { ReducedImage } from '../../api/resources';
import { ImageComponent } from '../image';

type ImageGalleryProps = {
  images: ReducedImage[];
};

export const ThingImages = ({ images }: ImageGalleryProps) => {
  const [selectedImage, setSelectedImage] = useState<ReducedImage | null>(
    images.length === 0 ? null : images[0]
  );

  // Determine the class for the selected thumbnail
  const getThumbnailClass = (image: ReducedImage) => {
    return image === selectedImage ? 'border-4 border-blue-500' : '';
  };

  return selectedImage ? (
    <div className="w-full relative">
      <ImageComponent
        image={selectedImage}
        defaultWidth={1024}
        className="w-full h-auto mb-4 rounded-sm"
        alt="Main"
      />
      <div className="flex flex-wrap justify-between gap-2">
        {images.map((image) => (
          <ImageComponent
            key={image.id}
            image={image}
            defaultWidth={300}
            alt="Thumbnail"
            onClick={() => setSelectedImage(image)}
            className={`w-24 h-24 object-cover object-center cursor-pointer rounded-sm shadow-sm hover:shadow-lg transition duration-300 ease-in-out transform hover:scale-105 ${getThumbnailClass(image)}`}
          />
        ))}
      </div>
    </div>
  ) : null;
};
