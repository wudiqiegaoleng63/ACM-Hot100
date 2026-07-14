import { afterEach, describe, expect, it, vi } from 'vitest';

import {
  createSampleRun,
  createSubmission,
  getDraft,
  getHealth,
  getLanguages,
  getProblem,
  getProblems,
  getSampleRun,
  getSubmission,
  isTerminalRunStatus,
  isTerminalSubmissionStatus,
  saveDraft,
} from './problems-api';

describe('problems API contract', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('parses the backend problem list DTO without field guessing', async () => {
    const payload = {
      items: [
        {
          id: 'problem-1',
          slug: 'two-sum-target',
          order_index: 1,
          title: '两数目标和',
          difficulty: 'EASY',
          tags: [{ slug: 'array', name: '数组' }],
          progress_state: 'NOT_STARTED',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    };
    vi.stubGlobal('fetch', jsonFetch(payload));

    await expect(getProblems()).resolves.toEqual(payload);
  });

  it('parses the complete problem detail DTO with KB memory units', async () => {
    const payload = {
      id: 'problem-1',
      slug: 'two-sum-target',
      order_index: 1,
      title: '两数目标和',
      difficulty: 'EASY',
      stage: 'hot100',
      tags: [{ slug: 'array', name: '数组' }],
      progress_state: 'ATTEMPTED',
      statement_md: '题面',
      input_format_md: '输入',
      output_format_md: '输出',
      constraints_md: '范围',
      hints_md: '提示',
      time_limit_ms: 1000,
      memory_limit_kb: 262144,
      sample_cases: [
        {
          id: 'sample-1',
          order_index: 1,
          input_data: '4 9\n2 7 11 15\n',
          expected_output: '1 2\n',
          explanation_md: '解释',
        },
      ],
    };
    vi.stubGlobal('fetch', jsonFetch(payload));

    await expect(getProblem('two-sum-target')).resolves.toEqual(payload);
  });

  it('rejects enum casing that differs from the backend contract', async () => {
    vi.stubGlobal(
      'fetch',
      jsonFetch({
        items: [
          {
            id: 'problem-1',
            slug: 'two-sum-target',
            order_index: 1,
            title: '两数目标和',
            difficulty: 'easy',
            tags: [],
            progress_state: 'not_started',
          },
        ],
        total: 1,
        page: 1,
        page_size: 20,
      }),
    );

    await expect(getProblems()).rejects.toThrow();
  });

  it('parses draft source, language, and update timestamp', async () => {
    const payload = {
      source_code: 'int main() {}',
      language_key: 'cpp17',
      updated_at: '2026-07-13T14:30:00.123456Z',
    };
    vi.stubGlobal('fetch', jsonFetch(payload));

    await expect(getDraft('two-sum-target', 'cpp17')).resolves.toEqual(payload);
  });

  it('sends draft source and parses the saved draft response', async () => {
    const payload = {
      source_code: 'print(1)',
      language_key: 'python3',
      updated_at: '2026-07-13T14:31:00Z',
    };
    const fetchMock = jsonFetch(payload);
    vi.stubGlobal('fetch', fetchMock);

    await expect(saveDraft('two-sum-target', 'python3', 'print(1)')).resolves.toEqual(payload);
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/problems/two-sum-target/drafts/python3',
      expect.objectContaining({ body: JSON.stringify({ source_code: 'print(1)' }) }),
    );
  });

  it('parses only the public language configuration fields', async () => {
    const payload = [
      {
        key: 'cpp17',
        display_name: 'C++17',
        editor_language: 'cpp',
        source_template: 'int main() {}',
      },
    ];
    vi.stubGlobal('fetch', jsonFetch(payload));

    await expect(getLanguages()).resolves.toEqual(payload);
  });

  it('creates a sample run and parses the response with QUEUED status', async () => {
    const payload = {
      id: 'run-1',
      language_key: 'cpp17',
      sample_case_id: 'sample-1',
      input_data: '1 2\n',
      status: 'QUEUED',
      output_data: '',
      error_message: '',
      created_at: '2026-07-14T00:00:00Z',
      updated_at: '2026-07-14T00:00:00Z',
      started_at: null,
      finished_at: null,
      expires_at: '2026-07-15T00:00:00Z',
    };
    const fetchMock = jsonFetch(payload);
    vi.stubGlobal('fetch', fetchMock);

    await expect(
      createSampleRun('two-sum-target', {
        language_key: 'cpp17',
        source_code: 'int main() {}',
        sample_case_id: 'sample-1',
      }),
    ).resolves.toEqual(payload);
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/problems/two-sum-target/run',
      expect.objectContaining({ method: 'POST' }),
    );
  });

  it('polls a sample run and parses the AC terminal status', async () => {
    const payload = {
      id: 'run-1',
      language_key: 'cpp17',
      sample_case_id: null,
      input_data: 'custom',
      status: 'AC',
      output_data: '3\n',
      error_message: '',
      created_at: '2026-07-14T00:00:00Z',
      updated_at: '2026-07-14T00:00:01Z',
      started_at: '2026-07-14T00:00:00.5Z',
      finished_at: '2026-07-14T00:00:01Z',
      expires_at: '2026-07-15T00:00:00Z',
    };
    vi.stubGlobal('fetch', jsonFetch(payload));

    await expect(getSampleRun('run-1')).resolves.toEqual(payload);
  });

  it('rejects an unknown sample run status', async () => {
    const payload = {
      id: 'run-1',
      language_key: 'cpp17',
      sample_case_id: null,
      input_data: '',
      status: 'UNKNOWN',
      output_data: '',
      error_message: '',
      created_at: '2026-07-14T00:00:00Z',
      updated_at: '2026-07-14T00:00:00Z',
      started_at: null,
      finished_at: null,
      expires_at: '2026-07-15T00:00:00Z',
    };
    vi.stubGlobal('fetch', jsonFetch(payload));

    await expect(getSampleRun('run-1')).rejects.toThrow();
  });

  it('identifies AC and SYSTEM_ERROR as terminal statuses', () => {
    expect(isTerminalRunStatus('AC')).toBe(true);
    expect(isTerminalRunStatus('SYSTEM_ERROR')).toBe(true);
    expect(isTerminalRunStatus('QUEUED')).toBe(false);
    expect(isTerminalRunStatus('RUNNING')).toBe(false);
  });

  it('creates a formal submission using the validated QUEUED response contract', async () => {
    const payload = {
      id: 'submission-1',
      status: 'QUEUED',
      created_at: '2026-07-14T00:00:00Z',
    };
    const fetchMock = jsonFetch(payload);
    vi.stubGlobal('fetch', fetchMock);

    await expect(createSubmission('two-sum-target', 'cpp17', 'int main() {}')).resolves.toEqual(payload);
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/problems/two-sum-target/submissions',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ language_key: 'cpp17', source_code: 'int main() {}' }),
      }),
    );
  });

  it('parses a formal submission without hidden case output', async () => {
    const payload = submissionPayload('WA');
    vi.stubGlobal('fetch', jsonFetch(payload));

    await expect(getSubmission('submission-1')).resolves.toEqual(payload);
    expect(payload.case_results[0]).not.toHaveProperty('actual_output');
  });

  it('identifies every formal terminal status', () => {
    for (const status of ['AC', 'WA', 'TLE', 'MLE', 'RE', 'CE', 'SYSTEM_ERROR'] as const) {
      expect(isTerminalSubmissionStatus(status)).toBe(true);
    }
    for (const status of ['QUEUED', 'COMPILING', 'RUNNING'] as const) {
      expect(isTerminalSubmissionStatus(status)).toBe(false);
    }
  });

  it('rejects an invalid formal submission status', async () => {
    vi.stubGlobal('fetch', jsonFetch({ ...submissionPayload('WA'), status: 'PENDING' }));

    await expect(getSubmission('submission-1')).rejects.toThrow();
  });

  it('parses the health response with judge_mode', async () => {
    const payload = {
      status: 'ok',
      services: { mysql: 'ok', redis: 'ok' },
      judge_mode: 'mock',
    };
    vi.stubGlobal('fetch', jsonFetch(payload));

    await expect(getHealth()).resolves.toEqual(payload);
  });
});

function submissionPayload(status: 'WA' | 'AC') {
  return {
    id: 'submission-1',
    problem_slug: 'two-sum-target',
    problem_title: '两数目标和',
    language_key: 'cpp17',
    source_code: 'int main() {}',
    status,
    passed_cases: status === 'AC' ? 8 : 2,
    total_cases: 8,
    time_ms: 42,
    memory_kb: 2048,
    compiler_output: '',
    error_message: '',
    case_results: [{ case_index: 2, status, time_ms: 10, memory_kb: 1024, is_sample: false }],
    created_at: '2026-07-14T00:00:00Z',
    judged_at: '2026-07-14T00:00:01Z',
  };
}

function jsonFetch(payload: unknown) {
  return vi.fn<typeof fetch>().mockResolvedValue(
    new Response(JSON.stringify(payload), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }),
  );
}
