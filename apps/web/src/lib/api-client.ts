const BASE_URL = '/api';

interface RequestOptions extends Omit<RequestInit, 'body'> {
  data?: unknown;
  params?: Record<string, string>;
}

async function apiClient<T>(
  endpoint: string,
  options: RequestOptions = {},
): Promise<T> {
  const { data, params, ...init } = options;

  let url = `${BASE_URL}${endpoint}`;
  if (params) {
    const searchParams = new URLSearchParams(params);
    url += `?${searchParams.toString()}`;
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

  if (!response.ok) {
    const error = await response.json().catch(() => ({
      message: response.statusText,
    }));
    throw new Error(error.message || `HTTP ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json();
}

export const api = {
  get: <T>(endpoint: string, params?: Record<string, string>) =>
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
