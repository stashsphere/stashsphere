import { useMemo } from 'react';
import { Masonry } from 'masonic';
import { Image } from '../api/resources';
import { AccentButton, DangerButton, PrimaryButton } from './shared';
import { ImageComponent } from './shared';

type InterActionProps = {
  onDelete?: (id: string) => void;
  onSelect?: (id: string, selected: boolean) => void;
};

type ImageTileProps = {
  index: number;
  width: number;
  data: Image;
  selected?: boolean;
} & InterActionProps;

export const ImageTile = ({ data: image, onDelete, onSelect, selected }: ImageTileProps) => {
  const usedText = useMemo(() => {
    switch (image.things.length) {
      case 0:
        return 'Used by no things';
      case 1:
        return 'Used by one thing';
      case 2:
        return 'Used by two things';
      default:
        return `Used by ${image.things.length} things`;
    }
  }, [image]);

  return (
    <div
      className="flex flex-col rounded-sm
      group
      hover:scale-102 hover:shadow-sm/30
      transition duration-300 ease-in-out transform"
      tabIndex={0}
    >
      <ImageComponent
        defaultWidth={1024}
        image={image}
        className="object-contain w-full h-full rounded-sm"
      />
      <div
        className="flex flex-col absolute opacity-0 bottom-0 left-0 right-0 bg-content/90 pointer-events-none
        group-hover:opacity-100 group-focus:opacity-100 group-hover:pointer-events-auto group-focus:pointer-events-auto
        transition duration-300 ease-in-out transform"
      >
        <span className="text-display">Name: {image.name}</span>
        <span className="text-display">{usedText}</span>
        {onDelete && image.actions.canDelete && (
          <DangerButton onClick={() => onDelete(image.id)}>Delete</DangerButton>
        )}
        {onSelect !== undefined &&
          selected !== undefined &&
          (selected ? (
            <AccentButton onClick={() => onSelect(image.id, false)}>Unselect</AccentButton>
          ) : (
            <PrimaryButton onClick={() => onSelect(image.id, true)}>Select</PrimaryButton>
          ))}
      </div>
    </div>
  );
};

type ImageListProps = {
  images: Image[];
  selectedImageIds?: string[];
} & InterActionProps;

export const ImageList = ({ images, selectedImageIds, ...rest }: ImageListProps) => {
  // HACK: think about a better way to do this
  // prevents weakmap error
  // could be mitigated by increasing the amount of items in the cache?
  const masonryKey = useMemo(() => images.map((img) => img.id).join('-'), [images]);

  return (
    <div className="w-full flex">
      <Masonry
        key={masonryKey}
        items={images}
        maxColumnCount={3}
        columnGutter={20}
        render={({ data, width, index }) => (
          <ImageTile
            key={data.id}
            data={data}
            width={width}
            index={index}
            {...rest}
            selected={selectedImageIds?.includes(data.id)}
            onDelete={rest.onDelete}
            onSelect={rest.onSelect}
          />
        )}
      />
    </div>
  );
};
