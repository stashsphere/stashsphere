import { useContext, useEffect, useRef, useState } from 'react';
import { Property } from '../../api/resources';
import { AxiosContext } from '../../context/axios';
import { getAutoComplete } from '../../api/search';

interface PropertyRowProps {
  property: Property;
  onChange: (property: Property) => void;
  onDelete: () => void;
}

const PropertyRow: React.FC<PropertyRowProps> = ({ property, onChange, onDelete }) => {
  const axiosInstance = useContext(AxiosContext);
  const [nameSuggestions, setNameSuggestions] = useState<string[]>([]);
  const [valueSuggestions, setValueSuggestions] = useState<string[]>([]);
  const nameDebounceRef = useRef<NodeJS.Timeout | null>(null);
  const valueDebounceRef = useRef<NodeJS.Timeout | null>(null);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (nameDebounceRef.current) clearTimeout(nameDebounceRef.current);
      if (valueDebounceRef.current) clearTimeout(valueDebounceRef.current);
    };
  }, []);

  const fetchNameSuggestions = (name: string) => {
    if (!axiosInstance) return;
    if (nameDebounceRef.current) clearTimeout(nameDebounceRef.current);

    nameDebounceRef.current = setTimeout(() => {
      getAutoComplete(axiosInstance, name, null)
        .then((result) => {
          if (result.completionType === 'name') {
            setNameSuggestions(result.values);
          }
        })
        .catch((err) => {
          console.error('Autocomplete error:', err);
        });
    }, 300);
  };

  const fetchValueSuggestions = (name: string, value: string) => {
    if (!axiosInstance) return;
    if (valueDebounceRef.current) clearTimeout(valueDebounceRef.current);

    valueDebounceRef.current = setTimeout(() => {
      getAutoComplete(axiosInstance, name, value)
        .then((result) => {
          if (result.completionType === 'value') {
            setValueSuggestions(result.values);
          }
        })
        .catch((err) => {
          console.error('Autocomplete error:', err);
        });
    }, 300);
  };

  const handleNameChange = (name: string) => {
    switch (property.type) {
      case 'string':
        onChange({ type: 'string', name, value: property.value, unit: undefined });
        break;
      case 'float':
        onChange({ type: 'float', name, value: property.value, unit: property.unit });
        break;
      case 'datetime':
        onChange({ type: 'datetime', name, value: property.value, unit: undefined });
        break;
    }
    fetchNameSuggestions(name);
  };

  const handleValueChange = (value: string) => {
    console.log(value);
    switch (property.type) {
      case 'float':
        onChange({ type: 'float', name: property.name, value: Number(value), unit: property.unit });
        break;
      case 'string':
        onChange({ type: 'string', name: property.name, value, unit: undefined });
        if (property.name) {
          fetchValueSuggestions(property.name, value);
        }
        break;
      case 'datetime':
        onChange({ type: 'datetime', name: property.name, value, unit: undefined });
        break;
    }
  };

  const handleTypeChange = (type: string) => {
    switch (type) {
      case 'datetime': {
        const value =
          property.type !== 'datetime' ? new Date().toISOString() : (property.value as string);
        onChange({ type: 'datetime', name: property.name, value, unit: undefined });
        break;
      }
      case 'string': {
        const value = property.type !== 'string' ? '' : (property.value as string);
        onChange({ type: 'string', name: property.name, value, unit: undefined });
        break;
      }
      case 'float': {
        const value = property.type !== 'float' ? 0 : (property.value as number);
        const unit = property.type === 'float' ? property.unit : '';
        onChange({ type: 'float', name: property.name, value, unit });
        break;
      }
      default:
        console.error('Invalid property type');
    }
  };

  const handleUnitChange = (unit: string) => {
    if (property.type === 'float') {
      onChange({ type: 'float', name: property.name, value: property.value, unit });
    }
  };

  // Render value input based on type
  const renderValueInput = () => {
    switch (property.type) {
      case 'float':
        return (
          <input
            type="number"
            value={property.value}
            onChange={(e) => handleValueChange(e.target.value)}
            className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
            placeholder="Enter number"
          />
        );
      case 'string':
        return (
          <>
            <input
              type="text"
              value={property.value}
              onChange={(e) => handleValueChange(e.target.value)}
              className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
              placeholder="Enter text"
              list="value-suggestions"
            />
            <datalist id="value-suggestions">
              {valueSuggestions.map((suggestion, i) => (
                <option key={i} value={suggestion} />
              ))}
            </datalist>
          </>
        );
      case 'datetime': {
        const formattedDate = new Date(property.value).toISOString().split('T')[0];
        return (
          <input
            type="date"
            value={formattedDate}
            onChange={(e) => {
              // Convert YYYY-MM-DD to ISO timestamp by appending time and converting
              const isoString = e.target.value + 'T00:00:00.000Z';
              handleValueChange(isoString);
            }}
            className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
          />
        );
      }
    }
  };

  return (
    <div className="grid grid-cols-1 sm:grid-cols-5 gap-2 p-3 border border-gray-200 rounded-sm items-start">
      <div className="sm:col-span-1">
        <label className="block text-xs font-medium text-display mb-1">Name</label>
        <input
          type="text"
          value={property.name}
          onChange={(e) => handleNameChange(e.target.value)}
          className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
          placeholder="Property name"
          list="name-suggestions"
        />
        <datalist id="name-suggestions">
          {nameSuggestions.map((suggestion, i) => (
            <option key={i} value={suggestion} />
          ))}
        </datalist>
      </div>

      <div className="sm:col-span-2">
        <label className="block text-xs font-medium text-display mb-1">Value</label>
        {renderValueInput()}
      </div>

      <div className="sm:col-span-1">
        <label className="block text-xs font-medium text-display mb-1">Type</label>
        <select
          onChange={(e) => handleTypeChange(e.target.value)}
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
            onChange={(e) => handleUnitChange(e.target.value)}
            placeholder="Unit"
            className="w-full mt-1 text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1 text-xs"
          />
        )}
      </div>

      <div className="sm:col-span-1">
        <label className="block text-xs font-medium text-display mb-1 invisible">Actions</label>
        <button
          onClick={onDelete}
          className="w-full sm:w-auto px-3 py-1 text-sm text-red-600 hover:text-red-800 hover:bg-red-50 rounded-sm transition-colors"
        >
          Delete
        </button>
      </div>
    </div>
  );
};

export default PropertyRow;
