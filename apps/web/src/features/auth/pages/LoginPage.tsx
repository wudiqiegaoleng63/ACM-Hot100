import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Link, useNavigate, useLocation } from 'react-router';
import { useAuth } from '@/features/auth/contexts/auth-context';
import { AUTH_ERROR_CODES } from '@/features/auth/lib/auth-api';
import { ApiError } from '@/lib/api-client';
import { LogIn, Mail, Lock } from 'lucide-react';

const loginSchema = z.object({
  email: z.string().min(1, '请输入邮箱').email('邮箱格式不正确'),
  password: z.string().min(1, '请输入密码'),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login } = useAuth();

  const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/problems';

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginFormData) => {
    try {
      await login(data);
      navigate(from, { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === AUTH_ERROR_CODES.invalidCredentials) {
          setError('root.serverError', { message: '邮箱或密码不正确' });
        } else if (err.code === AUTH_ERROR_CODES.emailNotVerified) {
          setError('root.serverError', { message: '邮箱未验证，请先查收验证邮件' });
        } else {
          setError('root.serverError', { message: err.message });
        }
      } else {
        setError('root.serverError', { message: '登录失败，请稍后重试' });
      }
    }
  };

  return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <div
        className="w-full max-w-md p-8 rounded-lg shadow-sm"
        style={{ backgroundColor: 'var(--surface)' }}
      >
        <div className="flex items-center justify-center gap-2 mb-6">
          <LogIn size={24} style={{ color: 'var(--accent)' }} />
          <h1 className="text-2xl font-bold">登录</h1>
        </div>

        {errors.root?.serverError && (
          <div
            className="mb-4 p-3 rounded text-sm"
            style={{ backgroundColor: 'var(--danger-soft)', color: 'var(--danger)' }}
          >
            {errors.root.serverError.message}
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1.5" style={{ color: 'var(--text)' }}>
              邮箱
            </label>
            <div className="relative">
              <Mail
                size={16}
                className="absolute left-3 top-1/2 -translate-y-1/2"
                style={{ color: 'var(--text-muted)' }}
              />
              <input
                type="email"
                {...register('email')}
                placeholder="your@email.com"
                className="w-full h-10 pl-9 pr-3 rounded text-sm border outline-none transition-colors"
                style={{
                  borderColor: errors.email ? 'var(--danger)' : 'var(--border)',
                  backgroundColor: 'var(--surface)',
                }}
                onFocus={(e) => !errors.email && (e.currentTarget.style.borderColor = 'var(--accent)')}
                onBlur={(e) => (e.currentTarget.style.borderColor = errors.email ? 'var(--danger)' : 'var(--border)')}
              />
            </div>
            {errors.email && (
              <p className="mt-1 text-xs" style={{ color: 'var(--danger)' }}>
                {errors.email.message}
              </p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-1.5" style={{ color: 'var(--text)' }}>
              密码
            </label>
            <div className="relative">
              <Lock
                size={16}
                className="absolute left-3 top-1/2 -translate-y-1/2"
                style={{ color: 'var(--text-muted)' }}
              />
              <input
                type="password"
                {...register('password')}
                placeholder="输入密码"
                className="w-full h-10 pl-9 pr-3 rounded text-sm border outline-none transition-colors"
                style={{
                  borderColor: errors.password ? 'var(--danger)' : 'var(--border)',
                  backgroundColor: 'var(--surface)',
                }}
                onFocus={(e) => !errors.password && (e.currentTarget.style.borderColor = 'var(--accent)')}
                onBlur={(e) => (e.currentTarget.style.borderColor = errors.password ? 'var(--danger)' : 'var(--border)')}
              />
            </div>
            {errors.password && (
              <p className="mt-1 text-xs" style={{ color: 'var(--danger)' }}>
                {errors.password.message}
              </p>
            )}
          </div>

          <div className="flex justify-end">
            <Link
              to="/forgot-password"
              className="text-xs no-underline"
              style={{ color: 'var(--accent)' }}
            >
              忘记密码？
            </Link>
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full h-10 rounded text-white text-sm font-medium transition-colors disabled:opacity-50"
            style={{ backgroundColor: 'var(--accent)' }}
            onMouseEnter={(e) => !isSubmitting && (e.currentTarget.style.backgroundColor = 'var(--accent-hover)')}
            onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'var(--accent)')}
          >
            {isSubmitting ? '登录中...' : '登录'}
          </button>
        </form>

        <p className="mt-6 text-center text-sm" style={{ color: 'var(--text-muted)' }}>
          还没有账号？{' '}
          <Link to="/register" className="no-underline font-medium" style={{ color: 'var(--accent)' }}>
            注册
          </Link>
        </p>
      </div>
    </div>
  );
}
