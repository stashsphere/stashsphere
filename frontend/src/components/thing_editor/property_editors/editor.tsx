import { useState } from 'react';
import { Property, SchemaEntry } from '../../../api/resources';
import { SchemaIcon } from '../../shared/svg_icon';
import { Icon } from '../../shared';
import { StringEditor } from './schemas/string';
import { NumberEditor } from './schemas/number';
import { BooleanEditor } from './schemas/boolean';
import { DatetimeEditor } from './schemas/datetime';
import { EnumEditor } from './schemas/enum';
import { OverrideEditor } from './schemas/override';

export type EditorProps = {
  schemaKey: string;
  schema: SchemaEntry;
  onSave: (property: Property) => void;
  onClear?: () => void;
  initialValue?: Property;
};

const resolveBaseType = (valueDef: Record<string, unknown>): 'string' | 'number' | 'boolean' => {
  const t = valueDef.type as string | undefined;
  switch (t) {
    case 'number':
      return 'number';
    case 'boolean':
      return 'boolean';
    default:
      return 'string';
  }
};

export const PropertyEditor = ({
  onSave,
  onClear,
  schema,
  schemaKey,
  initialValue,
}: EditorProps) => {
  const isEditing = !!initialValue;
  const icon = schema.icons?.[0];

  const valueDef = schema.schema.properties.value;
  const valueType = valueDef.type as string | undefined;
  const format = valueDef.format as string | undefined;
  const hasEnum = Array.isArray(valueDef.enum);
  const baseType = resolveBaseType(valueDef);
  const [override, setOverride] = useState(false);

  const editor = (() => {
    if (override) {
      return (
        <OverrideEditor
          schemaKey={schemaKey}
          schema={schema}
          isEditing={isEditing}
          onSave={onSave}
          initialValue={initialValue}
          baseType={baseType}
        />
      );
    } else if (valueType === 'boolean') {
      return (
        <BooleanEditor
          schemaKey={schemaKey}
          schema={schema}
          isEditing={isEditing}
          onSave={onSave}
          initialValue={initialValue}
        />
      );
    } else if (valueType === 'number') {
      return (
        <NumberEditor
          schemaKey={schemaKey}
          schema={schema}
          isEditing={isEditing}
          onSave={onSave}
          initialValue={initialValue}
        />
      );
    } else if (valueType === 'string' && format === 'date-time') {
      return (
        <DatetimeEditor
          schemaKey={schemaKey}
          schema={schema}
          onSave={onSave}
          initialValue={initialValue}
          isEditing={isEditing}
        />
      );
    } else if (hasEnum) {
      return (
        <EnumEditor
          schemaKey={schemaKey}
          schema={schema}
          onSave={onSave}
          initialValue={initialValue}
          isEditing={isEditing}
        />
      );
    }
    return (
      <StringEditor
        schemaKey={schemaKey}
        schema={schema}
        isEditing={isEditing}
        onSave={onSave}
        initialValue={initialValue}
      />
    );
  })();

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {icon && <SchemaIcon icon={icon} className="text-lg text-display-light" />}
          <span className="text-sm font-medium text-display">{schemaKey}</span>
          {onClear && (
            <button
              type="button"
              onClick={onClear}
              className="text-display-light hover:text-accent transition-colors"
            >
              <Icon icon="mdi--clear" />
            </button>
          )}
        </div>
        <button
          type="button"
          onClick={() => setOverride(!override)}
          className={`text-xs transition-colors ${
            override
              ? 'font-medium text-accent hover:text-accent'
              : 'text-display-light hover:text-display'
          }`}
        >
          {override ? 'Use schema editor' : 'Override'}
        </button>
      </div>
      {editor}
    </div>
  );
};
