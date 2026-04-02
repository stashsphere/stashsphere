import { useContext, useEffect, useMemo, useState } from 'react';
import { Property, SchemaCollection } from '../../api/resources';
import { getSchemaCollection } from '../../api/properties';
import { AxiosContext } from '../../context/axios';
import { PropertyList } from '../shared/property_list';
import { KeyAutocomplete } from './key_autocomplete';
import { PropertyEditor } from './property_editors/editor';
import CustomEditor from './property_editors/custom';
import { Modal } from '../shared/modal';
import { SchemaGrid } from './schema_grid';
import { SecondaryButton } from '../shared/button';

interface Props {
  properties: Property[];
  onUpdateProperties?: (properties: Property[]) => void;
}

const Properties: React.FC<Props> = ({ properties, onUpdateProperties }) => {
  const [localProperties, setLocalProperties] = useState<Property[]>([]);
  const axiosInstance = useContext(AxiosContext);
  const [schemaCollection, setSchemaCollection] = useState<SchemaCollection | undefined>(undefined);
  const [selectedKey, setSelectedKey] = useState<string | null>(null);
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [customMode, setCustomMode] = useState(false);
  const [customInitialKey, setCustomInitialKey] = useState('');
  const [showSchemaBrowser, setShowSchemaBrowser] = useState(false);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getSchemaCollection(axiosInstance).then(setSchemaCollection);
  }, [axiosInstance]);

  useEffect(() => {
    setLocalProperties(properties);
  }, [properties]);

  function handleSaveProperty(prop: Property) {
    let newProperties;
    if (editingIndex !== null) {
      newProperties = localProperties.map((p, i) => (i === editingIndex ? prop : p));
    } else {
      newProperties = [...localProperties, prop];
    }
    setSelectedKey(null);
    setEditingIndex(null);
    setCustomMode(false);
    setCustomInitialKey('');
    if (onUpdateProperties) onUpdateProperties(newProperties);
  }

  const deleteProperty = (indexToDelete: number) => {
    const newProperties = localProperties.filter((_, index) => index !== indexToDelete);
    setLocalProperties(newProperties);
    if (onUpdateProperties) onUpdateProperties(newProperties);
  };

  const handleClearEditor = () => {
    setSelectedKey(null);
    setEditingIndex(null);
    setCustomMode(false);
    setCustomInitialKey('');
  };

  const handleCustom = (key: string) => {
    setSelectedKey(null);
    setEditingIndex(null);
    setCustomMode(true);
    setCustomInitialKey(key);
  };

  const handleEditProperty = (index: number) => {
    const prop = properties[index];
    setEditingIndex(index);
    if (schemaCollection && schemaCollection.schemas[prop.name]) {
      setSelectedKey(prop.name);
      setCustomMode(false);
      setCustomInitialKey('');
    } else {
      setSelectedKey(null);
      setCustomMode(true);
      setCustomInitialKey('');
    }
  };

  const usedKeys = useMemo(() => {
    const keys = new Set(properties.map((p) => p.name));
    if (editingIndex !== null) {
      keys.delete(properties[editingIndex].name);
    }
    return keys;
  }, [properties, editingIndex]);

  const schema = selectedKey && schemaCollection ? schemaCollection.schemas[selectedKey] : null;

  if (!schemaCollection) {
    return (
      <>
        <h2 className="text-xl font-bold mb-4 text-secondary">Properties</h2>
        <p>Loading Schemas</p>
      </>
    );
  }

  return (
    <>
      <h2 className="text-xl font-bold mb-4 text-secondary">Properties</h2>

      <div className="flex gap-2 items-start">
        <div className="flex-1">
          <KeyAutocomplete
            collection={schemaCollection}
            onSelect={(key) => {
              setSelectedKey(key);
              setEditingIndex(null);
              setCustomMode(false);
              setCustomInitialKey('');
            }}
            onCustom={handleCustom}
            selectedKey={selectedKey}
            isCustom={customMode}
            onClear={handleClearEditor}
            usedKeys={usedKeys}
          />
        </div>
        {!selectedKey && !customMode && (
          <SecondaryButton onClick={() => setShowSchemaBrowser(true)} className="whitespace-nowrap">
            Browse All
          </SecondaryButton>
        )}
      </div>

      {schema && selectedKey && (
        <div className="border-t border-secondary-800 pt-4">
          <PropertyEditor
            key={`${selectedKey}-${editingIndex}`}
            schemaKey={selectedKey}
            schema={schema}
            onSave={handleSaveProperty}
            onClear={handleClearEditor}
            initialValue={editingIndex !== null ? properties[editingIndex] : undefined}
          />
        </div>
      )}

      {customMode && (
        <div className="border-t border-secondary-100 pt-4">
          <CustomEditor
            key={`custom-${editingIndex}-${customInitialKey}`}
            onSave={handleSaveProperty}
            onClear={handleClearEditor}
            initialKey={customInitialKey}
            initialValue={editingIndex !== null ? properties[editingIndex] : undefined}
          />
        </div>
      )}

      <PropertyList
        properties={properties}
        schemaCollection={schemaCollection}
        editingIndex={editingIndex}
        onEdit={handleEditProperty}
        onDelete={deleteProperty}
      />

      <Modal
        isOpen={showSchemaBrowser}
        onClose={() => setShowSchemaBrowser(false)}
        title="Browse Property Schemas"
        size="xl"
      >
        <SchemaGrid
          schemas={schemaCollection}
          onSelectSchema={(key) => {
            setSelectedKey(key);
            setShowSchemaBrowser(false);
            setEditingIndex(null);
            setCustomMode(false);
            setCustomInitialKey('');
          }}
        />
      </Modal>
    </>
  );
};

export default Properties;
