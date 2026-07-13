import { useState, useEffect, useCallback } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Link } from 'react-router';
import { useRegister } from '@/features/auth/hooks/use-auth';
import { resendVerification } from '@/features/auth/lib/auth-api';
import { ApiError } from '@/lib/api-client';
import { UserPlus, Mail, Lock, User, CheckCircle } from 'lucide-react';

const registerSchema = z
  .object({
    email: z.string().min(1, '请输入邮箱').email('邮箱格式不正确'),
    username: z
      .string()
      .min(2, '用户名至少 2 个字符')
      .max(20, '用户名最多 20 个字符')
      .regex(/^[a-zA-Z0-9_-]+$/, '用户名只能包含字母、数字、下划线和连字符'),
    password: z.string().min(8, '密码至少 8 个字符'),
    confirmPassword: z.string().min(1, '请确认密码'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: '两次密码输入不一致',
    path: ['confirmPassword'],
  });

type RegisterFormData = z.infer<typeof registerSchema>;

export default function RegisterPage() {
  const [successEmail, setSuccessEmail] = useState<string | null>(null);
  const [resendCooldown, setResendCooldown] = useState(0);
  const [resendMessage, setResendMessage] = useState<string | null>(null);

  const registerMutation = useRegister();

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  });

  // Cooldown timer
  useEffect(() => {
    if (resendCooldown <= 0) return;
    const timer = setTimeout(() => setResendCooldown((c) => c - 1), 1000);
    return () => clearTimeout(timer);
  }, [resendCooldown]);

  const onSubmit = async (data: RegisterFormData) => {
    try {
      await registerMutation.mutateAsync({
        email: data.email,
        username: data.username,
        password: data.password,
      });
      setSuccessEmail(data.email);
      setResendCooldown(60);
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === 'EMAIL_ALREADY_EXISTS') {
          setError('email', { message: '该邮箱已被注册' });
        } else if (err.code === 'USERNAME_ALREADY_EXISTS') {
          setError('username', { message: '该用户名已被使用' });
        } else {
          setError('root.serverError', { message: err.message });
        }
      } else {
        setError('root.serverError', { message: '注册失败，请稍后重试' });
      }
    }
  };

  const handleResend = useCallback(async () => {
    if (!successEmail || resendCooldown > 0) return;
    try {
      await resendVerification(successEmail);
      setResendCooldown(60);
      setResendMessage('验证邮件已重新发送');
    } catch {
      setResendMessage('发送失败，请稍后重试');
    }
  }, [successEmail, resendCooldown]);

  // Mask email
  const maskEmail = (email: string) => {
    const [local, domain] = email.split('@');
    if (!local || !domain) return email;
    const visible = local.length <= 2 ? local : local.slice(0, 2) + '***';
    return `${visible}@${domain}`;
  };

  // Success state: show "check email"
  if (successEmail) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div
          className="w-full max-w-md p-8 rounded-lg shadow-sm text-center"
          style={{ backgroundColor: 'var(--surface)' }}
        >
          <CheckCircle size={48} className="mx-auto mb-4" style={{ color: 'var(--success)' }} />
          <h1 className="text-2xl font-bold mb-2">注册成功</h1>
          <p className="text-sm mb-1" style={{ color: 'var(--text-muted)' }}>
            我们已向 <span className="font-medium" style={{ color: 'var(--text)' }}>{maskEmail(successEmail)}</span> 发送了验证邮件
          </p>
          <p className="text-sm mb-6" style={{ color: 'var(--text-muted)' }}>
            请查收邮件并点击验证链接完成注册
          </p>

          <button
            onClick={handleResend}
            disabled={resendCooldown > 0}
            className="mb-4 text-sm no-underline disabled:opacity-50 border-none bg-transparent cursor-pointer"
            style={{ color: 'var(--accent)' }}
          >
            {resendCooldown > 0 ? `重新发送 (${resendCooldown}s)` : '重新发送验证邮件'}
          </button>

          {resendMessage && (
            <p className="text-xs mb-4" style={{ color: 'var(--text-muted)' }}>
              {resendMessage}
            </p>
          )}

          <Link
            to="/login"
            className="block w-full h-10 leading-10 rounded text-white text-sm font-medium text-center no-underline transition-colors"
            style={{ backgroundColor: 'var(--accent)' }}
          >
            前往登录
          </Link>
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
          <UserPlus size={24} style={{ color: 'var(--accent)' }} />
          <h1 className="text-2xl font-bold">注册</h1>
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
            <label className="block text-sm font-medium mb-1.5">邮箱</label>
            <div className="relative">
              <Mail size={16} className="absolute left-3 top-1/2 -translate-y-1/2" style={{ color: 'var(--text-muted)' }} />
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
              <p className="mt-1 text-xs" style={{ color: 'var(--danger)' }}>{errors.email.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-1.5">用户名</label>
            <div className="relative">
              <User size={16} className="absolute left-3 top-1/2 -translate-y-1/2" style={{ color: 'var(--text-muted)' }} />
              <input
                type="text"
                {...register('username')}
                placeholder="username"
                className="w-full h-10 pl-9 pr-3 rounded text-sm border outline-none transition-colors"
                style={{
                  borderColor: errors.username ? 'var(--danger)' : 'var(--border)',
                  backgroundColor: 'var(--surface)',
                }}
                onFocus={(e) => !errors.username && (e.currentTarget.style.borderColor = 'var(--accent)')}
                onBlur={(e) => (e.currentTarget.style.borderColor = errors.username ? 'var(--danger)' : 'var(--border)')}
              />
            </div>
            {errors.username && (
              <p className="mt-1 text-xs" style={{ color: 'var(--danger)' }}>{errors.username.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-1.5">密码</label>
            <p className="text-xs mb-1.5" style={{ color: 'var(--text-muted)' }}>
              至少 8 个字符
            </p>
            <div className="relative">
              <Lock size={16} className="absolute left-3 top-1/2 -translate-y-1/2" style={{ color: 'var(--text-muted)' }} />
              <input
                type="password"
                {...register('password')}
                placeholder="设置密码"
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
                placeholder="再次输入密码"
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
            {isSubmitting ? '注册中...' : '注册'}
          </button>
        </form>

        <p className="mt-6 text-center text-sm" style={{ color: 'var(--text-muted)' }}>
          已有账号？{' '}
          <Link to="/login" className="no-underline font-medium" style={{ color: 'var(--accent)' }}>
            登录
          </Link>
        </p>
      </div>
    </div>
  );
}
