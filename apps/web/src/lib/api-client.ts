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
let refreshGeneration = 0;
let latestRefreshResult = false;
let sessionGeneration = 0;

const AUTH_ENDPOINTS_WITHOUT_AUTO_REFRESH = new Set([
  '/auth/me',
  '/auth/login',
  '/auth/register',
  '/auth/refresh',
  '/auth/verify-email',
  '/auth/resend-verification',
  '/auth/forgot-password',
  '/auth/reset-password',
]);

const SESSION_BOUNDARY_ENDPOINTS = new Set([
  '/auth/login',
  '/auth/logout',
  '/auth/logout-all',
  '/auth/reset-password',
]);

function normalizedEndpoint(endpoint: string): string {
  const path = endpoint.split('?', 1)[0] ?? endpoint;
  return path.length > 1 ? path.replace(/\/+$/, '') : path;
}

async function refreshAccessToken(): Promise<boolean> {
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    try {
      const res = await fetch(`${BASE_URL}/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      });
      latestRefreshResult = res.ok;
      refreshGeneration += 1;
      return latestRefreshResult;
    } catch {
      latestRefreshResult = false;
      refreshGeneration += 1;
      return latestRefreshResult;
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
  const requestRefreshGeneration = refreshGeneration;
  const requestSessionGeneration = sessionGeneration;
  const endpointPath = normalizedEndpoint(endpoint);

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

  // Handle protected-request 401s with a single shared refresh.
  if (
    response.status === 401 &&
    !_retry &&
    !AUTH_ENDPOINTS_WITHOUT_AUTO_REFRESH.has(endpointPath) &&
    requestSessionGeneration === sessionGeneration
  ) {
    const refreshed =
      requestRefreshGeneration < refreshGeneration
        ? latestRefreshResult
        : await refreshAccessToken();
    if (refreshed) {
      return apiClient<T>(endpoint, { ...options, _retry: true });
    }
  }

  if (response.ok && SESSION_BOUNDARY_ENDPOINTS.has(endpointPath)) {
    sessionGeneration += 1;
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
