import { useContext, useEffect, useId, useRef, useState } from 'react';
import { Property, PropertyNameSuggestion } from '../../api/resources';
import { AxiosContext } from '../../context/axios';
import { getAutoComplete } from '../../api/search';
import { Toggle } from '../shared/toggle';

interface PropertyRowProps {
  property: Property;
  onChange: (property: Property) => void;
  onDelete: () => void;
}

const PropertyRow: React.FC<PropertyRowProps> = ({ property, onChange, onDelete }) => {
  const axiosInstance = useContext(AxiosContext);
  const [nameSuggestions, setNameSuggestions] = useState<PropertyNameSuggestion[]>([]);
  const [valueSuggestions, setValueSuggestions] = useState<string[]>([]);
  const [unitSuggestions, setUnitSuggestions] = useState<string[]>([]);
  const nameDebounceRef = useRef<NodeJS.Timeout | null>(null);
  const valueDebounceRef = useRef<NodeJS.Timeout | null>(null);
  const rowId = useId();

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
            setNameSuggestions(result.suggestions);
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

  const applyNameSuggestion = (suggestion: PropertyNameSuggestion, name: string) => {
    setUnitSuggestions(suggestion.units);

    switch (suggestion.type) {
      case 'float': {
        let value: number;
        if (property.type === 'string') {
          const parsed = parseFloat(property.value);
          value = isNaN(parsed) ? 0 : parsed;
        } else if (property.type === 'float') {
          value = property.value;
        } else {
          value = 0;
        }
        onChange({ type: 'float', name, value, unit: suggestion.units[0] ?? '' });
        break;
      }
      case 'string': {
        let value: string;
        if (property.type === 'float') {
          value = String(property.value);
        } else if (property.type === 'datetime') {
          value = '';
        } else if (property.type === 'boolean') {
          value = String(property.value);
        } else {
          value = property.value;
        }
        onChange({ type: 'string', name, value, unit: undefined });
        break;
      }
      case 'datetime': {
        onChange({ type: 'datetime', name, value: new Date().toISOString(), unit: undefined });
        break;
      }
    }
  };

  const handleNameChange = (name: string) => {
    const suggestion = nameSuggestions.find((s) => s.name === name);
    if (suggestion) {
      applyNameSuggestion(suggestion, name);
    } else {
      setUnitSuggestions([]);
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
        case 'boolean':
          onChange({ type: 'boolean', name, value: property.value });
          break;
      }
    }
    fetchNameSuggestions(name);
  };

  const handleValueChange = (value: string) => {
    switch (property.type) {
      case 'float':
        onChange({ type: 'float', name: property.name, value: Number(value), unit: property.unit });
        break;
      case 'string':
        onChange({ type: 'string', name: property.name, value, unit: undefined });
        if (property.name) fetchValueSuggestions(property.name, value);
        break;
      case 'datetime':
        onChange({ type: 'datetime', name: property.name, value, unit: undefined });
        break;
    }
  };

  const handleBooleanChange = (value: boolean) => {
    onChange({ type: 'boolean', name: property.name, value });
  };

  const handleTypeChange = (type: string) => {
    switch (type) {
      case 'datetime': {
        onChange({
          type: 'datetime',
          name: property.name,
          value: new Date().toISOString(),
          unit: undefined,
        });
        break;
      }
      case 'string': {
        let value: string;
        if (property.type === 'float') {
          value = String(property.value);
        } else if (property.type === 'datetime') {
          value = '';
        } else if (property.type === 'boolean') {
          value = String(property.value);
        } else {
          value = property.value;
        }
        onChange({ type: 'string', name: property.name, value, unit: undefined });
        break;
      }
      case 'float': {
        let value: number;
        if (property.type === 'string') {
          const parsed = parseFloat(property.value);
          value = isNaN(parsed) ? 0 : parsed;
        } else if (property.type === 'datetime') {
          value = 0;
        } else if (property.type == 'boolean') {
          value = 0;
        } else {
          value = property.value;
        }
        onChange({ type: 'float', name: property.name, value, unit: '' });
        break;
      }
      case 'boolean': {
        onChange({ type: 'boolean', name: property.name, value: false });
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

  const renderValueInput = () => {
    switch (property.type) {
      case 'float':
        return (
          <input
            type="number"
            value={property.value}
            onFocus={(e) => e.target.select()}
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
              list={`${rowId}-value-suggestions`}
            />
            <datalist id={`${rowId}-value-suggestions`}>
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
            onChange={(e) => handleValueChange(e.target.value + 'T00:00:00.000Z')}
            className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
          />
        );
      }
      case 'boolean': {
        return (
          <div className="flex my-auto gap-5">
            <Toggle
              value={property.value}
              onChange={(newValue) => handleBooleanChange(newValue)}
              children={undefined}
            />
            <span className="text-display">{property.value ? 'true' : 'false'}</span>
          </div>
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
          list={`${rowId}-name-suggestions`}
        />
        <datalist id={`${rowId}-name-suggestions`}>
          {nameSuggestions.map((suggestion, i) => (
            <option key={i} value={suggestion.name} />
          ))}
        </datalist>
      </div>

      <div className="sm:col-span-2 h-full flex flex-col justify-between">
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
          <option value="boolean">Bool</option>
        </select>
        {property.type === 'float' && (
          <>
            <input
              type="text"
              value={property.unit || ''}
              onChange={(e) => handleUnitChange(e.target.value)}
              placeholder="Unit"
              className="w-full mt-1 text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1 text-xs"
              list={`${rowId}-unit-suggestions`}
            />
            {unitSuggestions.length > 0 && (
              <datalist id={`${rowId}-unit-suggestions`}>
                {unitSuggestions.map((unit, i) => (
                  <option key={i} value={unit} />
                ))}
              </datalist>
            )}
          </>
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
