import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Link } from 'react-router';
import { forgotPassword } from '@/features/auth/lib/auth-api';
import { KeyRound, Mail } from 'lucide-react';

const forgotSchema = z.object({
  email: z.string().min(1, '请输入邮箱').email('邮箱格式不正确'),
});

type ForgotFormData = z.infer<typeof forgotSchema>;

export default function ForgotPasswordPage() {
  const [submitted, setSubmitted] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ForgotFormData>({
    resolver: zodResolver(forgotSchema),
  });

  const onSubmit = async (_data: ForgotFormData) => {
    // Always call the API, but show the same message regardless
    try {
      await forgotPassword(_data.email);
    } catch {
      // Swallow errors to prevent email enumeration
    }
    setSubmitted(true);
  };

  if (submitted) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div
          className="w-full max-w-md p-8 rounded-lg shadow-sm text-center"
          style={{ backgroundColor: 'var(--surface)' }}
        >
          <Mail size={48} className="mx-auto mb-4" style={{ color: 'var(--accent)' }} />
          <h1 className="text-2xl font-bold mb-2">重置邮件已发送</h1>
          <p className="text-sm mb-6" style={{ color: 'var(--text-muted)' }}>
            如果该邮箱已注册，我们已发送重置邮件
          </p>
          <Link
            to="/login"
            className="inline-block h-10 leading-10 px-6 rounded text-white text-sm font-medium no-underline"
            style={{ backgroundColor: 'var(--accent)' }}
          >
            返回登录
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
          <KeyRound size={24} style={{ color: 'var(--accent)' }} />
          <h1 className="text-2xl font-bold">忘记密码</h1>
        </div>

        <p className="text-sm mb-6" style={{ color: 'var(--text-muted)' }}>
          输入注册时使用的邮箱，我们将发送密码重置链接
        </p>

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

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full h-10 rounded text-white text-sm font-medium transition-colors disabled:opacity-50"
            style={{ backgroundColor: 'var(--accent)' }}
            onMouseEnter={(e) => !isSubmitting && (e.currentTarget.style.backgroundColor = 'var(--accent-hover)')}
            onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'var(--accent)')}
          >
            {isSubmitting ? '发送中...' : '发送重置邮件'}
          </button>
        </form>

        <p className="mt-6 text-center text-sm" style={{ color: 'var(--text-muted)' }}>
          记得密码？{' '}
          <Link to="/login" className="no-underline font-medium" style={{ color: 'var(--accent)' }}>
            登录
          </Link>
        </p>
      </div>
    </div>
  );
}
