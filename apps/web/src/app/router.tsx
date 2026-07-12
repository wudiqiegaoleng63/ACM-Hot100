import { lazy, ComponentType, LazyExoticComponent } from 'react';
import {
  createBrowserRouter,
  type RouteObject,
} from 'react-router';
import RootLayout from '@/app/layouts/RootLayout';

function lazyPage<T extends ComponentType<unknown>>(
  importFn: () => Promise<{ default: T }>,
): LazyExoticComponent<T> {
  return lazy(importFn);
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
    <div className="flex flex-col items-center justify-center min-h-[40vh]">
      <h1 className="text-2xl font-bold mb-2">Something went wrong</h1>
      <p style={{ color: 'var(--text-muted)' }}>
        An unexpected error occurred. Please try again.
      </p>
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
      { path: 'submissions', element: <SubmissionsPage /> },
      { path: 'submissions/:id', element: <SubmissionDetailPage /> },
      { path: 'profile', element: <ProfilePage /> },
      { path: 'login', element: <LoginPage /> },
      { path: 'register', element: <RegisterPage /> },
      { path: 'verify-email', element: <VerifyEmailPage /> },
      { path: 'forgot-password', element: <ForgotPasswordPage /> },
      { path: 'reset-password', element: <ResetPasswordPage /> },
    ],
  },
];

export const router = createBrowserRouter(routes);
