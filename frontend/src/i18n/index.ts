import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import enCommon from '../locales/en/common.json';
import enPlaylist from '../locales/en/playlist.json';
import enFilters from '../locales/en/filters.json';
import enDownloads from '../locales/en/downloads.json';
import enLogs from '../locales/en/logs.json';
import enDialogs from '../locales/en/dialogs.json';
import enRadarrSonarr from '../locales/en/radarrSonarr.json';
import enErrors from '../locales/en/errors.json';
import enHome from '../locales/en/home.json';
import enSettings from '../locales/en/settings.json';

import frCommon from '../locales/fr/common.json';
import frPlaylist from '../locales/fr/playlist.json';
import frFilters from '../locales/fr/filters.json';
import frDownloads from '../locales/fr/downloads.json';
import frLogs from '../locales/fr/logs.json';
import frDialogs from '../locales/fr/dialogs.json';
import frRadarrSonarr from '../locales/fr/radarrSonarr.json';
import frErrors from '../locales/fr/errors.json';
import frHome from '../locales/fr/home.json';
import frSettings from '../locales/fr/settings.json';

export const LANGUAGE_STORAGE_KEY = 'stalkeer_language';

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    fallbackLng: 'en',
    supportedLngs: ['en', 'fr'],
    defaultNS: 'common',
    ns: ['common', 'playlist', 'filters', 'downloads', 'logs', 'dialogs', 'radarrSonarr', 'errors', 'home', 'settings'],
    resources: {
      en: {
        common: enCommon,
        playlist: enPlaylist,
        filters: enFilters,
        downloads: enDownloads,
        logs: enLogs,
        dialogs: enDialogs,
        radarrSonarr: enRadarrSonarr,
        errors: enErrors,
        home: enHome,
        settings: enSettings,
      },
      fr: {
        common: frCommon,
        playlist: frPlaylist,
        filters: frFilters,
        downloads: frDownloads,
        logs: frLogs,
        dialogs: frDialogs,
        radarrSonarr: frRadarrSonarr,
        errors: frErrors,
        home: frHome,
        settings: frSettings,
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
