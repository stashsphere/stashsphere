import { FormEvent, ReactNode, useEffect, useState } from 'react';
import { Thing } from '../api/resources';
import ThingInfo from './thing_info';

export type ListEditorData = {
  name: string;
  selectedThingIDs: string[];
};

type ListEditorProps = {
  children?: ReactNode;
  list?: ListEditorData;
  selectableThings: Thing[];
  onChange: (list: ListEditorData) => void;
};

export const ListEditor = ({ children, list, onChange, selectableThings }: ListEditorProps) => {
  const [name, setName] = useState('');
  const [selectedThingIDs, setSelectedThingIDs] = useState<string[]>([]);

  useEffect(() => {
    if (!list) {
      return;
    }
    setName(list.name);
    setSelectedThingIDs(list.selectedThingIDs);
  }, [list]);

  const onSubmit = (event: FormEvent) => {
    event.preventDefault();
    const data = {
      name,
      selectedThingIDs,
    };
    onChange(data);
  };

  const onThingSelect = (thingID: string, isChecked: boolean) => {
    if (isChecked) {
      setSelectedThingIDs([...selectedThingIDs, thingID]);
    } else {
      const index = selectedThingIDs.indexOf(thingID);
      if (index > -1) {
        const updatedSelectedThingIDs = [...selectedThingIDs];
        updatedSelectedThingIDs.splice(index, 1);
        setSelectedThingIDs(updatedSelectedThingIDs);
      }
    }
  };

  return (
    <form onSubmit={onSubmit}>
      <div className="mb-4">
        <label htmlFor="email" className="block text-primary text-sm font-medium">
          Name
        </label>
        <input
          type="text"
          id="name"
          name="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="mt-1 p-2 text-display border border-gray-300 rounded-sm"
        />
      </div>
      {selectableThings.map((thing) => (
        <div className="flex flex-row border border-gray-300">
          <div className="px-4 border border-r-gray-300">
            <input
              type="checkbox"
              checked={selectedThingIDs.includes(thing.id)}
              onChange={(e) => onThingSelect(thing.id, e.target.checked)}
            />
          </div>
          <ThingInfo thing={thing} key={thing.id} />
        </div>
      ))}
      {children}
    </form>
  );
};
