import { lazy, type ComponentType, type LazyExoticComponent, type ReactNode } from 'react';
import { createBrowserRouter, type RouteObject } from 'react-router';

import RootLayout from '@/app/layouts/RootLayout';
import { ProtectedRoute } from '@/features/auth/contexts/auth-context';

function lazyPage<T extends ComponentType<unknown>>(
  importFn: () => Promise<{ default: T }>,
): LazyExoticComponent<T> {
  return lazy(importFn);
}

function protectedPage(children: ReactNode) {
  return <ProtectedRoute>{children}</ProtectedRoute>;
}

const HomePage = lazyPage(() => import('@/app/pages/HomePage'));
const ProblemListPage = lazyPage(() => import('@/features/problems/pages/ProblemListPage'));
const ProblemDetailPage = lazyPage(() => import('@/features/problems/pages/ProblemDetailPage'));
const SubmissionsPage = lazyPage(() => import('@/features/submissions/pages/SubmissionsPage'));
const SubmissionDetailPage = lazyPage(() => import('@/features/submissions/pages/SubmissionDetailPage'));
const ProfilePage = lazyPage(() => import('@/features/profile/pages/ProfilePage'));
const LoginPage = lazyPage(() => import('@/features/auth/pages/LoginPage'));
const RegisterPage = lazyPage(() => import('@/features/auth/pages/RegisterPage'));
const VerifyEmailPage = lazyPage(() => import('@/features/auth/pages/VerifyEmailPage'));
const ForgotPasswordPage = lazyPage(() => import('@/features/auth/pages/ForgotPasswordPage'));
const ResetPasswordPage = lazyPage(() => import('@/features/auth/pages/ResetPasswordPage'));

function ErrorBoundary() {
  return (
    <div className="flex min-h-[40vh] flex-col items-center justify-center text-center">
      <h1 className="mb-2 text-2xl font-bold">页面加载失败</h1>
      <p style={{ color: 'var(--text-muted)' }}>请刷新页面后重试。</p>
    </div>
  );
}

const routes: RouteObject[] = [
  {
    path: '/',
    element: <RootLayout />,
    errorElement: <ErrorBoundary />,
    children: [
      { index: true, element: <HomePage /> },
      { path: 'problems', element: <ProblemListPage /> },
      { path: 'problems/:slug', element: <ProblemDetailPage /> },
      { path: 'submissions', element: protectedPage(<SubmissionsPage />) },
      { path: 'submissions/:id', element: protectedPage(<SubmissionDetailPage />) },
      { path: 'profile', element: protectedPage(<ProfilePage />) },
      { path: 'login', element: <LoginPage /> },
      { path: 'register', element: <RegisterPage /> },
      { path: 'verify-email', element: <VerifyEmailPage /> },
      { path: 'forgot-password', element: <ForgotPasswordPage /> },
      { path: 'reset-password', element: <ResetPasswordPage /> },
    ],
  },
];

export const router = createBrowserRouter(routes);
