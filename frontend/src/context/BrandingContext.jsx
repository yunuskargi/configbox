import { createContext, useContext, useEffect, useState } from 'react';
import api from '../api/client';

const BrandingContext = createContext({});

export function BrandingProvider({ children }) {
  const [appTitle, setAppTitle] = useState('');

  useEffect(() => {
    api.get('/settings/branding').then((r) => setAppTitle(r.data.app_title || '')).catch(() => {});
  }, []);

  const refresh = () => {
    api.get('/settings/branding').then((r) => setAppTitle(r.data.app_title || '')).catch(() => {});
  };

  return (
    <BrandingContext.Provider value={{ appTitle, refresh }}>
      {children}
    </BrandingContext.Provider>
  );
}

export const useBranding = () => useContext(BrandingContext);
