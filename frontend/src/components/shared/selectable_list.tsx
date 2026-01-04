import { useCallback, useMemo } from 'react';
import { List } from '../../api/resources';
import { Icon, ImageComponent } from '.';

type SelectableListProps = {
  list: List;
  selected: boolean;
  onSelect: (listId: string, selected: boolean) => void;
};

export const SelectableList = ({ list, selected, onSelect }: SelectableListProps) => {
  const handleClick = useCallback(() => {
    onSelect(list.id, !selected);
  }, [list.id, selected, onSelect]);

  // Get up to 4 images from the list's things for a preview grid
  const previewImages = useMemo(() => {
    return list.things
      .map((thing) => thing.images[0])
      .filter((img) => img !== undefined)
      .slice(0, 4);
  }, [list.things]);

  const thingCount = list.things.length;

  const previewContent =
    previewImages.length > 0 ? (
      <div className="grid grid-cols-2 grid-rows-2 gap-1 w-full h-full">
        {previewImages.map((image) => (
          <div key={image.id} className="flex items-center justify-center overflow-hidden">
            <ImageComponent
              image={image}
              defaultWidth={128}
              className="object-contain h-full w-full"
            />
          </div>
        ))}
      </div>
    ) : (
      <span>
        <Icon icon="mdi--format-list-bulleted" className="text-4xl" />
      </span>
    );

  return (
    <div
      className={`relative flex flex-col gap-2 items-start border rounded-md p-2 cursor-pointer transition-colors ${
        selected ? 'bg-accent/10 border-accent' : 'border-secondary'
      }`}
      onClick={handleClick}
    >
      {selected && (
        <div className="absolute top-2 right-2 z-10">
          <Icon icon="mdi--check-circle" className="text-accent" size="small" />
        </div>
      )}
      <div className="flex w-40 h-40 items-center justify-center bg-brand-900 p-2 rounded-md">
        {previewContent}
      </div>
      <div className="w-40">
        <h3 className="text-display text-sm truncate">{list.name}</h3>
        <p className="text-secondary text-xs">
          {thingCount === 0 ? 'Empty' : thingCount === 1 ? '1 thing' : `${thingCount} things`}
        </p>
      </div>
    </div>
  );
};
