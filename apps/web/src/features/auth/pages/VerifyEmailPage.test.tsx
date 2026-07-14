import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { StrictMode } from 'react';
import { MemoryRouter, Route, Routes } from 'react-router';
import { afterEach, describe, expect, it, vi } from 'vitest';

import VerifyEmailPage from './VerifyEmailPage';

describe('VerifyEmailPage', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('submits an atomic verification token once under StrictMode', async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify({ message: 'Email verified successfully' }), {
        headers: { 'Content-Type': 'application/json' },
      }),
    );
    vi.stubGlobal('fetch', fetchMock);
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

    render(
      <StrictMode>
        <QueryClientProvider client={queryClient}>
          <MemoryRouter initialEntries={['/verify-email?token=single-use-token']}>
            <Routes>
              <Route path="/verify-email" element={<VerifyEmailPage />} />
            </Routes>
          </MemoryRouter>
        </QueryClientProvider>
      </StrictMode>,
    );

    expect(await screen.findByRole('heading', { name: '邮箱验证成功' })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});
