import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import {
  getSubmissionPollInterval,
  SUBMISSION_FAST_POLL_DURATION_MS,
  SUBMISSION_FAST_POLL_MS,
  SUBMISSION_MAX_POLL_MS,
  SUBMISSION_SLOW_POLL_MS,
  useSubmission,
} from '@/features/problems/hooks/use-problems';
import { problemKeys, progressKeys } from '@/lib/query-keys';

const queuedSubmission = submissionPayload('QUEUED');
const acceptedSubmission = submissionPayload('AC');

describe('formal submission polling', () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('uses the required polling cadence and stops for terminal states or at 60 seconds', () => {
    expect(getSubmissionPollInterval('QUEUED', 0)).toBe(SUBMISSION_FAST_POLL_MS);
    expect(getSubmissionPollInterval('RUNNING', SUBMISSION_FAST_POLL_DURATION_MS - 1)).toBe(SUBMISSION_FAST_POLL_MS);
    expect(getSubmissionPollInterval('RUNNING', SUBMISSION_FAST_POLL_DURATION_MS)).toBe(SUBMISSION_SLOW_POLL_MS);
    expect(getSubmissionPollInterval('RUNNING', SUBMISSION_MAX_POLL_MS)).toBe(false);
    for (const status of ['AC', 'WA', 'TLE', 'MLE', 'RE', 'CE', 'SYSTEM_ERROR'] as const) {
      expect(getSubmissionPollInterval(status, 1)).toBe(false);
    }
  });

  it('stops polling on a terminal response and invalidates AC progress caches', async () => {
    let callCount = 0;
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockImplementation(() => {
      callCount++;
      return Promise.resolve(jsonResponse(acceptedSubmission));
    }));
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    queryClient.setQueryData(problemKeys.list(), { items: [] });
    queryClient.setQueryData(progressKeys.all, { solved: 0 });
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useSubmission('submission-1'), { wrapper: wrapper(queryClient) });

    await waitFor(() => expect(result.current.data?.status).toBe('AC'));
    await waitFor(() => {
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: problemKeys.all });
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: progressKeys.all });
    });

    await act(async () => new Promise((resolve) => setTimeout(resolve, 900)));
    expect(callCount).toBe(1);
  });

  it('stops after 60 seconds and resumes when manually refreshed', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockResolvedValue(jsonResponse(queuedSubmission)));
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useSubmission('submission-1'), { wrapper: wrapper(queryClient) });

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1);
    });
    expect(result.current.data?.status).toBe('QUEUED');

    await act(async () => {
      await vi.advanceTimersByTimeAsync(SUBMISSION_MAX_POLL_MS);
    });
    expect(result.current.pollTimedOut).toBe(true);

    await act(async () => {
      await result.current.refresh();
    });
    expect(result.current.pollTimedOut).toBe(false);
  });
});

function wrapper(queryClient: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

function submissionPayload(status: 'QUEUED' | 'AC') {
  return {
    id: 'submission-1',
    problem_slug: 'two-sum-target',
    problem_title: '两数目标和',
    language_key: 'cpp17',
    source_code: 'int main() {}',
    status,
    passed_cases: status === 'AC' ? 8 : 0,
    total_cases: status === 'AC' ? 8 : 0,
    time_ms: status === 'AC' ? 42 : null,
    memory_kb: status === 'AC' ? 2048 : null,
    compiler_output: '',
    error_message: '',
    case_results: [],
    created_at: '2026-07-14T00:00:00Z',
    judged_at: status === 'AC' ? '2026-07-14T00:00:01Z' : null,
  };
}

function jsonResponse(payload: unknown) {
  return new Response(JSON.stringify(payload), { headers: { 'Content-Type': 'application/json' } });
}
