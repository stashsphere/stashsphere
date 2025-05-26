import { useEffect, useState } from 'react';
import { Property } from '../../api/resources';
import { NeutralButton } from '../button';

interface Props {
  properties: Property[];
  onUpdateProperties?: (properties: Property[]) => void;
}

const PropertyEditor: React.FC<Props> = ({ properties, onUpdateProperties }) => {
  const [localProperties, setLocalProperties] = useState<Property[]>([]);

  useEffect(() => {
    setLocalProperties(properties);
  }, [properties]);

  const handleValueChange = (index: number, value: string) => {
    const newProperties = [...localProperties];
    switch (newProperties[index].type) {
      case 'datetime':
        newProperties[index].value = value;
        break;
      case 'string':
        newProperties[index].value = value;
        break;
      case 'float':
        newProperties[index].value = Number(value);
        break;
    }
    setLocalProperties(newProperties);
    if (onUpdateProperties) onUpdateProperties(newProperties);
  };

  const handleUnitChange = (index: number, value: string) => {
    const newProperties = [...localProperties];
    const atIndex = newProperties[index];
    if (atIndex.type === 'float') {
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
      case 'datetime':
        if (localProperties[index].type !== 'datetime') {
          newProperties[index].value = new Date().toISOString();
        }
        newProperties[index].type = value;
        break;
      case 'string':
        if (localProperties[index].type !== 'string') {
          newProperties[index].value = '';
        }
        newProperties[index].type = value;
        break;
      case 'float':
        if (localProperties[index].type !== 'float') {
          newProperties[index].value = 0;
        }
        newProperties[index].type = value;
        break;
      default:
        console.error('Invalid property type');
    }
    setLocalProperties(newProperties);
  };

  const addProperty = (event: React.MouseEvent) => {
    event.preventDefault();
    setLocalProperties([
      ...localProperties,
      { name: '', value: '', type: 'string', unit: undefined },
    ]);
  };

  const deleteProperty = (indexToDelete: number) => {
    const newProperties = localProperties.filter((_, index) => index !== indexToDelete);
    setLocalProperties(newProperties);
    if (onUpdateProperties) onUpdateProperties(newProperties);
  };

  const inputForPropertyType = (prop: Property, index: number) => {
    switch (prop.type) {
      case 'float':
        return (
          <input
            type="number"
            value={prop.value}
            onChange={(e) => handleValueChange(index, e.target.value)}
            className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
            placeholder="Enter number"
          />
        );
      case 'string':
        return (
          <input
            type="text"
            value={prop.value}
            onChange={(e) => handleValueChange(index, e.target.value)}
            className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
            placeholder="Enter text"
          />
        );
      case 'datetime': {
        const formattedDate = new Date(prop.value).toISOString().split('T')[0];
        return (
          <input
            type="date"
            value={formattedDate}
            onChange={(e) => handleValueChange(index, e.target.value)}
            className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
          />
        );
      }
    }
  };

  return (
    <>
      <h2 className="text-xl font-bold mb-4 text-secondary">Properties</h2>
      <div className="overflow-x-auto">
        <div className="space-y-3">
          {localProperties.map((property, index) => (
            <div
              key={index}
              className="grid grid-cols-1 sm:grid-cols-5 gap-2 p-3 border border-gray-200 rounded-sm"
            >
              <div className="sm:col-span-1">
                <label className="block text-xs font-medium text-display mb-1">Name</label>
                <input
                  type="text"
                  value={property.name}
                  onChange={(e) => handleNameChange(index, e.target.value)}
                  className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
                  placeholder="Property name"
                />
              </div>

              <div className="sm:col-span-2">
                <label className="block text-xs font-medium text-display mb-1">Value</label>
                {inputForPropertyType(property, index)}
              </div>

              <div className="sm:col-span-1">
                <label className="block text-xs font-medium text-display mb-1">Type</label>
                <select
                  onChange={(e) => handleTypeChange(index, e.target.value)}
                  value={property.type}
                  className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
                >
                  <option value="string">Text</option>
                  <option value="float">Number</option>
                  <option value="datetime">Date</option>
                </select>
                {property.type === 'float' && (
                  <input
                    type="text"
                    value={property.unit || ''}
                    onChange={(e) => handleUnitChange(index, e.target.value)}
                    placeholder="Unit"
                    className="w-full mt-1 text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1 text-xs"
                  />
                )}
              </div>

              <div className="sm:col-span-1 flex items-end">
                <button
                  onClick={() => deleteProperty(index)}
                  className="w-full sm:w-auto px-3 py-1 text-sm text-red-600 hover:text-red-800 hover:bg-red-50 rounded-sm transition-colors"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
        <div className="mt-4">
          <NeutralButton onClick={addProperty}>Add Property</NeutralButton>
        </div>
      </div>
    </>
  );
};

export default PropertyEditor;
