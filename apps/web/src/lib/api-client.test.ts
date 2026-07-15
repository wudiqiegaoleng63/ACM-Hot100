import { afterEach, describe, expect, it, vi } from 'vitest';

import { ApiError, api } from './api-client';

describe('api client', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('sends filtered query parameters with credentials', async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify({ items: [{ id: 'problem-1' }] }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    );
    vi.stubGlobal('fetch', fetchMock);

    const result = await api.get<{ items: Array<{ id: string }> }>('/problems', {
      q: 'two sum',
      page: 2,
      difficulty: '',
      tag: undefined,
    });

    expect(result).toEqual({ items: [{ id: 'problem-1' }] });
    expect(fetchMock).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/problems?q=two+sum&page=2',
      expect.objectContaining({
        method: 'GET',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      }),
    );
  });

  it.each([
    '/auth/me',
    '/auth/login',
    '/auth/register',
    '/auth/refresh',
    '/auth/login?next=%2Fprofile',
    '/auth/login/',
    '/auth/verify-email',
    '/auth/resend-verification',
    '/auth/forgot-password',
    '/auth/reset-password',
  ])('does not refresh when %s returns 401', async (endpoint) => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          error: { code: 'UNAUTHORIZED', message: 'Unauthorized' },
          request_id: 'request-auth',
        }),
        {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        },
      ),
    );
    vi.stubGlobal('fetch', fetchMock);

    await expect(api.post(endpoint)).rejects.toBeInstanceOf(ApiError);
    expect(fetchMock).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenCalledWith(
      `/api/v1${endpoint}`,
      expect.objectContaining({ method: 'POST' }),
    );
  });

  it('shares one refresh for concurrent 401 responses and retries each request once', async () => {
    let resolveRefresh: ((response: Response) => void) | undefined;
    const refreshResponse = new Promise<Response>((resolve) => {
      resolveRefresh = resolve;
    });
    const attempts = new Map<string, number>();

    const fetchMock = vi.fn<typeof fetch>().mockImplementation((input) => {
      const url = String(input);
      if (url === '/api/v1/auth/refresh') {
        return refreshResponse;
      }

      const attempt = (attempts.get(url) ?? 0) + 1;
      attempts.set(url, attempt);
      if (attempt === 1) {
        return Promise.resolve(new Response(null, { status: 401 }));
      }
      return Promise.resolve(
        new Response(JSON.stringify({ url }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      );
    });
    vi.stubGlobal('fetch', fetchMock);

    const requests = [api.get('/problems'), api.get('/profile/summary')];
    await vi.waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    resolveRefresh?.(new Response(null, { status: 204 }));

    await expect(Promise.all(requests)).resolves.toEqual([
      { url: '/api/v1/problems' },
      { url: '/api/v1/profile/summary' },
    ]);
    expect(
      fetchMock.mock.calls.filter(([input]) => String(input) === '/api/v1/auth/refresh'),
    ).toHaveLength(1);
    expect(attempts).toEqual(
      new Map([
        ['/api/v1/problems', 2],
        ['/api/v1/profile/summary', 2],
      ]),
    );
  });

  it('does not start another refresh for a concurrent request whose 401 arrives late', async () => {
    let resolveSlowRequest: ((response: Response) => void) | undefined;
    const slowResponse = new Promise<Response>((resolve) => {
      resolveSlowRequest = resolve;
    });
    const attempts = new Map<string, number>();

    const fetchMock = vi.fn<typeof fetch>().mockImplementation((input) => {
      const url = String(input);
      if (url === '/api/v1/auth/refresh') {
        return Promise.resolve(new Response(null, { status: 204 }));
      }

      const attempt = (attempts.get(url) ?? 0) + 1;
      attempts.set(url, attempt);
      if (url === '/api/v1/profile/summary' && attempt === 1) {
        return slowResponse;
      }
      if (attempt === 1) {
        return Promise.resolve(new Response(null, { status: 401 }));
      }
      return Promise.resolve(
        new Response(JSON.stringify({ url }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      );
    });
    vi.stubGlobal('fetch', fetchMock);

    const fastRequest = api.get('/problems');
    const slowRequest = api.get('/profile/summary');
    await expect(fastRequest).resolves.toEqual({ url: '/api/v1/problems' });
    resolveSlowRequest?.(new Response(null, { status: 401 }));
    await expect(slowRequest).resolves.toEqual({ url: '/api/v1/profile/summary' });

    expect(
      fetchMock.mock.calls.filter(([input]) => String(input) === '/api/v1/auth/refresh'),
    ).toHaveLength(1);
  });

  it('does not repeat a failed refresh for a concurrent request whose 401 arrives late', async () => {
    let resolveSlowRequest: ((response: Response) => void) | undefined;
    const slowResponse = new Promise<Response>((resolve) => {
      resolveSlowRequest = resolve;
    });

    const fetchMock = vi.fn<typeof fetch>().mockImplementation((input) => {
      const url = String(input);
      if (url === '/api/v1/auth/refresh') {
        return Promise.resolve(new Response(null, { status: 401 }));
      }
      if (url === '/api/v1/profile/summary') {
        return slowResponse;
      }
      return Promise.resolve(new Response(null, { status: 401 }));
    });
    vi.stubGlobal('fetch', fetchMock);

    const fastRequest = api.get('/problems');
    const slowRequest = api.get('/profile/summary');
    await expect(fastRequest).rejects.toBeInstanceOf(ApiError);
    resolveSlowRequest?.(new Response(null, { status: 401 }));
    await expect(slowRequest).rejects.toBeInstanceOf(ApiError);

    expect(
      fetchMock.mock.calls.filter(([input]) => String(input) === '/api/v1/auth/refresh'),
    ).toHaveLength(1);
  });

  it('does not replay a delayed protected request after the session changes', async () => {
    let resolveProtected: ((response: Response) => void) | undefined;
    const protectedResponse = new Promise<Response>((resolve) => {
      resolveProtected = resolve;
    });

    const fetchMock = vi.fn<typeof fetch>().mockImplementation((input) => {
      const url = String(input);
      if (url === '/api/v1/problems') {
        return protectedResponse;
      }
      if (url === '/api/v1/auth/login') {
        return Promise.resolve(
          new Response(JSON.stringify({ message: 'Login successful' }), {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          }),
        );
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });
    vi.stubGlobal('fetch', fetchMock);

    const protectedRequest = api.get('/problems');
    await expect(
      api.post('/auth/login', { email: 'new@example.com', password: 'password' }),
    ).resolves.toEqual({ message: 'Login successful' });
    resolveProtected?.(
      new Response(
        JSON.stringify({
          error: { code: 'UNAUTHORIZED', message: 'Unauthorized' },
          request_id: 'request-before-login',
        }),
        {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        },
      ),
    );

    await expect(protectedRequest).rejects.toBeInstanceOf(ApiError);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('does not retry a request more than once after refresh', async () => {
    const fetchMock = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(new Response(null, { status: 401 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            error: { code: 'UNAUTHORIZED', message: 'Unauthorized' },
            request_id: 'request-retry',
          }),
          {
            status: 401,
            headers: { 'Content-Type': 'application/json' },
          },
        ),
      );
    vi.stubGlobal('fetch', fetchMock);

    await expect(api.get('/profile')).rejects.toBeInstanceOf(ApiError);
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });

  it('serializes request data and exposes structured API errors', async () => {
    const fetchMock = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ id: 'user-1' }), {
          status: 201,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            error: { code: 'VALIDATION_ERROR', message: '邮箱格式错误' },
            request_id: 'request-1',
          }),
          {
            status: 422,
            headers: { 'Content-Type': 'application/json' },
          },
        ),
      );
    vi.stubGlobal('fetch', fetchMock);

    await expect(
      api.post<{ id: string }>('/auth/register', { email: 'user@example.com' }),
    ).resolves.toEqual({ id: 'user-1' });
    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      '/api/v1/auth/register',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify({ email: 'user@example.com' }),
      }),
    );

    const request = api.get('/problems', { page: -1 });
    await expect(request).rejects.toBeInstanceOf(ApiError);
    await expect(request).rejects.toMatchObject({
      name: 'ApiError',
      code: 'VALIDATION_ERROR',
      message: '邮箱格式错误',
      requestId: 'request-1',
    });
  });
});
