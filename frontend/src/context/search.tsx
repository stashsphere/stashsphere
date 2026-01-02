import {createContext} from 'react';

export const SearchContext = createContext({
    searchTerm: '',
    setSearchTerm: (_v: string) => {},
});