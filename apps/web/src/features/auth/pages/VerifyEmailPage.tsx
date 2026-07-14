import { useState, useEffect, useCallback, useRef } from 'react';
import { useSearchParams, useNavigate, Link } from 'react-router';
import { verifyEmail, resendVerification } from '@/features/auth/lib/auth-api';
import { ApiError } from '@/lib/api-client';
import { MailCheck, MailX, Clock, RefreshCw } from 'lucide-react';

type VerifyState = 'verifying' | 'success' | 'expired' | 'already-used' | 'error';

export default function VerifyEmailPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get('token') || '';

  const [state, setState] = useState<VerifyState>(token ? 'verifying' : 'error');
  const [resendCooldown, setResendCooldown] = useState(0);
  const [resendEmail, setResendEmail] = useState('');
  const [resendMessage, setResendMessage] = useState<string | null>(null);
  const verificationRequest = useRef<ReturnType<typeof verifyEmail> | null>(null);

  useEffect(() => {
    if (!token) return;
    const request = verificationRequest.current ?? verifyEmail(token);
    verificationRequest.current = request;

    let cancelled = false;
    request.then(() => {
      if (!cancelled) setState('success');
    }).catch((err: unknown) => {
      if (cancelled) return;
      if (err instanceof ApiError) {
        if (err.code === 'TOKEN_EXPIRED') {
          setState('expired');
        } else if (err.code === 'TOKEN_ALREADY_USED') {
          setState('already-used');
        } else {
          setState('error');
        }
      } else {
        setState('error');
      }
    });

    return () => { cancelled = true; };
  }, [token]);

  useEffect(() => {
    if (resendCooldown <= 0) return;
    const timer = setTimeout(() => setResendCooldown((c) => c - 1), 1000);
    return () => clearTimeout(timer);
  }, [resendCooldown]);

  const handleResend = useCallback(async () => {
    if (!resendEmail || resendCooldown > 0) return;
    try {
      await resendVerification(resendEmail);
      setResendCooldown(60);
      setResendMessage('验证邮件已重新发送');
    } catch {
      setResendMessage('发送失败，请稍后重试');
    }
  }, [resendEmail, resendCooldown]);

  // Auto-redirect on success
  useEffect(() => {
    if (state !== 'success') return;
    const timer = setTimeout(() => navigate('/login'), 3000);
    return () => clearTimeout(timer);
  }, [state, navigate]);

  const renderContent = () => {
    switch (state) {
      case 'verifying':
        return (
          <>
            <div className="animate-spin w-12 h-12 border-4 rounded-full mx-auto mb-4" style={{ borderColor: 'var(--border)', borderTopColor: 'var(--accent)' }} />
            <h1 className="text-2xl font-bold mb-2">验证中</h1>
            <p className="text-sm" style={{ color: 'var(--text-muted)' }}>正在验证您的邮箱...</p>
          </>
        );

      case 'success':
        return (
          <>
            <MailCheck size={48} className="mx-auto mb-4" style={{ color: 'var(--success)' }} />
            <h1 className="text-2xl font-bold mb-2">邮箱验证成功</h1>
            <p className="text-sm mb-4" style={{ color: 'var(--text-muted)' }}>
              即将跳转到登录页面...
            </p>
            <Link
              to="/login"
              className="inline-block h-10 leading-10 px-6 rounded text-white text-sm font-medium no-underline"
              style={{ backgroundColor: 'var(--accent)' }}
            >
              前往登录
            </Link>
          </>
        );

      case 'expired':
        return (
          <>
            <Clock size={48} className="mx-auto mb-4" style={{ color: 'var(--warning)' }} />
            <h1 className="text-2xl font-bold mb-2">验证链接已过期</h1>
            <p className="text-sm mb-4" style={{ color: 'var(--text-muted)' }}>
              该验证链接已失效，请重新发送验证邮件
            </p>
            <div className="space-y-3">
              <input
                type="email"
                value={resendEmail}
                onChange={(e) => setResendEmail(e.target.value)}
                placeholder="输入注册邮箱"
                className="w-full h-10 px-3 rounded text-sm border outline-none"
                style={{ borderColor: 'var(--border)', backgroundColor: 'var(--surface)' }}
              />
              <button
                onClick={handleResend}
                disabled={resendCooldown > 0 || !resendEmail}
                className="w-full h-10 rounded text-white text-sm font-medium disabled:opacity-50 border-none cursor-pointer"
                style={{ backgroundColor: 'var(--accent)' }}
              >
                {resendCooldown > 0 ? `重新发送 (${resendCooldown}s)` : '重新发送验证邮件'}
              </button>
              {resendMessage && (
                <p className="text-xs" style={{ color: 'var(--text-muted)' }}>{resendMessage}</p>
              )}
            </div>
          </>
        );

      case 'already-used':
        return (
          <>
            <MailX size={48} className="mx-auto mb-4" style={{ color: 'var(--info)' }} />
            <h1 className="text-2xl font-bold mb-2">邮箱已验证</h1>
            <p className="text-sm mb-4" style={{ color: 'var(--text-muted)' }}>
              该邮箱已经验证过了，请直接登录
            </p>
            <Link
              to="/login"
              className="inline-block h-10 leading-10 px-6 rounded text-white text-sm font-medium no-underline"
              style={{ backgroundColor: 'var(--accent)' }}
            >
              前往登录
            </Link>
          </>
        );

      default:
        return (
          <>
            <RefreshCw size={48} className="mx-auto mb-4" style={{ color: 'var(--danger)' }} />
            <h1 className="text-2xl font-bold mb-2">验证失败</h1>
            <p className="text-sm mb-4" style={{ color: 'var(--text-muted)' }}>
              无法完成邮箱验证，请重新尝试
            </p>
            <Link
              to="/register"
              className="inline-block h-10 leading-10 px-6 rounded text-white text-sm font-medium no-underline"
              style={{ backgroundColor: 'var(--accent)' }}
            >
              返回注册
            </Link>
          </>
        );
    }
  };

  return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <div
        className="w-full max-w-md p-8 rounded-lg shadow-sm text-center"
        style={{ backgroundColor: 'var(--surface)' }}
      >
        {renderContent()}
      </div>
    </div>
  );
}
