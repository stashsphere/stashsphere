import { useState } from 'react';
import { EditorProps } from '../editor';
import { useValidator } from '../../../../hooks/useValidator';
import { EditorRow } from '../editor_row';
import { substituteUnits } from '../../../../lib/units';

export const NumberEditor = ({
  schemaKey,
  schema,
  onSave,
  initialValue,
  isEditing,
}: EditorProps & {
  isEditing: boolean;
}) => {
  const unitDef = schema.schema.properties.unit;
  const unitOptions = unitDef && Array.isArray(unitDef.enum) ? (unitDef.enum as string[]) : [];
  const [value, setValue] = useState(
    initialValue?.type === 'float' ? String(initialValue.value) : ''
  );
  const [error, setError] = useState<string | null>(null);
  const [unit, setUnit] = useState(
    (initialValue?.type === 'float' && initialValue?.unit) ?? unitOptions[0] ?? ''
  );
  const validate = useValidator(schema.schema);
  const valueDef = schema.schema.properties.value;

  const min = valueDef.minimum as number | undefined;
  const max = valueDef.maximum as number | undefined;

  const handleSubmit = () => {
    const num = parseFloat(value);
    if (isNaN(num)) return;
    const obj: Record<string, unknown> = { name: schemaKey, value: num };
    if (unitOptions.length > 0) {
      // TODO improve default value
      obj.unit = unit || 'unit';
    }
    const err = validate(obj);
    if (err) {
      setError(err);
      return;
    }
    onSave({ type: 'float', name: schemaKey, value: num, unit: unit || 'unit' });
    setValue('');
    setError(null);
  };

  return (
    <EditorRow
      error={error}
      disabled={!value || isNaN(parseFloat(value))}
      onSubmit={handleSubmit}
      isEditing={isEditing}
    >
      <div className="flex-1 min-w-[180px]">
        <label className="block text-xs font-medium text-display mb-1">
          Value
          {min !== undefined && max !== undefined && (
            <span className="ml-1 text-display-light font-normal">
              {min} to {max}
            </span>
          )}
        </label>
        <input
          type="number"
          value={value}
          onChange={(e) => {
            setValue(e.target.value);
            setError(null);
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              handleSubmit();
            }
          }}
          placeholder={`Enter ${schemaKey}`}
          step="any"
          className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
        />
      </div>
      {unitOptions.length > 0 && (
        <div>
          <label className="block text-xs font-medium text-display mb-1">Unit</label>
          <div className="flex flex-wrap gap-1">
            {unitOptions.map((u) => (
              <button
                key={u}
                onClick={() => setUnit(u)}
                className={`rounded-sm px-2 py-1 text-sm transition-colors ${unit === u ? 'bg-primary text-onprimary' : 'bg-secondary-900 text-display hover:bg-secondary-800'}`}
              >
                {substituteUnits(u)}
              </button>
            ))}
          </div>
        </div>
      )}
    </EditorRow>
  );
};
