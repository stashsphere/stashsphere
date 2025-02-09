import { useMemo } from "react";
import { List } from "../api/resources";
import ImageGrid from "./image_grid";
import { Icon } from "./icon";

type ListInfoProps = {
  list: List;
};

export const ListInfo = ({ list }: ListInfoProps) => {
  
  const imageHashes = useMemo(() => {
    return list.things.map(thing => thing.images[0]?.id).filter(e => e !== undefined)
  }, [list])

   
  return (
    <div className="flex flex-col gap-4 flex-start items-start border border-secondary rounded-md p-1">
      <div className="flex w-80 min-h-60 items-center justify-center">
        <ImageGrid imageIds={imageHashes}/>
      </div>
      <div className="w-80">
      <a href={`/lists/${list.id}`}>
        <h2 className="text-display text-xl mb-2">{list.name}</h2>
        <div className="flex flex-row gap-2">
            <h2 className="text-display"><Icon icon="mdi--user" /> {list.owner.name}</h2>
            <h2 className="text-display"><Icon icon="mdi--animation" /> {list.things.length} things</h2>
        </div>
      </a>
      </div>
    </div>
  )
};
