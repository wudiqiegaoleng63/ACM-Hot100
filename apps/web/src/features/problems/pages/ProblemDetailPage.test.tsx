import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { createMemoryRouter, RouterProvider } from 'react-router';

import { AuthProvider } from '@/features/auth/contexts/auth-context';
import { afterEach, describe, expect, it, vi } from 'vitest';

import ProblemDetailPage from './ProblemDetailPage';

describe('ProblemDetailPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('renders Markdown math and public samples from the API', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn<typeof fetch>().mockImplementation((input) => {
        if (String(input) === '/api/v1/auth/me') return Promise.resolve(unauthorizedResponse());
        if (String(input) === '/api/v1/auth/refresh') return Promise.resolve(unauthorizedResponse());
        if (String(input) === '/api/v1/languages') return Promise.resolve(jsonResponse([]));
        if (String(input).endsWith('/navigation')) {
          return Promise.resolve(jsonResponse({ prev: null, next: null }));
        }
        return Promise.resolve(jsonResponse(problemPayload));
      }),
    );

    renderPage();
    expect(screen.getByText('正在加载题目…')).toBeInTheDocument();
    expect(await screen.findByRole('heading', { name: '两数目标和' })).toBeInTheDocument();
    expect(document.querySelector('.katex')).not.toBeNull();
    expect(screen.getByText('样例 1')).toBeInTheDocument();
    expect(screen.getByText('256 MB')).toBeInTheDocument();
  });

  it('renders an error state when the problem request fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn<typeof fetch>().mockImplementation((input) => {
        if (String(input) === '/api/v1/auth/me') return Promise.resolve(unauthorizedResponse());
        if (String(input) === '/api/v1/auth/refresh') return Promise.resolve(unauthorizedResponse());
        return Promise.resolve(jsonResponse({ error: { code: 'NOT_FOUND', message: 'not found' }, request_id: '1' }, 404));
      }),
    );

    renderPage();
    expect(await screen.findByText('题目加载失败')).toBeInTheDocument();
  });
});

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const router = createMemoryRouter(
    [{ path: '/problems/:slug', element: <ProblemDetailPage /> }],
    { initialEntries: ['/problems/two-sum-target'] },
  );
  return render(
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </QueryClientProvider>,
  );
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function unauthorizedResponse() {
  return jsonResponse({ error: { code: 'UNAUTHORIZED', message: 'Authentication required' }, request_id: 'auth' }, 401);
}

const problemPayload = {
  id: 'problem-1',
  slug: 'two-sum-target',
  order_index: 1,
  title: '两数目标和',
  difficulty: 'EASY',
  stage: 'hot100',
  tags: [{ slug: 'array', name: '数组' }],
  progress_state: null,
  statement_md: '给定 $n$ 个整数。',
  input_format_md: '输入整数 $n$。',
  output_format_md: '输出答案。',
  constraints_md: '$1 \\le n \\le 10^5$',
  hints_md: '',
  time_limit_ms: 1000,
  memory_limit_kb: 262144,
  sample_cases: [
    {
      id: 'sample-1',
      order_index: 1,
      input_data: '4 9\n2 7 11 15\n',
      expected_output: '1 2\n',
      explanation_md: '第 1 和第 2 个数之和为 9。',
    },
  ],
};
