import { useContext, useEffect, useMemo, useState } from "react";
import { AxiosContext } from "../context/axios";
import { SearchResult } from "../api/resources";
import { getSearchResult } from "../api/search";
import { SearchResultComponent } from "../components/search_result";
import { SearchContext } from "../context/search";

export const Search = () => {
    const axiosInstance = useContext(AxiosContext);
    const { searchTerm } = useContext(SearchContext);
    const [result, setResult] = useState<SearchResult | undefined>(undefined);

    useEffect(() => {
        if (!axiosInstance) {
            return;
        }
        if (searchTerm != "") {
            getSearchResult(axiosInstance, searchTerm).then((res) => {
                setResult(res);
            }).catch((err) => {
                console.log("Error: ", err);
            })
        } else {
            setResult(undefined);
        }
    }, [axiosInstance, searchTerm]);
    
    const resultParam = useMemo(() => {
        if (result) {
            return result;
        } else {
            return {
                things: [],
                lists: [],
            }
        }
    }, [result]);
    
    return <>
        <SearchResultComponent result={resultParam} />
    </>
}