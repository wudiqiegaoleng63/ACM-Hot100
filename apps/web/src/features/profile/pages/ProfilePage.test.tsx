import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import ProfilePage from './ProfilePage';

describe('ProfilePage', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('renders real totals, three-state distribution, and stage groups', async () => {
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockImplementation((input) => {
      const url = String(input);
      if (url === '/api/v1/profile/summary') {
        return Promise.resolve(jsonResponse({ total_problems: 5, solved: 2, attempted: 1, not_started: 2 }));
      }
      if (url === '/api/v1/profile/progress-by-stage') {
        return Promise.resolve(jsonResponse([
          { stage: '数组与哈希', total: 2, solved: 1, attempted: 1, not_started: 0 },
          { stage: '树与图', total: 1, solved: 0, attempted: 0, not_started: 1 },
        ]));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    }));
    renderPage();

    expect(await screen.findByText('2 / 5')).toBeInTheDocument();
    expect(screen.getByText('已通过')).toBeInTheDocument();
    expect(screen.getByText('尝试中')).toBeInTheDocument();
    expect(screen.getByText('未开始')).toBeInTheDocument();
    expect(screen.getByText('数组与哈希')).toBeInTheDocument();
    expect(screen.getByText('树与图')).toBeInTheDocument();
    expect(screen.queryByText(/连续/)).not.toBeInTheDocument();
    expect(screen.queryByText(/排名/)).not.toBeInTheDocument();
  });
});

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={queryClient}><ProfilePage /></QueryClientProvider>);
}

function jsonResponse(payload: unknown) {
  return new Response(JSON.stringify(payload), { headers: { 'Content-Type': 'application/json' } });
}
