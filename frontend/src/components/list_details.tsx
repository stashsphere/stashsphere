import { useContext, useEffect, useState } from "react";
import { List } from "../api/resources";
import { AxiosContext } from "../context/axios";
import { getList } from "../api/lists";
import { DangerButton, SecondaryButton } from "./button";
import ThingInfo from "./thing_info";

interface ListDetailsProps {
  id: string;
}

export const ListDetails = (props: ListDetailsProps) => {
  const [list, setList] = useState<null | List>(null);
  const axiosInstance = useContext(AxiosContext);

  useEffect(() => {
    if (!axiosInstance) {
      return;
    }
    getList(axiosInstance, props.id).then(setList);
  }, [axiosInstance, props.id]);

  if (list === null) {
    return <h1>Loading</h1>;
  } else {
    return (
      <>
        <div className="flex flex-row justify-between mb-4">
          <h1 className="text-2xl text-accent">{list.name}</h1>
          <div className="grid gap-x-2 grid-cols-3">
            <a href={`/lists/${list.id}/edit`}>
              <SecondaryButton className="w-full">Edit</SecondaryButton>
            </a>
            <a href={`/lists/${list.id}/share`}>
              <SecondaryButton className="w-full">Share</SecondaryButton>
            </a>
            <a href="#">
              <DangerButton className="w-full">Delete</DangerButton>
            </a>
          </div>
        </div>
        <div className="flex flex-row gap-4 mt-4 flex-wrap">
        {list.things.map((thing) => (
          <ThingInfo thing={thing} key={thing.id} />
        ))}
        </div>
      </>
    );
  }
};
