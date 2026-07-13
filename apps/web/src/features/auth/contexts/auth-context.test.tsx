import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { AuthProvider, ProtectedRoute } from './auth-context';

describe('authentication route protection', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('waits for auth/me before redirecting a protected route', async () => {
    let resolveMe: ((response: Response) => void) | undefined;
    const meResponse = new Promise<Response>((resolve) => {
      resolveMe = resolve;
    });
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockImplementation(() => meResponse));

    renderProtectedRoute();
    expect(screen.getByText('正在恢复登录状态…')).toBeInTheDocument();
    expect(screen.queryByText('登录页面')).not.toBeInTheDocument();

    resolveMe?.(
      new Response(
        JSON.stringify({
          error: { code: 'UNAUTHORIZED', message: 'Authentication required' },
          request_id: 'request-auth',
        }),
        {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        },
      ),
    );
    expect(await screen.findByText('登录页面')).toBeInTheDocument();
  });

  it('renders protected content when auth/me returns a user envelope', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(
          JSON.stringify({
            user: {
              id: 'user-1',
              email: 'user@example.local',
              username: 'learner',
              email_verified_at: '2026-07-13T00:00:00Z',
              status: 'ACTIVE',
              created_at: '2026-07-13T00:00:00Z',
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        ),
      ),
    );
    renderProtectedRoute();
    expect(await screen.findByText('受保护内容')).toBeInTheDocument();
  });
});

function renderProtectedRoute() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <MemoryRouter initialEntries={['/profile']}>
          <Routes>
            <Route
              path="/profile"
              element={(
                <ProtectedRoute>
                  <div>受保护内容</div>
                </ProtectedRoute>
              )}
            />
            <Route path="/login" element={<div>登录页面</div>} />
          </Routes>
        </MemoryRouter>
      </AuthProvider>
    </QueryClientProvider>,
  );
}
