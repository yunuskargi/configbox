import { createContext, useContext, useState } from 'react';
import { getTranslations } from '../i18n';

const LangContext = createContext({});

export function LangProvider({ children }) {
  const [lang, setLang] = useState(() => localStorage.getItem('lang') || 'en');

  const t = getTranslations(lang);

  const switchLang = (code) => {
    setLang(code);
    localStorage.setItem('lang', code);
  };

  return (
    <LangContext.Provider value={{ lang, t, switchLang }}>
      {children}
    </LangContext.Provider>
  );
}

export const useLang = () => useContext(LangContext);
