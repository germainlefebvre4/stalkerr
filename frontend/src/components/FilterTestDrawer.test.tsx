import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { FilterTestDrawer } from './FilterTestDrawer';
import { FilterTestTarget, M3uSource } from '../types';

vi.mock('../services/api', () => ({
  api: {
    dryRunFilter: vi.fn(),
  },
  ApiError: class ApiError extends Error {
    code: string;
    constructor(code: string, message: string) {
      super(message);
      this.code = code;
    }
  },
}));

const sources: M3uSource[] = [];

describe('FilterTestDrawer', () => {
  afterEach(cleanup);

  it('renders nothing when target is null', () => {
    const { container } = render(
      <I18nextProvider i18n={i18n}>
        <FilterTestDrawer target={null} onOpenChange={vi.fn()} sources={sources} />
      </I18nextProvider>
    );
    expect(container).toBeEmptyDOMElement();
  });

  it('shows the target label as the title in single mode', () => {
    const target: FilterTestTarget = {
      mode: 'single',
      attribute: 'group_title',
      includePatterns: '',
      excludePatterns: '',
      label: 'Test: Group Title — Origin (config.yml)',
    };
    render(
      <I18nextProvider i18n={i18n}>
        <FilterTestDrawer target={target} onOpenChange={vi.fn()} sources={sources} />
      </I18nextProvider>
    );
    expect(screen.getByText('Test: Group Title — Origin (config.yml)')).toBeInTheDocument();
  });

  it('shows the target label as the title in combined mode', () => {
    const target: FilterTestTarget = {
      mode: 'combined',
      groupTitleInclude: '',
      groupTitleExclude: '',
      tvgNameInclude: '',
      tvgNameExclude: '',
      label: 'Test everything',
    };
    render(
      <I18nextProvider i18n={i18n}>
        <FilterTestDrawer target={target} onOpenChange={vi.fn()} sources={sources} />
      </I18nextProvider>
    );
    expect(screen.getByText('Test everything')).toBeInTheDocument();
  });
});
