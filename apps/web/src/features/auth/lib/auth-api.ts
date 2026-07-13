import { api } from '@/lib/api-client';

// --- Types ---

export interface User {
  id: string;
  email: string;
  username: string;
  email_verified: boolean;
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
  return api.post('/auth/reset-password', { token, password });
}

export function getCurrentUser(): Promise<User> {
  return api.get('/auth/me');
}
