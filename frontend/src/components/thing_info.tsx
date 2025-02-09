import { useContext } from "react";
import { Thing } from "../api/resources";
import { ConfigContext } from "../context/config";
import { Icon } from "./icon";
import PropertyList from "./property_list";

type ThingInfoProps = {
  thing: Thing;
};

const ThingInfo = ({ thing }: ThingInfoProps) => {
  const config = useContext(ConfigContext);

  const firstImageId = thing.images[0]?.id;
  const firstImageContent = firstImageId ? (
    <img src={`${config.apiHost}/api/images/${firstImageId}`} alt="Image" className="object-contain h-full w-full" />
  ) : (
    <span>
      <Icon height="100%" icon="mdi--image-off-outline" />
    </span>
  );


  return (
    <div className="flex flex-col gap-4 flex-start items-start">
      <div className="flex w-80 h-80 items-center justify-center bg-brand-900 p-2 rounded-md">
        {firstImageContent}
      </div>
      <div className="w-80">
        <a href={`/things/${thing.id}`}>
          <div className="flex flex-col">
            <h2 className="text-display">{thing.name}</h2>
            <h3 className="text-display text-sm"><Icon icon="mdi--user" /> {thing.owner.name}</h3>
          </div>
        </a>
        <PropertyList properties={thing.properties} collapsable={true} />
      </div>
    </div>
  )
};

export default ThingInfo;
