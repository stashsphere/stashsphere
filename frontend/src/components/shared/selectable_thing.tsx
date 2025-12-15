import { useCallback } from 'react';
import { Thing } from '../../api/resources';
import { Icon, ImageComponent } from '.';

type SelectableThingProps = {
  thing: Thing;
  selected: boolean;
  onSelect: (thingId: string, selected: boolean) => void;
};

export const SelectableThing = ({ thing, selected, onSelect }: SelectableThingProps) => {
  const handleClick = useCallback(() => {
    onSelect(thing.id, !selected);
  }, [thing.id, selected, onSelect]);
  const firstImage = thing.images[0];
  const firstImageContent = firstImage ? (
    <ImageComponent
      image={firstImage}
      defaultWidth={256}
      className="object-contain h-full w-full"
    />
  ) : (
    <span>
      <Icon icon="mdi--image-off-outline" />
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
        <div className="absolute top-2 right-2">
          <Icon icon="mdi--check-circle" className="text-accent" size="small" />
        </div>
      )}
      <div className="flex w-48 h-48 items-center justify-center bg-brand-900 p-2 rounded-md">
        {firstImageContent}
      </div>
      <h3 className="text-display text-sm w-48 truncate">{thing.name}</h3>
    </div>
  );
};
