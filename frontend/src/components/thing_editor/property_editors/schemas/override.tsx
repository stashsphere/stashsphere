import { useState } from 'react';
import { EditorProps } from '../editor';
import { EditorRow } from '../editor_row';
import { Toggle } from '../../../shared/toggle';

export const OverrideEditor = ({
  schemaKey,
  schema,
  onSave,
  initialValue,
  isEditing,
  baseType,
}: EditorProps & {
  isEditing: boolean;
  baseType: 'string' | 'number' | 'boolean';
}) => {
  const [stringValue, setStringValue] = useState(
    initialValue && (initialValue.type === 'string' || initialValue.type === 'datetime')
      ? initialValue.value
      : ''
  );
  const [numberValue, setNumberValue] = useState(
    initialValue?.type === 'float' ? String(initialValue.value) : ''
  );
  const [boolValue, setBoolValue] = useState(
    initialValue?.type === 'boolean' ? initialValue.value : false
  );

  const unitDef = schema.schema.properties.unit;
  const unitOptions = unitDef && Array.isArray(unitDef.enum) ? (unitDef.enum as string[]) : [];
  const [unit, setUnit] = useState(initialValue?.unit ?? unitOptions[0] ?? '');

  const handleSubmit = () => {
    if (baseType === 'string') {
      if (!stringValue.trim()) return;
      onSave({ type: 'string', name: schemaKey, value: stringValue.trim(), unit: undefined });
      setStringValue('');
    } else if (baseType === 'number') {
      const num = parseFloat(numberValue);
      if (isNaN(num)) return;
      onSave({ type: 'float', name: schemaKey, value: num, unit: unit || 'unit' });
      setNumberValue('');
    } else {
      onSave({ type: 'boolean', name: schemaKey, value: boolValue, unit: undefined });
      setBoolValue(false);
    }
  };

  if (baseType === 'boolean') {
    return (
      <EditorRow error={null} disabled={false} onSubmit={handleSubmit} isEditing={isEditing}>
        <div className="flex-1">
          <label className="block text-xs font-medium text-display mb-1">Value</label>
          <div className="flex items-center gap-3 rounded-sm border border-secondary shadow-xs px-3 py-1.5">
            <Toggle
              onChange={(v) => {
                setBoolValue(v);
              }}
              value={boolValue}
            />
            <span className="text-sm text-display">{boolValue ? 'true' : 'false'}</span>
          </div>
        </div>
      </EditorRow>
    );
  }

  if (baseType === 'number') {
    return (
      <EditorRow
        error={null}
        disabled={!numberValue || isNaN(parseFloat(numberValue))}
        onSubmit={handleSubmit}
        isEditing={isEditing}
      >
        <div className="flex-1 min-w-[180px]">
          <label className="block text-xs font-medium text-display mb-1">Value</label>
          <input
            type="number"
            value={numberValue}
            onChange={(e) => setNumberValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleSubmit();
            }}
            placeholder={`Enter ${schemaKey}...`}
            step="any"
            className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
          />
        </div>
        <div className="flex-1 min-w-[180px]">
          <label className="block text-xs font-medium text-display mb-1">Unit</label>
          <input
            type="text"
            value={unit}
            onChange={(e) => setUnit(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleSubmit();
            }}
            placeholder={`Enter unit...`}
            className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
          />
        </div>
      </EditorRow>
    );
  }

  return (
    <EditorRow
      error={null}
      disabled={!stringValue.trim()}
      onSubmit={handleSubmit}
      isEditing={isEditing}
    >
      <div className="flex-1">
        <label className="block text-xs font-medium text-display mb-1">Value</label>
        <input
          type="text"
          value={stringValue}
          onChange={(e) => setStringValue(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleSubmit();
          }}
          placeholder={`Enter ${schemaKey}...`}
          className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
        />
      </div>
    </EditorRow>
  );
};
