import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useSearchParams, useNavigate, Link } from 'react-router';
import { resetPassword } from '@/features/auth/lib/auth-api';
import { ApiError } from '@/lib/api-client';
import { KeyRound, Lock, CheckCircle } from 'lucide-react';

const resetSchema = z
  .object({
    password: z.string().min(8, '密码至少 8 个字符'),
    confirmPassword: z.string().min(1, '请确认密码'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: '两次密码输入不一致',
    path: ['confirmPassword'],
  });

type ResetFormData = z.infer<typeof resetSchema>;

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get('token') || '';

  const [success, setSuccess] = useState(false);
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<ResetFormData>({
    resolver: zodResolver(resetSchema),
  });

  const onSubmit = async (data: ResetFormData) => {
    if (!token) {
      setError('root.serverError', { message: '无效的重置链接' });
      return;
    }
    try {
      await resetPassword(token, data.password);
      setSuccess(true);
      setTimeout(() => navigate('/login', { state: { message: '密码重置成功，请重新登录' } }), 2000);
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === 'TOKEN_EXPIRED' || err.code === 'INVALID_TOKEN') {
          setError('root.serverError', { message: '重置链接已失效，请重新申请' });
        } else {
          setError('root.serverError', { message: err.message });
        }
      } else {
        setError('root.serverError', { message: '重置失败，请稍后重试' });
      }
    }
  };

  if (!token) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div
          className="w-full max-w-md p-8 rounded-lg shadow-sm text-center"
          style={{ backgroundColor: 'var(--surface)' }}
        >
          <KeyRound size={48} className="mx-auto mb-4" style={{ color: 'var(--danger)' }} />
          <h1 className="text-2xl font-bold mb-2">链接无效</h1>
          <p className="text-sm mb-6" style={{ color: 'var(--text-muted)' }}>
            该密码重置链接无效或缺失
          </p>
          <Link
            to="/forgot-password"
            className="inline-block h-10 leading-10 px-6 rounded text-white text-sm font-medium no-underline"
            style={{ backgroundColor: 'var(--accent)' }}
          >
            重新申请重置
          </Link>
        </div>
      </div>
    );
  }

  if (success) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div
          className="w-full max-w-md p-8 rounded-lg shadow-sm text-center"
          style={{ backgroundColor: 'var(--surface)' }}
        >
          <CheckCircle size={48} className="mx-auto mb-4" style={{ color: 'var(--success)' }} />
          <h1 className="text-2xl font-bold mb-2">密码重置成功</h1>
          <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
            即将跳转到登录页面...
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <div
        className="w-full max-w-md p-8 rounded-lg shadow-sm"
        style={{ backgroundColor: 'var(--surface)' }}
      >
        <div className="flex items-center justify-center gap-2 mb-6">
          <KeyRound size={24} style={{ color: 'var(--accent)' }} />
          <h1 className="text-2xl font-bold">重置密码</h1>
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
            <label className="block text-sm font-medium mb-1.5">新密码</label>
            <p className="text-xs mb-1.5" style={{ color: 'var(--text-muted)' }}>
              至少 8 个字符
            </p>
            <div className="relative">
              <Lock size={16} className="absolute left-3 top-1/2 -translate-y-1/2" style={{ color: 'var(--text-muted)' }} />
              <input
                type="password"
                {...register('password')}
                placeholder="设置新密码"
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
              <p className="mt-1 text-xs" style={{ color: 'var(--danger)' }}>{errors.password.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-1.5">确认密码</label>
            <div className="relative">
              <Lock size={16} className="absolute left-3 top-1/2 -translate-y-1/2" style={{ color: 'var(--text-muted)' }} />
              <input
                type="password"
                {...register('confirmPassword')}
                placeholder="再次输入新密码"
                className="w-full h-10 pl-9 pr-3 rounded text-sm border outline-none transition-colors"
                style={{
                  borderColor: errors.confirmPassword ? 'var(--danger)' : 'var(--border)',
                  backgroundColor: 'var(--surface)',
                }}
                onFocus={(e) => !errors.confirmPassword && (e.currentTarget.style.borderColor = 'var(--accent)')}
                onBlur={(e) => (e.currentTarget.style.borderColor = errors.confirmPassword ? 'var(--danger)' : 'var(--border)')}
              />
            </div>
            {errors.confirmPassword && (
              <p className="mt-1 text-xs" style={{ color: 'var(--danger)' }}>{errors.confirmPassword.message}</p>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full h-10 rounded text-white text-sm font-medium transition-colors disabled:opacity-50"
            style={{ backgroundColor: 'var(--accent)' }}
            onMouseEnter={(e) => !isSubmitting && (e.currentTarget.style.backgroundColor = 'var(--accent-hover)')}
            onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'var(--accent)')}
          >
            {isSubmitting ? '重置中...' : '重置密码'}
          </button>
        </form>
      </div>
    </div>
  );
}
