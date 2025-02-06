import { createContext } from 'react'
import { Profile } from '../api/resources';

type AuthContextValue = {
    loggedIn: boolean;
    profile: Profile | null;
    invalidateProfile: () => void;
}

export const AuthContext = createContext<AuthContextValue>({ loggedIn: false, profile: null, invalidateProfile: () => {} });