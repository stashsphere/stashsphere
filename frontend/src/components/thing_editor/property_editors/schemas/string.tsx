import { useState } from 'react';
import { useValidator } from '../../../../hooks/useValidator';
import { EditorRow } from '../editor_row';
import { EditorProps } from '../editor';

export const StringEditor = ({
  schemaKey,
  schema,
  onSave,
  initialValue,
  isEditing,
}: EditorProps & {
  isEditing: boolean;
}) => {
  const [value, setValue] = useState(initialValue?.type === 'string' ? initialValue.value : '');
  const [error, setError] = useState<string | null>(null);

  const validate = useValidator(schema.schema);

  const valueDef = schema.schema.properties.value;
  // this may be `email`, `uri` etc.
  const format = valueDef.format as string | undefined;

  const inputType = format === 'email' ? 'email' : format === 'uri' ? 'url' : 'text';
  const placeHolder =
    format === 'email'
      ? 'user@example.com'
      : format === 'uri'
        ? 'https://example.com'
        : `Enter ${schemaKey}`;

  const handleSubmit = () => {
    if (!value.trim()) {
      return;
    }
    if (!validate) {
      console.error('validate function is null');
      return;
    }
    const err = validate({ name: schemaKey, value: value.trim() });
    if (err) {
      setError(err);
      return;
    }
    onSave({ type: 'string', name: schemaKey, value: value.trim(), unit: undefined });
    setValue('');
    setError(null);
  };

  return (
    <EditorRow error={error} disabled={!value.trim()} onSubmit={handleSubmit} isEditing={isEditing}>
      <div className="flex-1">
        <label className="block text-xs font-medium text-display mb-1">
          Value
          {format && <span className="ml-1 text-display-light font-normal">{format}</span>}
        </label>
        <input
          type={inputType}
          value={value}
          onChange={(e) => {
            setValue(e.target.value);
            setError(null);
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleSubmit();
          }}
          placeholder={placeHolder}
          className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
        />
      </div>
    </EditorRow>
  );
};
