import { useContext, useState } from "react";
import { Property, Thing } from "../api/resources";
import { ConfigContext } from "../context/config";
import { Icon } from "./icon";

const ThingPropertyList = ({ properties }: { properties: Property[] }) => {
  const [collapsed, setCollapsed] = useState(true);

  const collapsedDisplay = 3;

  const filteredProperties = collapsed
    ? properties.slice(0, collapsedDisplay)
    : properties;
  const hasMore = properties.length > collapsedDisplay;

  const formatValue = (value: Property["value"]) => {
    if (typeof value === "string") {
      return `${value}`;
    } else if (typeof value === "number" && !isNaN(value)) {
      return `${value}`;
    } else {
      return String(value);
    }
  };

  return (
    <>
      <ul className="list-inside text-onneutral text-sm">
        {filteredProperties.map((property) => (
          <li key={property.name} className="bg-neutral-primary rounded-lg p-1" title={property.value as string}>
            <b>{property.name}</b>: {formatValue(property.value)}
          </li>
        ))}
      </ul>
      {hasMore ? (
        <button
          onClick={() => {
            setCollapsed(!collapsed);
          }}
        >
          {collapsed ? "Show more" : "Show less"}
        </button>
      ) : null}
    </>
  );
};

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
          <ThingPropertyList properties={thing.properties} />
        </a>
      </div>
    </div>
  )
};

export default ThingInfo;
