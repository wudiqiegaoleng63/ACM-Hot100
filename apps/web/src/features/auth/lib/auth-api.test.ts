import { afterEach, describe, expect, it, vi } from 'vitest';

import { resetPassword } from './auth-api';

describe('auth API contract', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('sends the reset password field expected by the backend', async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify({ message: 'Password reset successfully' }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    );
    vi.stubGlobal('fetch', fetchMock);

    await resetPassword('reset-token', 'NewPassword123!');

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/auth/reset-password',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          token: 'reset-token',
          new_password: 'NewPassword123!',
        }),
      }),
    );
  });
});
