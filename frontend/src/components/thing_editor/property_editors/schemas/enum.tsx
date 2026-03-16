import { useState } from 'react';
import { useValidator } from '../../../../hooks/useValidator';
import { EditorRow } from '../editor_row';
import { EditorProps } from '../editor';
import { substituteUnits } from '../../../../lib/units';

export const EnumEditor = ({
  schemaKey,
  schema,
  onSave,
  initialValue,
  isEditing,
}: EditorProps & {
  isEditing: boolean;
}) => {
  const valueDef = schema.schema.properties.value;
  const unitDef = schema.schema.properties.unit;
  const unitOptions = unitDef && Array.isArray(unitDef.enum) ? (unitDef.enum as string[]) : [];
  const options = valueDef.enum as string[];

  const [value, setValue] = useState(initialValue ? String(initialValue.value) : '');
  const [unit, setUnit] = useState(initialValue?.unit ?? '');
  const [error, setError] = useState<string | null>(null);

  const validate = useValidator(schema.schema);

  const handleSubmit = () => {
    if (!value) return;
    const candidate: Record<string, unknown> = { name: schemaKey, value };
    if (!!unitDef && unitOptions.length > 0) {
      candidate.unit = unit || unitOptions[0];
    }
    const err = validate(candidate);
    if (err) {
      setError(err);
      return;
    }
    if (!!unitDef && unitOptions.length > 0) {
      onSave({ type: 'float', name: schemaKey, value: 0, unit: unit || unitOptions[0] });
    } else {
      onSave({ type: 'string', name: schemaKey, value, unit: undefined });
    }
    setValue('');
    setError(null);
  };

  return (
    <EditorRow error={error} disabled={!value} onSubmit={handleSubmit} isEditing={isEditing}>
      <div className="flex-1 min-w-[180px]">
        <label className="block text-xs font-medium text-display mb-1">Value</label>
        <div className="flex flex-wrap gap-1">
          {options.map((opt) => (
            <button
              key={opt}
              onClick={() => {
                setValue(opt);
                setError(null);
              }}
              className={`rounded-sm px-2 py-1 text-sm transition-colors ${value === opt ? 'bg-primary text-onprimary' : 'bg-secondary-900 text-display hover:bg-secondary-800'}`}
            >
              {opt}
            </button>
          ))}
        </div>
      </div>
      {!!unitDef && unitOptions.length > 0 && (
        <div>
          <label className="block text-xs font-medium text-display mb-1">Unit</label>
          <div className="flex flex-wrap gap-1">
            {unitOptions.map((u) => (
              <button
                key={u}
                onClick={() => {
                  setUnit(u);
                  setError(null);
                }}
                className={`rounded-sm px-2 py-1 text-sm transition-colors ${(unit || unitOptions[0]) === u ? 'bg-primary text-onprimary' : 'bg-secondary-900 text-display hover:bg-secondary-800'}`}
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
