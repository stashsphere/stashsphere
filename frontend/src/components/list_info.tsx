import { useMemo } from "react";
import { List } from "../api/resources";
import { YellowButton } from "./button";
import ImageGrid from "./image_grid";
import { Labeled } from "./labeled";

type ListInfoProps = {
  list: List;
};

export const ListInfo = ({ list }: ListInfoProps) => {
  
  const imageHashes = useMemo(() => {
    return list.things.map(thing => thing.images[0]?.id).filter(e => e !== undefined)
  }, [list])
   
  return (
    <div className="flex flex-col gap-4 flex-start items-start">
      <ImageGrid imageIds={imageHashes}/>
      <a href={`/lists/${list.id}`}>
        <div className="flex flex-col">
          <h2 className="text-display">{list.name}</h2>
        </div>
      </a>
    </div>
  )

  return (
    <div>
      <a href={`/lists/${list.id}`}>
        <h2 className="text-xl text-primary">{list.name}</h2>
      </a>
      <div className="flex flex-wrap">
        <div className="w-full md:w-1/3 p-4">
          <h2 className="font-bold">Things</h2>
          <ul>
            {list.things.map((thing) => (
              <li key={thing.id}>
                <a href={`/things/${thing.id}`}>{thing.name}</a>
              </li>
            ))}
          </ul>
        </div>
        <div className="w-full md:w-1/3 p-4">
          <Labeled label="ID">{list.id}</Labeled>
          <Labeled label="Owner">{list.owner.name}</Labeled>
        </div>
        <div className="w-full md:w-1/3 p-4">
          <a href={`/lists/${list.id}`}><YellowButton>Go to List</YellowButton></a>
        </div>
      </div>
    </div>
  );
};
