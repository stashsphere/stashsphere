import { Image } from '../../api/resources';
import { AccentButton, DangerButton, NeutralButton } from '../shared';
import { ImageComponent } from '../shared';

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
  return (
    <div className="relative group bg-white rounded-sm overflow-hidden border border-gray-200 hover:border-gray-300 transition-colors">
      <div className="aspect-square">
        <ImageComponent defaultWidth={400} image={image} className="w-full h-full object-contain" />
      </div>

      <div className="absolute bg-black/50 flex items-center justify-center gap-1 h-10 bottom-0 w-full flex flex-col">
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
              <NeutralButton
                onClick={() => {
                  return onSelect(image.id, true);
                }}
                className="text-xs py-0.5 px-1 flex-1 min-h-0 h-auto"
              >
                +
              </NeutralButton>
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
