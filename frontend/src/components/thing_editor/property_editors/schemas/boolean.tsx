import { useState } from 'react';
import { useValidator } from '../../../../hooks/useValidator';
import { EditorProps } from '../editor';
import { EditorRow } from '../editor_row';
import { Toggle } from '../../../shared/toggle';

export const BooleanEditor = ({
  schemaKey,
  schema,
  onSave,
  initialValue,
  isEditing,
}: EditorProps & {
  isEditing: boolean;
}) => {
  const [value, setValue] = useState(initialValue?.type === 'boolean' ? initialValue.value : false);
  // this might not be necessary
  const [error, setError] = useState<string | null>(null);

  const validate = useValidator(schema.schema);

  const handleSubmit = () => {
    if (!validate) {
      console.error('validate function is null');
      return;
    }
    const err = validate({ name: schemaKey, value });
    if (err) {
      setError(err);
      return;
    }
    onSave({ type: 'boolean', name: schemaKey, value, unit: undefined });
    setValue(false);
    setError(null);
  };

  return (
    <EditorRow error={error} disabled={false} onSubmit={handleSubmit} isEditing={isEditing}>
      <div className="flex-1">
        <label className="block text-xs font-medium text-display mb-1">Value</label>
        <div className="flex items-center gap-3 rounded-sm border border-secondary shadow-xs px-3 py-1.5">
          <Toggle
            onChange={(v) => {
              setValue(v);
              setError(null);
            }}
            value={value}
          />
          <span className="text-sm text-display">{value ? 'true' : 'false'}</span>
        </div>
      </div>
    </EditorRow>
  );
};
