import { useState, useMemo } from 'react';
import { SchemaCollection, SchemaEntry } from '../../api/resources';
import { SchemaIcon } from '../shared/svg_icon';

interface SchemaGridProps {
  schemas: SchemaCollection;
  onSelectSchema: (schemaKey: string) => void;
}

type PropertyType = 'number' | 'string' | 'boolean' | 'datetime' | 'enum';

interface SchemaCardData {
  key: string;
  title: string;
  description: string;
  type: PropertyType;
  units: string[];
  icon: string;
}

const getPropertyType = (schema: SchemaEntry): PropertyType => {
  const valueProps = schema.schema.properties.value;

  if (!valueProps) return 'string';

  // Check for enum first
  if ('enum' in valueProps && valueProps.enum) {
    return 'enum';
  }

  // Check type
  const type = 'type' in valueProps ? valueProps.type : undefined;

  if (type === 'number') {
    return 'number';
  }

  if (type === 'boolean') {
    return 'boolean';
  }

  if (type === 'string') {
    const format = 'format' in valueProps ? valueProps.format : undefined;
    if (format === 'date-time') {
      return 'datetime';
    }
    return 'string';
  }

  return 'string';
};

const getUnits = (schema: SchemaEntry): string[] => {
  const unitProps = schema.schema.properties.unit;

  if (!unitProps || !('enum' in unitProps)) {
    return [];
  }

  return (unitProps.enum as string[]) || [];
};

const getTypeBadgeColor = (type: PropertyType): string => {
  switch (type) {
    case 'number':
      return 'bg-blue-100 text-blue-800';
    case 'string':
      return 'bg-green-100 text-green-800';
    case 'boolean':
      return 'bg-purple-100 text-purple-800';
    case 'datetime':
      return 'bg-orange-100 text-orange-800';
    case 'enum':
      return 'bg-pink-100 text-pink-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
};

const getTypeLabel = (type: PropertyType): string => {
  switch (type) {
    case 'number':
      return 'Number';
    case 'string':
      return 'String';
    case 'boolean':
      return 'Boolean';
    case 'datetime':
      return 'Datetime';
    case 'enum':
      return 'Enum';
    default:
      return 'Unknown';
  }
};

export const SchemaGrid: React.FC<SchemaGridProps> = ({ schemas, onSelectSchema }) => {
  const [searchQuery, setSearchQuery] = useState('');

  const schemaCards: SchemaCardData[] = useMemo(() => {
    return Object.entries(schemas.schemas).map(([key, schemaEntry]) => {
      const title = schemaEntry.schema.title || key;
      const description = schemaEntry.schema.description || '';
      const type = getPropertyType(schemaEntry);
      const units = getUnits(schemaEntry);
      const icon = schemaEntry.icons?.[0] || 'mdi:help-rhombus';

      return {
        key,
        title,
        description,
        type,
        units,
        icon,
      };
    });
  }, [schemas]);

  const filteredSchemas = useMemo(() => {
    if (!searchQuery.trim()) {
      return schemaCards;
    }

    const query = searchQuery.toLowerCase();

    return schemaCards.filter((card) => {
      // Search in title
      if (card.title.toLowerCase().includes(query)) {
        return true;
      }

      // Search in description
      if (card.description.toLowerCase().includes(query)) {
        return true;
      }

      // Search in key
      if (card.key.toLowerCase().includes(query)) {
        return true;
      }

      // Search in aliases
      const schemaEntry = schemas.schemas[card.key];
      const aliases = schemaEntry.schema['x-aliases'] || [];
      if (aliases.some((alias) => alias.toLowerCase().includes(query))) {
        return true;
      }

      return false;
    });
  }, [schemaCards, searchQuery, schemas]);

  return (
    <div className="flex flex-col" style={{ height: 'calc(100vh - 12rem)' }}>
      <div className="mb-4 flex-shrink-0">
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Search schemas by name, description, or alias..."
          className="w-full p-2 border border-gray-300 rounded-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary"
          autoFocus
        />
      </div>

      <div className="flex-1 overflow-y-auto -mx-3 sm:-mx-4 px-3 sm:px-4">
        {filteredSchemas.length === 0 ? (
          <div className="text-center py-8 text-gray-500">No schemas match your search</div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 pb-4">
            {filteredSchemas.map((card) => (
              <button
                key={card.key}
                onClick={() => onSelectSchema(card.key)}
                className="text-left p-3 border border-gray-200 rounded-sm hover:border-primary hover:bg-primary/5 transition-all focus:outline-none focus:ring-2 focus:ring-primary"
              >
                <div className="flex items-start gap-2 mb-2">
                  <div className="flex-shrink-0 mt-0.5">
                    <SchemaIcon icon={card.icon} className="text-xl" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <h3 className="font-semibold text-sm text-gray-900 truncate">{card.title}</h3>
                  </div>
                </div>

                {card.description && (
                  <p className="text-xs text-gray-600 mb-2 line-clamp-2">{card.description}</p>
                )}

                <div className="flex items-center gap-2 flex-wrap">
                  <span
                    className={`text-xs px-2 py-0.5 rounded-sm font-medium ${getTypeBadgeColor(card.type)}`}
                  >
                    {getTypeLabel(card.type)}
                  </span>

                  {/* Units */}
                  {card.units.length > 0 && (
                    <span className="text-xs text-gray-500">
                      Units: {card.units.slice(0, 3).join(', ')}
                      {card.units.length > 3 && '...'}
                    </span>
                  )}
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
