import { useMemo } from 'react';
import { Image } from '../../api/resources';
import { AccentButton, DangerButton, PrimaryButton } from '../button';
import { ImageComponent } from '../image';

type InterActionProps = {
  onDelete?: (id: string) => void;
  onSelect?: (id: string, selected: boolean) => void;
};

type ImageGridTileProps = {
  data: Image;
  selected?: boolean;
} & InterActionProps;

export const ImageGridTile = ({
  data: image,
  onDelete,
  onSelect,
  selected,
}: ImageGridTileProps) => {
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
    <div className="relative group bg-white rounded-sm overflow-hidden border border-gray-200 hover:border-gray-300 transition-colors">
      <div className="aspect-square">
        <ImageComponent defaultWidth={150} image={image} className="w-full h-full object-cover" />
      </div>

      <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex flex-col justify-end p-1">
        <div className="text-white text-xs mb-1">
          <div className="font-medium truncate text-xs">{image.name}</div>
          <div className="opacity-75 text-xs">{usedText}</div>
        </div>

        <div className="flex gap-1">
          {onSelect !== undefined &&
            selected !== undefined &&
            (selected ? (
              <AccentButton
                onClick={() => onSelect(image.id, false)}
                className="text-xs py-0.5 px-1 flex-1 min-h-0 h-auto"
              >
                ✓
              </AccentButton>
            ) : (
              <PrimaryButton
                onClick={() => onSelect(image.id, true)}
                className="text-xs py-0.5 px-1 flex-1 min-h-0 h-auto"
              >
                +
              </PrimaryButton>
            ))}
          {onDelete && 'actions' in image && image.actions.canDelete && (
            <DangerButton
              onClick={() => onDelete(image.id)}
              className="text-xs py-0.5 px-1 min-h-0 h-auto"
            >
              ×
            </DangerButton>
          )}
        </div>
      </div>
    </div>
  );
};

type ImageGridProps = {
  images: Image[];
  selectedImageIds?: string[];
} & InterActionProps;

export const ImageGrid = ({ images, selectedImageIds, ...rest }: ImageGridProps) => {
  return (
    <div className="w-full">
      <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-8 gap-1 sm:gap-2">
        {images.map((image) => (
          <ImageGridTile
            key={image.id}
            data={image}
            {...rest}
            selected={selectedImageIds?.includes(image.id)}
          />
        ))}
      </div>
    </div>
  );
};
