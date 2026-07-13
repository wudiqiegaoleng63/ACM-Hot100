import { api } from '@/lib/api-client';

export const AUTH_ERROR_CODES = {
  emailAlreadyExists: 'EMAIL_ALREADY_EXISTS',
  usernameAlreadyExists: 'USERNAME_ALREADY_EXISTS',
  invalidCredentials: 'INVALID_CREDENTIALS',
  emailNotVerified: 'EMAIL_NOT_VERIFIED',
} as const;

// --- Types ---

export interface User {
  id: string;
  email: string;
  username: string;
  email_verified_at: string | null;
  status: 'PENDING' | 'ACTIVE' | 'DISABLED';
  created_at: string;
}

interface AuthResponse {
  message: string;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

// --- API functions ---

export function register(data: RegisterRequest): Promise<{ message: string }> {
  return api.post('/auth/register', data);
}

export function verifyEmail(token: string): Promise<{ message: string }> {
  return api.post('/auth/verify-email', { token });
}

export function resendVerification(email: string): Promise<{ message: string }> {
  return api.post('/auth/resend-verification', { email });
}

export function login(data: LoginRequest): Promise<AuthResponse> {
  return api.post('/auth/login', data);
}

export function logout(): Promise<void> {
  return api.post('/auth/logout');
}

export function logoutAll(): Promise<void> {
  return api.post('/auth/logout-all');
}

export function forgotPassword(email: string): Promise<{ message: string }> {
  return api.post('/auth/forgot-password', { email });
}

export function resetPassword(token: string, password: string): Promise<{ message: string }> {
  return api.post('/auth/reset-password', { token, new_password: password });
}

export async function getCurrentUser(): Promise<User> {
  const response = await api.get<{ user: User }>('/auth/me');
  return response.user;
}
