import { createContext } from 'react';

// Holds the public share token when the app renders a publicly shared
// thing or list for an anonymous visitor. Components that build asset
// URLs read it so image requests carry the token.
export const PublicShareTokenContext = createContext<string | null>(null);
