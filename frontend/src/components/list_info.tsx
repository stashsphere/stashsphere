import { useMemo } from 'react';
import { List } from '../api/resources';
import ImageGrid from './image_grid';
import { Icon } from './shared';
import { UserNameAndUserId } from './shared/user';

type ListInfoProps = {
  list: List;
  compact?: boolean;
};

export const ListInfo = ({ list, compact = false }: ListInfoProps) => {
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
    <div
      className={`flex flex-col gap-4 flex-start items-start border border-secondary rounded-md p-1 justify-between ${compact ? 'w-60' : 'w-80'}`}
    >
      <div
        className={`flex items-center justify-center ${compact ? 'w-60 min-h-44' : 'w-80 min-h-60'}`}
      >
        <ImageGrid images={images} compact={compact} />
      </div>
      <div className={compact ? 'w-60' : 'w-80'}>
        <a href={`/lists/${list.id}`}>
          <h2 className="text-display text-xl mb-2">{list.name}</h2>
          <div className="flex flex-row gap-2 items-center">
            <UserNameAndUserId
              userId={list.owner.id}
              textColor="text-display"
              imageBorderColor="border-display"
            />
            <h2 className="text-display">
              <Icon icon="mdi--animation" /> {thingCount}
            </h2>
          </div>
        </a>
      </div>
    </div>
  );
};
