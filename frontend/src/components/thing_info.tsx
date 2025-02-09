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
    <div className="flex flex-col gap-4 flex-start items-start border border-secondary rounded-md p-1">
      <div className="flex w-80 h-80 items-center justify-center bg-brand-900 p-2 rounded-md">
        {firstImageContent}
      </div>
      <div className="w-80">
        <a href={`/things/${thing.id}`}>
          <h2 className="text-display text-xl mb-2">{thing.name}</h2>
          <div className="flex flex-row gap-2">
            <h2 className="text-display"><Icon icon="mdi--user" /> {thing.owner.name}</h2>
            <h2 className="text-display"><Icon icon="mdi--animation" /> {thing.quantity} {thing.quantityUnit}</h2>
          </div>
        </a>
        <PropertyList properties={thing.properties} collapsable={true} keyWidth="8rem"/>
      </div>
    </div>
  )
};

export default ThingInfo;
