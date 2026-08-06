import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import enCommon from '../locales/en/common.json';
import enPlaylist from '../locales/en/playlist.json';
import enFilters from '../locales/en/filters.json';
import enDownloads from '../locales/en/downloads.json';
import enLogs from '../locales/en/logs.json';
import enDialogs from '../locales/en/dialogs.json';

import frCommon from '../locales/fr/common.json';
import frPlaylist from '../locales/fr/playlist.json';
import frFilters from '../locales/fr/filters.json';
import frDownloads from '../locales/fr/downloads.json';
import frLogs from '../locales/fr/logs.json';
import frDialogs from '../locales/fr/dialogs.json';

export const LANGUAGE_STORAGE_KEY = 'stalkeer_language';

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    fallbackLng: 'en',
    supportedLngs: ['en', 'fr'],
    defaultNS: 'common',
    ns: ['common', 'playlist', 'filters', 'downloads', 'logs', 'dialogs'],
    resources: {
      en: {
        common: enCommon,
        playlist: enPlaylist,
        filters: enFilters,
        downloads: enDownloads,
        logs: enLogs,
        dialogs: enDialogs,
      },
      fr: {
        common: frCommon,
        playlist: frPlaylist,
        filters: frFilters,
        downloads: frDownloads,
        logs: frLogs,
        dialogs: frDialogs,
      },
    },
    detection: {
      order: ['localStorage'],
      lookupLocalStorage: LANGUAGE_STORAGE_KEY,
      caches: ['localStorage'],
    },
    interpolation: {
      escapeValue: false,
    },
  });

export default i18n;
