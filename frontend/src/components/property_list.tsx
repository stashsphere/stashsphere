import { useState } from "react";
import { Property } from "../api/resources";

import PropertyComponent from "./property";

const PropertyList = ({ properties, collapsable }: { properties: Property[], collapsable?: boolean }) => {
  const [collapsed, setCollapsed] = useState(true);

  const collapsedDisplay = collapsable ? 3 : properties.length;

  const filteredProperties = collapsed
    ? properties.slice(0, collapsedDisplay)
    : properties;
  const hasMore = properties.length > collapsedDisplay;

  return (
    <>
      <ul className="list-inside text-onneutral text-sm">
        {filteredProperties.map((property) => (
          <li key={property.name} className="bg-neutral-primary rounded-lg p-1">
            <PropertyComponent property={property} />
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

export default PropertyList;
