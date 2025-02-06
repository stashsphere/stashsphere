import { useContext, useEffect, useState } from "react";
import { AxiosContext } from "../context/axios";
import { SearchResult } from "../api/resources";
import { getSearchResult } from "../api/search";
import { SearchResultComponent } from "../components/search_result";

export const Search = () => {
    const axiosInstance = useContext(AxiosContext);
    // get query parameter
    const queryFromParameter = new URLSearchParams(window.location.search).get("query") || null;
    const [result, setResult] = useState<SearchResult | undefined>(undefined);
    const [query, setQuery] = useState("");

    useEffect(() => {
        if (queryFromParameter) {
            setQuery(queryFromParameter);
        }
    }, [queryFromParameter]);

    useEffect(() => {
        if (!axiosInstance) {
            return;
        }
        if (query != "") {
            getSearchResult(axiosInstance, query).then((res) => {
                setResult(res);
            }).catch((err) => {
                console.log("Error: ", err);
            })
        }
    }, [axiosInstance, query]);
    
    return <>
        <div className="search-container">
            <input type="text" placeholder="Search..." value={query} onChange={(e) => setQuery(e.target.value)} />
        </div>
        {result && <SearchResultComponent result={result} />}
    </>
}