import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/react';
import * as Tabs from '@radix-ui/react-tabs';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { ErrorsTab } from './ErrorsTab';
import { DownloadEnriched } from '../types';

afterEach(cleanup);

interface Overrides {
  reasonFilter?: string;
  setReasonFilter?: (reason: string) => void;
  errorsTotal?: number;
  errorsPage?: number;
  setErrorsPage?: (page: number | ((prev: number) => number)) => void;
  errorsLimit?: number;
  setErrorsLimit?: (limit: number) => void;
}

function renderErrorsTab(errors: DownloadEnriched[], overrides: Overrides = {}) {
  return render(
    <I18nextProvider i18n={i18n}>
      <Tabs.Root value="errors">
        <ErrorsTab
          errors={errors}
          errorsLoading={false}
          reasonFilter={overrides.reasonFilter ?? ''}
          setReasonFilter={overrides.setReasonFilter ?? (() => {})}
          errorsTotal={overrides.errorsTotal ?? errors.length}
          errorsPage={overrides.errorsPage ?? 1}
          setErrorsPage={overrides.setErrorsPage ?? (() => {})}
          errorsLimit={overrides.errorsLimit ?? 20}
          setErrorsLimit={overrides.setErrorsLimit ?? (() => {})}
          onFetchErrors={() => {}}
        />
      </Tabs.Root>
    </I18nextProvider>
  );
}

describe('ErrorsTab reason filter', () => {
  it('calls setReasonFilter with unknown_format when "Invalid extension" is selected', () => {
    const setReasonFilter = vi.fn();
    renderErrorsTab([], { setReasonFilter });

    fireEvent.change(screen.getByDisplayValue('All'), { target: { value: 'unknown_format' } });

    expect(setReasonFilter).toHaveBeenCalledWith('unknown_format');
  });

  it('calls setReasonFilter with missing_year when "Missing year" is selected', () => {
    const setReasonFilter = vi.fn();
    renderErrorsTab([], { setReasonFilter });

    fireEvent.change(screen.getByDisplayValue('All'), { target: { value: 'missing_year' } });

    expect(setReasonFilter).toHaveBeenCalledWith('missing_year');
  });

  it('calls setReasonFilter with year_mismatch when "Inconsistent year" is selected', () => {
    const setReasonFilter = vi.fn();
    renderErrorsTab([], { setReasonFilter });

    fireEvent.change(screen.getByDisplayValue('All'), { target: { value: 'year_mismatch' } });

    expect(setReasonFilter).toHaveBeenCalledWith('year_mismatch');
  });

  it('calls setReasonFilter with an empty value when returning to "All"', () => {
    const setReasonFilter = vi.fn();
    renderErrorsTab([], { setReasonFilter, reasonFilter: 'missing_year' });

    fireEvent.change(screen.getByDisplayValue('Missing year'), { target: { value: '' } });

    expect(setReasonFilter).toHaveBeenCalledWith('');
  });
});

describe('ErrorsTab pagination', () => {
  it('renders pagination controls below the table when there are errors', () => {
    renderErrorsTab([], { errorsTotal: 45, errorsPage: 1, errorsLimit: 20 });

    expect(screen.getByTitle('Next page')).toBeInTheDocument();
    expect(screen.getByDisplayValue('20')).toBeInTheDocument();
  });

  it('navigating to another page calls setErrorsPage with the target page and the expected offset semantics', () => {
    const setErrorsPage = vi.fn();
    renderErrorsTab([], { errorsTotal: 45, errorsPage: 1, errorsLimit: 20, setErrorsPage });

    fireEvent.click(screen.getByRole('button', { name: '2' }));

    expect(setErrorsPage).toHaveBeenCalledWith(2);
  });
});
