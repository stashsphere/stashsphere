import { ReducedImage, Image } from '../api/resources';
import { ImageComponent } from './shared';

type ImageGridProps = {
  images: ReducedImage[] | Image[];
  compact?: boolean;
};

const ImageGrid = ({ images: allImages, compact = false }: ImageGridProps) => {
  const images = allImages.slice(0, allImages.length > 4 ? 4 : allImages.length + 1);

  const containerClass = compact ? 'grid grid-cols-2 gap-2 m-2' : 'grid grid-cols-2 gap-2 m-4';
  const imageClass = compact
    ? 'flex w-24 h-24 items-center justify-center rounded-md'
    : 'flex w-30 h-30 items-center justify-center rounded-md';

  return (
    <div className={containerClass}>
      {images.map((image) => (
        <div className={imageClass} key={image.id}>
          <ImageComponent
            image={image}
            className="object-contain w-full h-full"
            defaultWidth={256}
          />
        </div>
      ))}
    </div>
  );
};

export default ImageGrid;
