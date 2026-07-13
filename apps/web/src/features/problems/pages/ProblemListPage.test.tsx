import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router';
import { afterEach, describe, expect, it, vi } from 'vitest';

import ProblemListPage from './ProblemListPage';

describe('ProblemListPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('renders loading and then an honest empty state', async () => {
    let resolveProblems: ((response: Response) => void) | undefined;
    const fetchMock = vi.fn<typeof fetch>().mockImplementation((input) => {
      if (String(input).startsWith('/api/v1/problems')) {
        return new Promise<Response>((resolve) => {
          resolveProblems = resolve;
        });
      }
      if (String(input) === '/api/v1/tags') {
        return Promise.resolve(jsonResponse([]));
      }
      return Promise.reject(new Error(`unexpected request ${String(input)}`));
    });
    vi.stubGlobal('fetch', fetchMock);

    renderPage(['/problems']);
    expect(screen.getByText('正在加载题单…')).toBeInTheDocument();
    resolveProblems?.(jsonResponse({ items: [], total: 0, page: 1, page_size: 20 }));
    expect(await screen.findByText('没有符合条件的题目')).toBeInTheDocument();
  });

  it('renders an actionable error state', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn<typeof fetch>().mockImplementation((input) => {
        if (String(input).startsWith('/api/v1/problems')) {
          return Promise.resolve(jsonResponse({ error: { code: 'ERROR', message: '失败' }, request_id: '1' }, 500));
        }
        return Promise.resolve(jsonResponse([]));
      }),
    );

    renderPage(['/problems']);
    expect(await screen.findByText('题单加载失败')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '重新加载' })).toBeInTheDocument();
  });

  it('synchronizes search and difficulty filters to the URL', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn<typeof fetch>().mockImplementation((input) => {
        if (String(input) === '/api/v1/tags') return Promise.resolve(jsonResponse([]));
        return Promise.resolve(jsonResponse({ items: [], total: 0, page: 1, page_size: 20 }));
      }),
    );
    renderPage(['/problems'], true);

    fireEvent.change(await screen.findByLabelText('搜索题目'), { target: { value: '两数' } });
    fireEvent.change(screen.getByLabelText('难度筛选'), { target: { value: 'EASY' } });

    await waitFor(() => {
      expect(screen.getByTestId('current-search')).toHaveTextContent('q=%E4%B8%A4%E6%95%B0');
      expect(screen.getByTestId('current-search')).toHaveTextContent('difficulty=EASY');
    });
  });
});

function renderPage(initialEntries: string[], showLocation = false) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={initialEntries}>
        <ProblemListPage />
        {showLocation && <LocationState />}
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

function LocationState() {
  const location = useLocation();
  return <output data-testid="current-search">{location.search}</output>;
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}
