import { useMemo, useState } from 'react';
import { Property, SchemaCollection } from '../../api/resources';
import { Icon } from './icon';
import { SchemaIcon } from './svg_icon';
import { substituteUnits } from '../../lib/units';
import { contrastBackground } from '../../lib/contrast';
import { validateColor } from '../../lib/validate_color';

const formatValue = (prop: Property): string => {
  if (prop.type === 'datetime') {
    return new Date(prop.value).toLocaleString();
  }
  return String(prop.value);
};

type PropertyRowProps = {
  property: Property;
  icon?: string;
  isEditing?: boolean;
  onEdit?: () => void;
  onDelete?: () => void;
};

const PropertyRow: React.FC<PropertyRowProps> = ({
  property,
  icon,
  isEditing,
  onEdit,
  onDelete,
}) => {
  const isClickable = !!onEdit;

  const iconOrDefault = icon ?? 'mdi:help-rhombus';
  const color = useMemo(
    () => (property.name === 'color' ? validateColor(property.value) : undefined),
    [property.name, property.value]
  );

  return (
    <div
      onClick={onEdit}
      className={`col-span-full grid grid-cols-subgrid items-center gap-x-2 rounded-sm px-3 py-2 transition-colors ${
        isEditing
          ? 'bg-primary/10 border border-primary'
          : isClickable
            ? 'hover:bg-secondary-900 cursor-pointer'
            : ''
      }`}
    >
      <span className="inline-flex items-center justify-center w-6 h-6">
        {property.name === 'color' ? (
          <span
            className="inline-flex items-center justify-center shrink-0 w-6 h-6 rounded-sm"
            style={{ backgroundColor: contrastBackground(color) }}
          >
            <SchemaIcon icon={iconOrDefault} className="text-base" color={color} />
          </span>
        ) : (
          <SchemaIcon icon={iconOrDefault} className="text-base text-display-light" />
        )}
      </span>
      <span className="font-medium text-display text-sm whitespace-nowrap">{property.name}</span>
      <span className="text-display-light">=</span>
      <span className="flex items-center gap-1.5 min-w-0">
        <span className="truncate text-sm text-display">{formatValue(property)}</span>
        {property.unit && (
          <span className="shrink-0 rounded-sm bg-secondary-900 px-1.5 py-0.5 text-xs text-display-light whitespace-nowrap">
            {substituteUnits(property.unit)}
          </span>
        )}
      </span>
      {onDelete && (
        <button
          className="text-display-light hover:text-warning transition-colors"
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
        >
          <Icon icon="mdi--trash" />
        </button>
      )}
    </div>
  );
};

type PropertyListProps = {
  properties: Property[];
  schemaCollection?: SchemaCollection;
  editingIndex?: number | null;
  onEdit?: (index: number) => void;
  onDelete?: (index: number) => void;
  collapsable?: boolean;
};

export const PropertyList: React.FC<PropertyListProps> = ({
  properties,
  schemaCollection,
  editingIndex,
  onEdit,
  onDelete,
  collapsable,
}) => {
  const [collapsed, setCollapsed] = useState(true);

  const collapsedDisplay = collapsable ? 3 : properties.length;
  const displayedProperties = collapsed ? properties.slice(0, collapsedDisplay) : properties;
  const hasMore = collapsable && properties.length > collapsedDisplay;

  const gridCols = onDelete
    ? 'grid-cols-[auto_auto_auto_1fr_auto]'
    : 'grid-cols-[auto_auto_auto_1fr]';

  return (
    <div className="overflow-x-auto">
      <div className={`grid ${gridCols} gap-y-3`}>
        {displayedProperties.map((prop, index) => {
          const icon = schemaCollection?.schemas[prop.name]?.icons?.[0];
          return (
            <PropertyRow
              key={index}
              property={prop}
              icon={icon}
              isEditing={editingIndex === index}
              onEdit={onEdit ? () => onEdit(index) : undefined}
              onDelete={onDelete ? () => onDelete(index) : undefined}
            />
          );
        })}
      </div>
      {hasMore && (
        <button onClick={() => setCollapsed((prev) => !prev)}>
          {collapsed ? 'Show more' : 'Show less'}
        </button>
      )}
    </div>
  );
};
