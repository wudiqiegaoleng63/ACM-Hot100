import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router';

import { AuthProvider } from '@/features/auth/contexts/auth-context';
import { authKeys } from '@/lib/query-keys';
import { afterEach, describe, expect, it, vi } from 'vitest';

import SubmissionDetailPage from './SubmissionDetailPage';

vi.mock('@/features/editor/components/CodeEditor', () => ({
  default: ({ value, readOnly }: { value: string; readOnly: boolean }) => (
    <pre data-testid="submission-code" data-readonly={String(readOnly)}>{value}</pre>
  ),
}));

const testUser = {
  id: 'user-1', email: 'user@example.local', username: 'learner',
  email_verified_at: '2026-07-13T00:00:00Z', status: 'ACTIVE', created_at: '2026-07-13T00:00:00Z',
};

const detail = {
  id: 'submission-1',
  problem_slug: 'two-sum-target',
  problem_title: '两数目标和',
  language_key: 'cpp17',
  source_code: 'int main() {}',
  status: 'CE',
  passed_cases: 0,
  total_cases: 0,
  time_ms: 0,
  memory_kb: 0,
  compiler_output: "error: expected ';'",
  error_message: '',
  case_results: [],
  created_at: '2026-07-14T00:00:00Z',
  judged_at: '2026-07-14T00:00:01Z',
};

describe('SubmissionDetailPage', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('renders owned submission details, code, and safe diagnostics', async () => {
    vi.stubGlobal('fetch', detailFetch(detail));
    renderPage();

    expect(await screen.findByText('两数目标和')).toBeInTheDocument();
    expect(screen.getByText('编译错误')).toBeInTheDocument();
    expect(screen.getByText(/error: expected/)).toBeInTheDocument();
    expect(screen.getByTestId('submission-code')).toHaveTextContent('int main() {}');
    expect(screen.getByTestId('submission-code')).toHaveAttribute('data-readonly', 'true');
  });

  it('copies code and returns to the problem with editable source state', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } });
    vi.stubGlobal('fetch', detailFetch(detail));
    renderPage();

    fireEvent.click(await screen.findByRole('button', { name: '复制代码' }));
    await waitFor(() => expect(writeText).toHaveBeenCalledWith('int main() {}'));
    expect(screen.getByRole('link', { name: '回到题目继续修改' })).toHaveAttribute('href', '/problems/two-sum-target');
  });

  it('uses one not-found message for missing or unauthorized submissions', async () => {
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockImplementation((input) => {
      if (String(input) === '/api/v1/auth/me') return Promise.resolve(jsonResponse({ user: testUser }));
      return Promise.resolve(new Response(JSON.stringify({
        error: { code: 'NOT_FOUND', message: 'submission not found' }, request_id: 'request-1',
      }), { status: 404, headers: { 'Content-Type': 'application/json' } }));
    }));
    renderPage();

    expect(await screen.findByText('提交不存在或无权查看')).toBeInTheDocument();
  });
});

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  queryClient.setQueryData(authKeys.me, testUser);
  return render(
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <MemoryRouter initialEntries={['/submissions/submission-1']}>
          <Routes><Route path="/submissions/:id" element={<SubmissionDetailPage />} /></Routes>
        </MemoryRouter>
      </AuthProvider>
    </QueryClientProvider>,
  );
}

function detailFetch(payload: unknown) {
  return vi.fn<typeof fetch>().mockImplementation((input) => {
    const url = String(input);
    if (url === '/api/v1/auth/me') return Promise.resolve(jsonResponse({ user: testUser }));
    if (url === '/api/v1/submissions/submission-1') return Promise.resolve(jsonResponse(payload));
    if (url === '/api/v1/languages') return Promise.resolve(jsonResponse([
      { key: 'cpp17', display_name: 'C++17', editor_language: 'cpp', source_template: 'template' },
    ]));
    return Promise.reject(new Error(`unexpected request: ${url}`));
  });
}

function jsonResponse(payload: unknown) {
  return new Response(JSON.stringify(payload), { headers: { 'Content-Type': 'application/json' } });
}
