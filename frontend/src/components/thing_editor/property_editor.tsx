import { useEffect, useState } from 'react';
import { Property } from '../../api/resources';
import { NeutralButton } from '../shared';
import PropertyRow from './property_row';

interface Props {
  properties: Property[];
  onUpdateProperties?: (properties: Property[]) => void;
}

const PropertyEditor: React.FC<Props> = ({ properties, onUpdateProperties }) => {
  const [localProperties, setLocalProperties] = useState<Property[]>([]);

  useEffect(() => {
    setLocalProperties(properties);
  }, [properties]);

  const handlePropertyChange = (index: number, property: Property) => {
    const newProperties = [...localProperties];
    newProperties[index] = property;
    setLocalProperties(newProperties);
    if (onUpdateProperties) onUpdateProperties(newProperties);
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

  return (
    <>
      <h2 className="text-xl font-bold mb-4 text-secondary">Properties</h2>
      <div className="overflow-x-auto">
        <div className="space-y-3">
          {localProperties.map((property, index) => (
            <PropertyRow
              key={index}
              property={property}
              onChange={(prop) => handlePropertyChange(index, prop)}
              onDelete={() => deleteProperty(index)}
            />
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
