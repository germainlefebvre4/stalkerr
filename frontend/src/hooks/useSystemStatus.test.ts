import { afterEach, describe, expect, it, vi } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useSystemStatus } from './useSystemStatus';
import { api } from '../services/api';

vi.mock('../services/api', () => ({
  api: {
    getSystemStatus: vi.fn(),
  },
  ApiError: class ApiError extends Error {
    code: string;
    constructor(code: string, message: string) {
      super(message);
      this.code = code;
    }
  },
}));

describe('useSystemStatus', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('does not fetch on mount', () => {
    renderHook(() => useSystemStatus());
    expect(api.getSystemStatus).not.toHaveBeenCalled();
  });

  it('fetches only when fetchStatus is explicitly invoked', async () => {
    const response = {
      database: { status: 'ok' as const },
      radarr: { status: 'not_configured' as const },
      sonarr: { status: 'not_configured' as const },
      tmdb: { status: 'not_configured' as const },
      disk: [],
    };
    vi.mocked(api.getSystemStatus).mockResolvedValue(response);

    const { result } = renderHook(() => useSystemStatus());
    expect(api.getSystemStatus).not.toHaveBeenCalled();

    act(() => {
      result.current.fetchStatus();
    });

    await waitFor(() => expect(result.current.status).toEqual(response));
    expect(api.getSystemStatus).toHaveBeenCalledTimes(1);
  });
});
