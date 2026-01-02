import { Axios } from 'axios';
import React from 'react';

export const AxiosContext = React.createContext<Axios | null>(null);
