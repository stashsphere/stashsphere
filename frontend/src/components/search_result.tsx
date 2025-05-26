import { SearchResult } from '../api/resources';
import { ListInfo } from './list_info';
import { ThingInfo } from './shared';

interface SearchResultProps {
  result: SearchResult;
}

export const SearchResultComponent = ({ result }: SearchResultProps) => {
  return (
    <>
      <h2 className="text-2xl mb-4 text-accent">Things</h2>
      {result.things.length == 0 ? (
        <p className="text-primary">No Things</p>
      ) : (
        <div className="flex flex-row gap-4 mt-4 flex-wrap justify-center">
          {result.things.map((thing) => (
            <ThingInfo thing={thing} key={thing.id} />
          ))}
        </div>
      )}
      <h2 className="text-2xl mb-4 text-accent">Lists</h2>
      {result.lists.length == 0 ? (
        <p className="text-primary">No Lists</p>
      ) : (
        <div className="flex flex-row gap-4 mt-4 flex-wrap justify-center">
          {result.lists.map((list) => (
            <ListInfo list={list} key={list.id} />
          ))}
        </div>
      )}
    </>
  );
};
