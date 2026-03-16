import { useContext, useEffect, useId, useRef, useState } from 'react';
import { Property, PropertyNameSuggestion } from '../../../api/resources';
import { AxiosContext } from '../../../context/axios';
import { getAutoComplete } from '../../../api/search';
import { Toggle } from '../../shared/toggle';
import { EditorRow } from './editor_row';
import { Icon } from '../../shared';

interface PropertyRowProps {
  onSave: (property: Property) => void;
  onClear: () => void;
  initialKey: string | undefined;
  initialValue: Property | undefined;
}

const CustomEditor = ({ onSave, onClear, initialKey, initialValue }: PropertyRowProps) => {
  const isEditing = !!initialValue;
  const axiosInstance = useContext(AxiosContext);
  const [nameSuggestions, setNameSuggestions] = useState<PropertyNameSuggestion[]>([]);
  const [valueSuggestions, setValueSuggestions] = useState<string[]>([]);
  const [unitSuggestions, setUnitSuggestions] = useState<string[]>([]);
  const nameDebounceRef = useRef<NodeJS.Timeout | null>(null);
  const valueDebounceRef = useRef<NodeJS.Timeout | null>(null);
  const [localProperty, setLocalProperty] = useState<Property>(
    initialValue || { name: initialKey || '', value: '', type: 'string', unit: undefined }
  );
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
        if (localProperty.type === 'string') {
          const parsed = parseFloat(localProperty.value);
          value = isNaN(parsed) ? 0 : parsed;
        } else if (localProperty.type === 'float') {
          value = localProperty.value;
        } else {
          value = 0;
        }
        setLocalProperty({ type: 'float', name, value, unit: suggestion.units[0] ?? '' });
        break;
      }
      case 'string': {
        let value: string;
        if (localProperty.type === 'float') {
          value = String(localProperty.value);
        } else if (localProperty.type === 'datetime') {
          value = '';
        } else if (localProperty.type === 'boolean') {
          value = String(localProperty.value);
        } else {
          value = localProperty.value;
        }
        setLocalProperty({ type: 'string', name, value, unit: undefined });
        break;
      }
      case 'datetime': {
        setLocalProperty({
          type: 'datetime',
          name,
          value: new Date().toISOString(),
          unit: undefined,
        });
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
      switch (localProperty.type) {
        case 'string':
          setLocalProperty({ type: 'string', name, value: localProperty.value, unit: undefined });
          break;
        case 'float':
          setLocalProperty({
            type: 'float',
            name,
            value: localProperty.value,
            unit: localProperty.unit,
          });
          break;
        case 'datetime':
          setLocalProperty({ type: 'datetime', name, value: localProperty.value, unit: undefined });
          break;
        case 'boolean':
          setLocalProperty({
            type: 'boolean',
            name,
            value: localProperty.value,
            unit: undefined,
          });
          break;
      }
    }
    fetchNameSuggestions(name);
  };

  const handleValueChange = (value: string) => {
    switch (localProperty.type) {
      case 'float':
        setLocalProperty({
          type: 'float',
          name: localProperty.name,
          value: Number(value),
          unit: localProperty.unit,
        });
        break;
      case 'string':
        setLocalProperty({ type: 'string', name: localProperty.name, value, unit: undefined });
        if (localProperty.name) fetchValueSuggestions(localProperty.name, value);
        break;
      case 'datetime':
        setLocalProperty({ type: 'datetime', name: localProperty.name, value, unit: undefined });
        break;
    }
  };

  const handleBooleanChange = (value: boolean) => {
    setLocalProperty({
      type: 'boolean',
      name: localProperty.name,
      value,
      unit: undefined,
    });
  };

  const handleTypeChange = (type: string) => {
    switch (type) {
      case 'datetime': {
        setLocalProperty({
          type: 'datetime',
          name: localProperty.name,
          value: new Date().toISOString(),
          unit: undefined,
        });
        break;
      }
      case 'string': {
        let value: string;
        if (localProperty.type === 'float') {
          value = String(localProperty.value);
        } else if (localProperty.type === 'datetime') {
          value = '';
        } else if (localProperty.type === 'boolean') {
          value = String(localProperty.value);
        } else {
          value = localProperty.value;
        }
        setLocalProperty({ type: 'string', name: localProperty.name, value, unit: undefined });
        break;
      }
      case 'float': {
        let value: number;
        if (localProperty.type === 'string') {
          const parsed = parseFloat(localProperty.value);
          value = isNaN(parsed) ? 0 : parsed;
        } else if (localProperty.type === 'datetime') {
          value = 0;
        } else if (localProperty.type == 'boolean') {
          value = 0;
        } else {
          value = localProperty.value;
        }
        setLocalProperty({ type: 'float', name: localProperty.name, value, unit: '' });
        break;
      }
      case 'boolean': {
        setLocalProperty({
          type: 'boolean',
          name: localProperty.name,
          value: false,
          unit: undefined,
        });
        break;
      }
      default:
        console.error('Invalid property type');
    }
  };

  const handleUnitChange = (unit: string) => {
    if (localProperty.type === 'float') {
      setLocalProperty({
        type: 'float',
        name: localProperty.name,
        value: localProperty.value,
        unit,
      });
    }
  };

  const renderValueInput = () => {
    switch (localProperty.type) {
      case 'float':
        return (
          <input
            type="number"
            value={localProperty.value}
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
              value={localProperty.value}
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
        const formattedDate = new Date(localProperty.value).toISOString().split('T')[0];
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
              value={localProperty.value}
              onChange={(newValue) => handleBooleanChange(newValue)}
              children={undefined}
            />
            <span className="text-display">{localProperty.value ? 'true' : 'false'}</span>
          </div>
        );
      }
    }
  };

  return (
    <EditorRow
      onSubmit={() => onSave(localProperty)}
      isEditing={isEditing}
      disabled={false}
      error={null}
    >
      <div
        className={`grid gap-x-2 gap-y-1 items-end ${localProperty.type === 'float' ? 'grid-cols-[1fr_2fr_auto_auto_auto]' : 'grid-cols-[1fr_2fr_auto_auto]'}`}
      >
        <label className="text-xs font-medium text-display">Name</label>
        <label className="text-xs font-medium text-display">Value</label>
        <label className="text-xs font-medium text-display">Type</label>
        {localProperty.type === 'float' && (
          <label className="text-xs font-medium text-display">Unit</label>
        )}
        <span />

        <div>
          <input
            type="text"
            value={localProperty.name}
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

        <div>{renderValueInput()}</div>

        <select
          onChange={(e) => handleTypeChange(e.target.value)}
          value={localProperty.type}
          className="text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
        >
          <option value="string">Text</option>
          <option value="float">Number</option>
          <option value="datetime">Date</option>
          <option value="boolean">Bool</option>
        </select>

        {localProperty.type === 'float' && (
          <div>
            <input
              type="text"
              value={localProperty.unit || ''}
              onChange={(e) => handleUnitChange(e.target.value)}
              placeholder="Unit"
              className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
              list={`${rowId}-unit-suggestions`}
            />
            {unitSuggestions.length > 0 && (
              <datalist id={`${rowId}-unit-suggestions`}>
                {unitSuggestions.map((unit, i) => (
                  <option key={i} value={unit} />
                ))}
              </datalist>
            )}
          </div>
        )}

        <button
          type="button"
          onClick={onClear}
          className="text-display-light hover:text-accent transition-colors px-1 py-1"
        >
          <Icon icon="mdi--clear" />
        </button>
      </div>
    </EditorRow>
  );
};

export default CustomEditor;
