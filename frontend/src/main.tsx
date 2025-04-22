import React from 'react';
import ReactDOM from 'react-dom/client';
import { App } from './App.tsx';
import { BrowserRouter } from 'react-router';
// TODO: refactor this
import { CookiesProvider } from 'react-cookie';

import './index.css';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <CookiesProvider>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </CookiesProvider>
  </React.StrictMode>
);
