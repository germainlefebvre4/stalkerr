import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { PlaylistGroupedView } from './PlaylistGroupedView';
import { MediaGroupItem, PlaylistItem } from '../types';
import { api } from '../services/api';
import { useIsMobile } from '../hooks/useMediaQuery';

vi.mock('../services/api', () => ({
  api: {
    getGroupItems: vi.fn(),
  },
}));

vi.mock('../hooks/useMediaQuery', () => ({
  useIsMobile: vi.fn(() => false),
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

const group: MediaGroupItem = {
  type: 'movie',
  movie_id: 1,
  title: 'Test Movie',
  year: 2020,
  latest_activity: '2026-01-01T00:00:00Z',
};

const item: PlaylistItem = {
  id: 42,
  tvg_name: 'Test Movie Item',
  group_title: 'Movies',
  content_type: 'movies',
  state: 'pending',
  line_content: 'raw-line',
  line_hash: 'hash-1',
  line_number: 1,
  created_at: '2026-01-01T00:00:00Z',
  downloaded_at: null,
};

function renderView() {
  const onOpenOverride = vi.fn();
  const onResetPipeline = vi.fn();
  const onRowClick = vi.fn();
  const setGroupsPage = vi.fn();
  render(
    <I18nextProvider i18n={i18n}>
      <PlaylistGroupedView
        groups={[group]}
        groupsLoading={false}
        groupsTotal={1}
        groupsPage={1}
        setGroupsPage={setGroupsPage}
        groupsLimit={10}
        onOpenOverride={onOpenOverride}
        onResetPipeline={onResetPipeline}
        onRowClick={onRowClick}
      />
    </I18nextProvider>
  );
  return { onOpenOverride, onResetPipeline, onRowClick };
}

async function expandGroup() {
  fireEvent.click(screen.getByText('Test Movie'));
  await screen.findByText('Test Movie Item');
}

describe('PlaylistGroupedView', () => {
  beforeEach(() => {
    vi.mocked(useIsMobile).mockReturnValue(false);
    vi.mocked(api.getGroupItems).mockResolvedValue({ data: [item], total: 1, limit: 10, offset: 0, total_pages: 1 });
  });

  it('expands a group on click and collapses it again on a second click', async () => {
    renderView();

    fireEvent.click(screen.getByText('Test Movie'));
    await screen.findByText('Test Movie Item');

    fireEvent.click(screen.getByText('Test Movie'));

    expect(screen.queryByText('Test Movie Item')).not.toBeInTheDocument();
  });

  it('calls onRowClick with the clicked item and keeps the group expanded', async () => {
    const { onRowClick } = renderView();
    await expandGroup();

    fireEvent.click(screen.getByText('Test Movie Item'));

    expect(onRowClick).toHaveBeenCalledWith(item);
    expect(screen.getByText('Test Movie Item')).toBeInTheDocument();
  });

  it('does not open the sidepanel when the Associate/Correct action button is clicked', async () => {
    const { onOpenOverride, onRowClick } = renderView();
    await expandGroup();

    fireEvent.click(screen.getByText('Associate 🔍'));

    expect(onOpenOverride).toHaveBeenCalledWith(item);
    expect(onRowClick).not.toHaveBeenCalled();
  });

  it('does not open the sidepanel when the Reset action button is clicked', async () => {
    const { onResetPipeline, onRowClick } = renderView();
    await expandGroup();

    fireEvent.click(screen.getByText('Reset ↻'));

    expect(onResetPipeline).toHaveBeenCalledWith(item.id, item.content_type);
    expect(onRowClick).not.toHaveBeenCalled();
  });
});
