import { ReactNode, useCallback } from 'react';
import { SelectableThing } from './shared';
import { SharingState, Thing } from '../api/resources';

export type ListEditorData = {
  name: string;
  selectedThingIDs: string[];
  sharingState: SharingState;
};

type ListEditorProps = {
  children?: ReactNode;
  list: ListEditorData;
  selectableThings: Thing[];
  onChange: (list: ListEditorData) => void;
};

export const ListEditor = ({ children, list, onChange, selectableThings }: ListEditorProps) => {
  const onThingSelect = useCallback(
    (thingID: string, isChecked: boolean) => {
      const selectedThingIDs = list.selectedThingIDs;
      if (isChecked) {
        onChange({
          ...list,
          selectedThingIDs: [...selectedThingIDs, thingID],
        });
      } else {
        const index = selectedThingIDs.indexOf(thingID);
        if (index > -1) {
          const updatedSelectedThingIDs = [...selectedThingIDs];
          updatedSelectedThingIDs.splice(index, 1);
          onChange({
            ...list,
            selectedThingIDs: updatedSelectedThingIDs,
          });
        }
      }
    },
    [list, onChange]
  );

  const onNameChange = useCallback(
    (value: string) => {
      onChange({
        ...list,
        name: value,
      });
    },
    [list, onChange]
  );

  return (
    <div>
      <div className="mb-4 flex justify-between">
        <div>
          <label htmlFor="email" className="block text-primary text-sm font-medium">
            Name
          </label>
          <input
            type="text"
            id="name"
            name="name"
            value={list.name}
            onChange={(e) => onNameChange(e.target.value)}
            className="mt-1 p-2 text-display border border-gray-300 rounded-sm"
          />
        </div>
        <div className="flex flex-col justify-end">{children}</div>
      </div>
      <div className="flex flex-wrap gap-4">
        {selectableThings.map((thing) => (
          <SelectableThing
            key={thing.id}
            thing={thing}
            selected={list.selectedThingIDs.includes(thing.id)}
            onSelect={onThingSelect}
          />
        ))}
      </div>
    </div>
  );
};
