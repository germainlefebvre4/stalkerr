import { useTranslation } from 'react-i18next';
import { ApiError } from '../services/api';

export function useApiErrorMessage() {
  const { t } = useTranslation();

  return (err: unknown): string => {
    const code = err instanceof ApiError ? err.code : undefined;
    return t(code ? `errors.${code}` : 'errors.generic', { defaultValue: t('errors.generic') });
  };
}
