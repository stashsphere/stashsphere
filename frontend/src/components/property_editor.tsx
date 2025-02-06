import { useEffect, useState } from "react";
import { Property } from "../api/resources";
import { NeutralButton } from "./button";

interface Props {
  properties: Property[];
  onUpdateProperties?: (properties: Property[]) => void;
}

const PropertyEditor: React.FC<Props> = ({
  properties,
  onUpdateProperties,
}) => {
  const [localProperties, setLocalProperties] = useState<Property[]>([]);

  useEffect(() => {
    setLocalProperties(properties);
  }, [properties]);

  const handleValueChange = (index: number, value: string) => {
    const newProperties = [...localProperties];
    switch (newProperties[index].type) {
      case "datetime":
        newProperties[index].value = value;
        break;
      case "string":
        newProperties[index].value = value;
        break;
      case "float":
        newProperties[index].value = Number(value);
        break;
    }
    setLocalProperties(newProperties);
    if (onUpdateProperties) onUpdateProperties(newProperties);
  };

  const handleUnitChange = (index: number, value: string) => {
    const newProperties = [...localProperties];
    const atIndex = newProperties[index];
    if (atIndex.type === "float") {
      atIndex.unit = value;
    } else {
      return;
    }
    setLocalProperties(newProperties);
    if (onUpdateProperties) onUpdateProperties(newProperties);
  };

  const handleNameChange = (index: number, value: string) => {
    const newProperties = [...localProperties];
    newProperties[index].name = value;
    setLocalProperties(newProperties);
    if (onUpdateProperties) onUpdateProperties(newProperties);
  };

  const handleTypeChange = (index: number, value: string) => {
    const newProperties = [...localProperties];
    switch (value) {
      case "datetime":
        if (localProperties[index].type !== "datetime") {
          newProperties[index].value = new Date().toISOString();
        }
        newProperties[index].type = value;
        break;
      case "string":
        if (localProperties[index].type !== "string") {
          newProperties[index].value = "";
        }
        newProperties[index].type = value;
        break;
      case "float":
        if (localProperties[index].type !== "float") {
          newProperties[index].value = 0;
        }
        newProperties[index].type = value;
        break;
      default:
        console.error("Invalid property type");
    }
    setLocalProperties(newProperties);
  };

  const addProperty = (event: React.MouseEvent) => {
    event.preventDefault();
    setLocalProperties([
      ...localProperties,
      { name: "", value: "", type: "string", unit: undefined },
    ]);
  };

  const deleteProperty = (indexToDelete: number) => {
    const newProperties = localProperties.filter(
      (_, index) => index !== indexToDelete
    );
    setLocalProperties(newProperties);
    if (onUpdateProperties) onUpdateProperties(newProperties);
  };

  const inputForPropertyType = (prop: Property, index: number) => {
    switch (prop.type) {
      case "float":
        return (
          <input
            type="number"
            value={prop.value}
            onChange={(e) => handleValueChange(index, e.target.value)}
            className="mt-1 block w-full text-display border border-secondary shadow-sm focus:border-secondary"
          />
        );
      case "string":
        return (
          <input
            type="string"
            value={prop.value}
            onChange={(e) => handleValueChange(index, e.target.value)}
            className="mt-1 block w-full text-display border border-secondary shadow-sm focus:border-secondary"
          />
        );
      case "datetime": {
        const formattedDate = new Date(prop.value).toISOString().split("T")[0];
        return (
          <input
            type="date"
            value={formattedDate}
            onChange={(e) => handleValueChange(index, e.target.value)}
            className="mt-1 block w-full text-display border border-secondary shadow-sm focus:border-secondary"
          />
        );
      }
    }
  };

  return (
    <>
      <h2 className="text-xl font-bold mb-4 text-secondary">Properties</h2>
      <div className="p-4">
        <table className="min-w-full divide-y divide-gray-200">
          <thead>
            <tr>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-display uppercase tracking-wider"
              >
                Name
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-display uppercase tracking-wider"
              >
                Value
              </th>
              <th scope="col" className="relative px-6 py-3"></th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {localProperties.map((property, index) => (
              <tr key={index}>
                <td className="px-2 py-2 whitespace-nowrap">
                  <input
                    type="text"
                    value={property.name}
                    onChange={(e) => handleNameChange(index, e.target.value)}
                    className="mt-1 block w-full text-display border border-secondary shadow-sm focus:border-secondary"
                  />
                </td>
                <td className="px-2 py-2 whitespace-nowrap">
                  {inputForPropertyType(property, index)}
                </td>
                <td className="px-4 text-display">
                  <input
                    type="text"
                    value={property.type === "float" ? property.unit : ""}
                    onChange={(e) => handleUnitChange(index, e.target.value)}
                    disabled={property.type !== "float"}
                  />
                </td>
                <td className="px-2 py-4 whitespace-nowrap text-display">
                  <select
                    onChange={(e) => handleTypeChange(index, e.target.value)}
                    value={property.type}
                  >
                    <option value="string">String</option>
                    <option value="float">Number</option>
                    <option value="datetime">Datetime</option>
                  </select>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <button
                    onClick={() => deleteProperty(index)}
                    className="text-danger-500 hover:text-danger-600 hover:border hover:border-danger hover:mx-0 mx-px px-px"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        <NeutralButton onClick={addProperty}>Add Property</NeutralButton>
      </div>
    </>
  );
};

export default PropertyEditor;
