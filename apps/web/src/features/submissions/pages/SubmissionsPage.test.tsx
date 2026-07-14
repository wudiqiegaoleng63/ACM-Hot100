import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';

import { AuthProvider } from '@/features/auth/contexts/auth-context';
import { authKeys } from '@/lib/query-keys';
import { afterEach, describe, expect, it, vi } from 'vitest';

import SubmissionsPage from './SubmissionsPage';

const testUser = {
  id: 'user-1', email: 'user@example.local', username: 'learner',
  email_verified_at: '2026-07-13T00:00:00Z', status: 'ACTIVE', created_at: '2026-07-13T00:00:00Z',
};

const payload = {
  items: [
    {
      id: 'submission-1',
      problem_slug: 'two-sum-target',
      problem_title: '两数目标和',
      language_key: 'cpp17',
      status: 'AC',
      passed_cases: 8,
      total_cases: 8,
      time_ms: 42,
      memory_kb: 2048,
      created_at: '2026-07-14T00:00:00Z',
    },
  ],
  total: 21,
  page: 1,
  page_size: 20,
};

describe('SubmissionsPage', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('renders the authenticated submission table and metrics', async () => {
    vi.stubGlobal('fetch', apiFetch(payload));
    renderPage();

    expect(await screen.findByText('两数目标和')).toBeInTheDocument();
    expect(screen.getAllByText('答案正确')).toHaveLength(2);
    expect(screen.getByText('42 ms')).toBeInTheDocument();
    expect(screen.getByText('2 MB')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '两数目标和' })).toHaveAttribute('href', '/submissions/submission-1');
  });

  it('synchronizes filters and pagination with the URL', async () => {
    const fetchMock = apiFetch(payload);
    vi.stubGlobal('fetch', fetchMock);
    renderPage('/submissions?status=WA&page=2');

    expect(await screen.findByLabelText('状态筛选')).toHaveValue('WA');
    fireEvent.change(screen.getByLabelText('语言筛选'), { target: { value: 'cpp17' } });

    await waitFor(() => {
      const urls = fetchMock.mock.calls.map(([input]) => String(input));
      expect(urls.some((url) => url.includes('status=WA') && url.includes('language=cpp17') && !url.includes('page=2'))).toBe(true);
    });
  });

  it('renders a real empty state', async () => {
    vi.stubGlobal('fetch', apiFetch({ ...payload, items: [], total: 0 }));
    renderPage();

    expect(await screen.findByText('暂无提交记录')).toBeInTheDocument();
  });
});

function renderPage(initialEntry = '/submissions') {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  queryClient.setQueryData(authKeys.me, testUser);
  return render(
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <MemoryRouter initialEntries={[initialEntry]}><SubmissionsPage /></MemoryRouter>
      </AuthProvider>
    </QueryClientProvider>,
  );
}

function apiFetch(submissions: unknown) {
  return vi.fn<typeof fetch>().mockImplementation((input) => {
    const url = String(input);
    if (url === '/api/v1/auth/me') return Promise.resolve(jsonResponse({ user: testUser }));
    if (url.startsWith('/api/v1/submissions')) return Promise.resolve(jsonResponse(submissions));
    if (url === '/api/v1/languages') return Promise.resolve(jsonResponse([
      { key: 'cpp17', display_name: 'C++17', editor_language: 'cpp', source_template: 'template' },
    ]));
    return Promise.reject(new Error(`unexpected request: ${url}`));
  });
}

function jsonResponse(payload: unknown) {
  return new Response(JSON.stringify(payload), { headers: { 'Content-Type': 'application/json' } });
}
