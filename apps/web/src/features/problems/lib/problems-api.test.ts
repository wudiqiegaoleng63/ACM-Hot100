import { afterEach, describe, expect, it, vi } from 'vitest';

import {
  getLanguages,
  getProblem,
  getProblems,
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
});

function jsonFetch(payload: unknown) {
  return vi.fn<typeof fetch>().mockResolvedValue(
    new Response(JSON.stringify(payload), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }),
  );
}
