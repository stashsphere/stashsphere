import { Axios } from "axios";
import { SearchResult } from "./resources";

export const getSearchResult = async (axios: Axios, query: string): Promise<SearchResult> => {
    const encodedQuery = encodeURIComponent(query);
    const response = await axios.get(`/search?query=${encodedQuery}`, {
        headers: {
            "Content-Type": "application/json"
        }
    })

    if (response.status != 200) {
        throw `Got error ${response}`
    }

    return response.data as SearchResult;
}