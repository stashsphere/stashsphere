import React from "react";
import { Property } from "../api/resources";
import { Icon } from "./icon";

interface PropertyListProps {
  properties: Property[];
}

const propertyDisplay = (property: Property) => {
  switch (property.type) {
    case "string":
      return (
        <>
          <Icon icon="mdi--format-text" />
          {property.value}
        </>
      );
    case "datetime":
      return (
        <>
          <Icon icon="mdi--date-range" />
          {property.value.toString()}
        </>
      );
    case "float":
      return (
        <>
          <Icon icon="mdi--hashtag" />
          {property.value}
        </>
      );
  }
};

const PropertyList: React.FC<PropertyListProps> = ({ properties }) => {
  if (properties.length === 0) {
    return null;
  } else {
    return (
      <table className="min-w-full table-auto">
        <thead>
          <tr className="bg-gray-200">
            <th className="px-4 py-2 text-left font-medium text-gray-700 uppercase tracking-wide">
              Property Name
            </th>
            <th className="px-4 py-2 text-left font-medium text-gray-700 uppercase tracking-wide">
              Value
            </th>
            <th className="px-4 py-2 text-left font-medium text-gray-700 uppercase tracking-wide">
              Unit
            </th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-300">
          {properties.map((property, index) => (
            <tr
              key={index}
              className="hover:bg-gray-100 transition duration-300 ease-in-out"
            >
              <td className="px-4 py-2 whitespace-nowrap">{property.name}</td>
              <td className="px-4 py-2 whitespace-nowrap">
                {propertyDisplay(property)}
              </td>
              <td className="px-4 py-2 whitespace-nowrap">
                {property.type === "float" ? property.unit : ""}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  }
};

export default PropertyList;
