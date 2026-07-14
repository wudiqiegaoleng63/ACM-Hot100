import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { useSampleRun } from '@/features/problems/hooks/use-problems';

const mockRunQueued = {
  id: 'run-1',
  language_key: 'cpp17',
  sample_case_id: 'sample-1',
  input_data: '1 2\n',
  status: 'QUEUED' as const,
  output_data: '',
  error_message: '',
  created_at: '2026-07-14T00:00:00Z',
  updated_at: '2026-07-14T00:00:00Z',
  started_at: null,
  finished_at: null,
  expires_at: '2026-07-15T00:00:00Z',
};

const mockRunAC = {
  ...mockRunQueued,
  status: 'AC' as const,
  output_data: '3\n',
  finished_at: '2026-07-14T00:00:01Z',
};

describe('useSampleRun', () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('returns undefined when runID is null', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useSampleRun(null), {
      wrapper: ({ children }) => <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>,
    });

    expect(result.current.data).toBeUndefined();
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches the sample run when runID is provided', async () => {
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify(mockRunQueued), { headers: { 'Content-Type': 'application/json' } }),
    ));
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useSampleRun('run-1'), {
      wrapper: ({ children }) => <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>,
    });

    await waitFor(() => expect(result.current.data?.status).toBe('QUEUED'));
  });

  it('stops polling when the run reaches a terminal status', async () => {
    let callCount = 0;
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockImplementation(() => {
      callCount++;
      return Promise.resolve(
        new Response(JSON.stringify(mockRunAC), { headers: { 'Content-Type': 'application/json' } }),
      );
    }));
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useSampleRun('run-1'), {
      wrapper: ({ children }) => <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>,
    });

    await waitFor(() => expect(result.current.data?.status).toBe('AC'));

    // Wait a bit to ensure no more fetches happen
    await act(async () => new Promise((resolve) => setTimeout(resolve, 200)));
    const callsAfterTerminal = callCount;

    await act(async () => new Promise((resolve) => setTimeout(resolve, 500)));
    expect(callCount).toBe(callsAfterTerminal);
  });

  it('cancels polling when the hook unmounts', async () => {
    vi.useFakeTimers();
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify(mockRunQueued), { headers: { 'Content-Type': 'application/json' } }),
    );
    vi.stubGlobal('fetch', fetchMock);
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { unmount } = renderHook(() => useSampleRun('run-1'), {
      wrapper: ({ children }) => <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>,
    });
    await act(async () => { await vi.advanceTimersByTimeAsync(1); });
    const callsBeforeUnmount = fetchMock.mock.calls.length;

    unmount();
    await act(async () => { await vi.advanceTimersByTimeAsync(5_000); });
    expect(fetchMock).toHaveBeenCalledTimes(callsBeforeUnmount);
  });
});
