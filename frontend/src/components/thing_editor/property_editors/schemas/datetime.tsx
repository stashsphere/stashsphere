import { useState } from 'react';
import { EditorProps } from '../editor';
import { useValidator } from '../../../../hooks/useValidator';
import { EditorRow } from '../editor_row';

export const DatetimeEditor = ({
  schemaKey,
  schema,
  onSave,
  initialValue,
  isEditing,
}: EditorProps & {
  isEditing: boolean;
}) => {
  const initDate = initialValue?.type === 'datetime' ? initialValue.value.slice(0, 10) : '';
  const initTime = initialValue?.type === 'datetime' ? initialValue.value.slice(11, 16) : '';
  const [date, setDate] = useState(initDate);
  const [time, setTime] = useState(initTime);
  const [error, setError] = useState<string | null>(null);

  const validate = useValidator(schema.schema);

  function handleSubmit() {
    if (!date) return;
    const t = time || '00:00';
    const isoValue = new Date(`${date}T${t}`).toISOString();
    const err = validate({ name: schemaKey, value: isoValue });
    if (err) {
      setError(err);
      return;
    }
    onSave({ type: 'datetime', name: schemaKey, value: isoValue, unit: undefined });
    setDate('');
    setTime('');
    setError(null);
  }

  return (
    <EditorRow error={error} disabled={!date} onSubmit={handleSubmit} isEditing={isEditing}>
      <div className="flex-1 min-w-[180px]">
        <label className="block text-xs font-medium text-display mb-1">Date</label>
        <input
          type="date"
          value={date}
          onChange={(e) => {
            setDate(e.target.value);
            setError(null);
          }}
          className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
        />
      </div>
      <div className="flex-1 min-w-[180px]">
        <label className="block text-xs font-medium text-display mb-1">Time</label>
        <input
          type="time"
          value={time}
          onChange={(e) => {
            setTime(e.target.value);
            setError(null);
          }}
          className="w-full text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1"
        />
      </div>
    </EditorRow>
  );
};
