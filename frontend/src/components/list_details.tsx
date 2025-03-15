import { useContext, useEffect, useMemo, useState } from "react";
import { List } from "../api/resources";
import { AxiosContext } from "../context/axios";
import { getList } from "../api/lists";
import { DangerButton, SecondaryButton } from "./button";
import ThingInfo from "./thing_info";

interface ListDetailsProps {
  id: string;
}

const ListActions = ({ list }: { list: List }) => {
  const colClass = useMemo(() => {
    let actionCount = 0;

    if (list.actions.canEdit) {
      actionCount++;
    }
    if (list.actions.canShare) {
      actionCount++;
    }
    if (list.actions.canDelete) {
      actionCount++;
    }
    return `grid-cols-${actionCount}`
  }, [list])

  return <div className={`grid gap-x-2 ${colClass}`}>
    {list.actions.canEdit &&
      <a href={`/lists/${list.id}/edit`}>
        <SecondaryButton className="w-full">Edit</SecondaryButton>
      </a>}
    {list.actions.canShare &&
      <a href={`/lists/${list.id}/share`}>
        <SecondaryButton className="w-full flex flex-row">Share <div className="rounded bg-secondary-200 text-onprimary mx-1 px-1">{list.shares.length}</div></SecondaryButton>
      </a>}
    {list.actions.canDelete && <a href="#">
      <DangerButton className="w-full">Delete</DangerButton>
    </a>}
  </div>
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
          <ListActions list={list} />
        </div>
        <div className="flex flex-row gap-4 mt-4 flex-wrap justify-center">
          {list.things.map((thing) => (
            <ThingInfo thing={thing} key={thing.id} />
          ))}
        </div>
      </>
    );
  }
};
