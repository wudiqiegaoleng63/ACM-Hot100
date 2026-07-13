const BASE_URL = '/api/v1';

// --- ApiError ---

export class ApiError extends Error {
  code: string;
  requestId: string;

  constructor(code: string, message: string, requestId: string) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.requestId = requestId;
  }
}

// --- Error response shape from backend ---

interface ApiErrorResponse {
  error: {
    code: string;
    message: string;
    details?: unknown;
  };
  request_id: string;
}

// --- Token refresh logic (single concurrent refresh) ---

let refreshPromise: Promise<boolean> | null = null;

async function refreshAccessToken(): Promise<boolean> {
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    try {
      const res = await fetch(`${BASE_URL}/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      });
      return res.ok;
    } catch {
      return false;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

// --- Core request options ---

interface RequestOptions extends Omit<RequestInit, 'body'> {
  data?: unknown;
  params?: Record<string, string | number | undefined>;
  _retry?: boolean;
}

// --- Core client ---

async function apiClient<T>(
  endpoint: string,
  options: RequestOptions = {},
): Promise<T> {
  const { data, params, _retry, ...init } = options;

  let url = `${BASE_URL}${endpoint}`;
  if (params) {
    const searchParams = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined && value !== '') {
        searchParams.set(key, String(value));
      }
    }
    const qs = searchParams.toString();
    if (qs) url += `?${qs}`;
  }

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...init.headers,
  };

  const config: RequestInit = {
    ...init,
    credentials: 'include',
    headers,
  };

  if (data !== undefined) {
    config.body = JSON.stringify(data);
  }

  const response = await fetch(url, config);

  // Handle 401 with automatic token refresh
  if (response.status === 401 && !_retry) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      return apiClient<T>(endpoint, { ...options, _retry: true });
    }
  }

  if (!response.ok) {
    let errorBody: ApiErrorResponse | null = null;
    try {
      errorBody = await response.json();
    } catch {
      // ignore parse error
    }

    if (errorBody?.error) {
      throw new ApiError(
        errorBody.error.code,
        errorBody.error.message,
        errorBody.request_id ?? '',
      );
    }

    throw new ApiError(
      String(response.status),
      response.statusText || `HTTP ${response.status}`,
      '',
    );
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json();
}

// --- Convenience methods ---

export const api = {
  get: <T>(endpoint: string, params?: Record<string, string | number | undefined>) =>
    apiClient<T>(endpoint, { method: 'GET', params }),

  post: <T>(endpoint: string, data?: unknown) =>
    apiClient<T>(endpoint, { method: 'POST', data }),

  put: <T>(endpoint: string, data?: unknown) =>
    apiClient<T>(endpoint, { method: 'PUT', data }),

  patch: <T>(endpoint: string, data?: unknown) =>
    apiClient<T>(endpoint, { method: 'PATCH', data }),

  delete: <T>(endpoint: string) =>
    apiClient<T>(endpoint, { method: 'DELETE' }),
};
