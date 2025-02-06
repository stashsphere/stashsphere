import { SearchResult } from "../api/resources";
import { ListInfo } from "./list_info";
import ThingInfo from "./thing_info";

interface SearchResultProps {
    result: SearchResult
}

export const SearchResultComponent = ({ result }: SearchResultProps) => {
    return <>
        <h2 className="text-xl font-bold mb-4 text-secondary">Things</h2>
        {result.things.length == 0 ? <p className="text-primary">No Things</p> : result.things.map((thing) => <ThingInfo thing={thing} key={thing.id} />)}
        <h2 className="text-xl font-bold mb-4 text-secondary">Lists</h2>
        {result.lists.length == 0  ? <p className="text-primary">No Lists</p> : result.lists.map((list) => <ListInfo list={list} key={list.id} />)}
    </>
}
