import { ReducedImage, Image } from '../api/resources';
import { ImageComponent } from './shared';

type ImageGridProps = {
  images: ReducedImage[] | Image[];
};

const ImageGrid = (props: ImageGridProps) => {
  const images = props.images.slice(0, props.images.length > 4 ? 4 : props.images.length + 1);

  return (
    <div className="grid grid-cols-2 gap-2 m-4">
      {images.map((image) => (
        <div className="flex w-30 h-30 items-center justify-center rounded-md">
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
