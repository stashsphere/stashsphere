import { useMemo } from 'react';
import { List } from '../api/resources';
import ImageGrid from './image_grid';
import { Icon } from './shared';

type ListInfoProps = {
  list: List;
};

export const ListInfo = ({ list }: ListInfoProps) => {
  const images = useMemo(() => {
    return list.things.map((thing) => thing.images[0]).filter((e) => e !== undefined);
  }, [list]);

  const thingCount = useMemo(() => {
    if (list.things.length === 0) {
      return 'No things in this list yet';
    } else if (list.things.length === 1) {
      return '1 thing';
    } else {
      return `${list.things.length} things`;
    }
  }, [list]);

  return (
    <div className="flex flex-col gap-4 flex-start items-start border border-secondary rounded-md p-1 justify-between">
      <div className="flex w-80 min-h-60 items-center justify-center">
        <ImageGrid images={images} />
      </div>
      <div className="w-80">
        <a href={`/lists/${list.id}`}>
          <h2 className="text-display text-xl mb-2">{list.name}</h2>
          <div className="flex flex-row gap-2">
            <h2 className="text-display">
              <Icon icon="mdi--user" /> {list.owner.name}
            </h2>
            <h2 className="text-display">
              <Icon icon="mdi--animation" /> {thingCount}
            </h2>
          </div>
        </a>
      </div>
    </div>
  );
};
