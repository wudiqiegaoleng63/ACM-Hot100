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
