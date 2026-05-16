import tr from './tr';
import en from './en';

const langs = { tr, en };

export function getTranslations(lang) {
  return langs[lang] || langs.en;
}

export const availableLangs = [
  { code: 'tr', label: 'Türkçe', flag: '🇹🇷' },
  { code: 'en', label: 'English', flag: '🇬🇧' },
];
