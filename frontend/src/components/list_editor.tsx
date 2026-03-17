import { ReactNode, useContext, useEffect, useMemo, useState } from 'react';
import { FetchedSelectableThing, SelectableThing } from './shared';
import { PagedThings, SharingState } from '../api/resources';
import { getThings } from '../api/things';
import { Pages } from './pages';
import { AuthContext } from '../context/auth';
import { AxiosContext } from '../context/axios';
import { SharingStateComponent } from './shared/sharing_state';

export type ListEditorData = {
  name: string;
  selectedThingIDs: string[];
  sharingState: SharingState;
};

type ListEditorProps = {
  children?: ReactNode;
  list?: ListEditorData;
  onChange: (list: ListEditorData) => void;
  defaultSharingState?: SharingState;
};

const thingsPerPage = 10;

export const ListEditor = ({
  children,
  list,
  onChange,
  defaultSharingState = 'private',
}: ListEditorProps) => {
  const authCtx = useContext(AuthContext);
  const axiosInstance = useContext(AxiosContext);
  const [name, setName] = useState('');
  const [selectedThingIDs, setSelectedThingIDs] = useState<string[]>([]);
  const [sharingStateAndDefault, setSharingStateAndDefault] = useState<[SharingState, boolean]>([
    'private',
    true,
  ]);
  const [currentPage, setCurrentPage] = useState(0);
  const [selectedPage, setSelectedPage] = useState(0);
  const [textFilter, setTextFilter] = useState('');

  const [selectableThingsPages, setSelectableThingsPages] = useState<PagedThings | undefined>(
    undefined
  );

  const [sharingState, usingDefaultSharingState] = sharingStateAndDefault;

  useEffect(() => {
    if (!list) {
      return;
    }
    setName(list.name);
    setSelectedThingIDs(list.selectedThingIDs);
    setSharingStateAndDefault([list.sharingState, false]);
  }, [list]);

  useEffect(() => {
    if (list) {
      // If we have a loaded list, don't override its sharing state
      return;
    }
    if (defaultSharingState) {
      setSharingStateAndDefault([defaultSharingState, true]);
    }
  }, [defaultSharingState, list]);

  useEffect(() => {
    const data = {
      name,
      selectedThingIDs,
      sharingState,
    };
    onChange(data);
  }, [name, selectedThingIDs, sharingState, onChange]);

  const onThingSelect = (thingID: string, isChecked: boolean) => {
    if (isChecked) {
      setSelectedThingIDs([...selectedThingIDs, thingID]);
    } else {
      setSelectedThingIDs(selectedThingIDs.filter((id) => id !== thingID));
    }
  };

  const onNameChange = (value: string) => {
    setName(value);
  };

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    if (authCtx.profile === null) {
      return;
    }
    getThings(axiosInstance, currentPage, thingsPerPage, [authCtx.profile.id], textFilter, [])
      .then(setSelectableThingsPages)
      .catch((reason) => {
        console.log(reason);
      });
  }, [authCtx.profile, axiosInstance, currentPage, textFilter]);

  const selectableThings = useMemo(() => {
    return selectableThingsPages?.things || [];
  }, [selectableThingsPages?.things]);

  const selectedThingIDsPage = useMemo(() => {
    const start = selectedPage * thingsPerPage;
    return selectedThingIDs.slice(start, start + thingsPerPage);
  }, [selectedThingIDs, selectedPage]);

  const onTextFilterChange = (filter: string) => {
    setCurrentPage(0);
    setTextFilter(filter);
  };

  const selectedTotalPages = Math.ceil(selectedThingIDs.length / thingsPerPage);

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
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            className="mt-1 p-2 text-display border border-gray-300 rounded-sm"
          />
        </div>
        <div className="flex flex-col justify-end">{children}</div>
      </div>
      {!list && (
        <div className="bg-neutral-900 border border-neutral-700 rounded-sm p-3 my-2 w-100">
          <p className="text-display text-sm flex">
            Sharing: <SharingStateComponent state={sharingState} />
            {usingDefaultSharingState && (
              <span className="text-secondary text-sm">(from your profile default)</span>
            )}
          </p>
        </div>
      )}

      <div className="flex gap-4 items-stretch">
        <div className="flex-1 flex flex-col justify-between">
          <div>
            <h3 className="text-primary text-sm font-medium mb-2">Search</h3>
            <div className="flex flex-wrap gap-4 content-start">
              {selectableThings.map((thing) => (
                <SelectableThing
                  key={thing.id}
                  thing={thing}
                  selected={selectedThingIDs.includes(thing.id)}
                  onSelect={onThingSelect}
                />
              ))}
            </div>
          </div>
          <div>
            <input
              type="text"
              id="textFilter"
              name="textFilter"
              value={textFilter}
              placeholder="Filter things"
              onChange={(e) => onTextFilterChange(e.target.value)}
              className="my-2 p-2 text-display border border-gray-300 rounded-sm w-2/3"
            />
            <Pages
              currentPage={currentPage}
              onPageChange={(n) => setCurrentPage(n)}
              pages={selectableThingsPages?.totalPageCount || 0}
            />
          </div>
        </div>

        <div className="flex-1 flex flex-col justify-between">
          <div>
            <h3 className="text-primary text-sm font-medium mb-2">
              Selected ({selectedThingIDs.length})
            </h3>
            <div className="flex flex-wrap gap-4 content-start">
              {selectedThingIDsPage.map((thingId) => (
                <FetchedSelectableThing
                  key={thingId}
                  thingId={thingId}
                  selected={true}
                  onSelect={onThingSelect}
                />
              ))}
            </div>
          </div>
          <div>
            <Pages
              currentPage={selectedPage}
              onPageChange={(n) => setSelectedPage(n)}
              pages={selectedTotalPages}
            />
          </div>
        </div>
      </div>
    </div>
  );
};
