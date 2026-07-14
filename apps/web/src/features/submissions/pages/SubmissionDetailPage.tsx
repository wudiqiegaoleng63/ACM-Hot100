import { ArrowLeft, Clipboard, Code2 } from 'lucide-react';
import { useState } from 'react';
import { Link, useParams } from 'react-router';

import { useAuth } from '@/features/auth/contexts/auth-context';
import CodeEditor from '@/features/editor/components/CodeEditor';
import { useLanguages } from '@/features/problems/hooks/use-problems';
import { useSubmission } from '@/features/submissions/hooks/use-submissions';
import { SubmissionBadge } from '@/features/submissions/pages/SubmissionsPage';

export default function SubmissionDetailPage() {
  const { id = '' } = useParams<{ id: string }>();
  const { user } = useAuth();
  const submission = useSubmission(id, user?.id);
  const languages = useLanguages();
  const [copyMessage, setCopyMessage] = useState('');

  if (submission.status === 'pending') return <PageMessage>正在加载提交详情…</PageMessage>;
  if (submission.status === 'error') {
    return (
      <PageMessage title="提交不存在或无权查看">
        <p>该提交不存在，或不属于当前账号。</p>
        <Link className="secondary-button mt-4 no-underline" to="/submissions">返回提交记录</Link>
      </PageMessage>
    );
  }

  const data = submission.data;
  const language = languages.data?.find((item) => item.key === data.language_key);
  const copyCode = async () => {
    try {
      const clipboard = navigator.clipboard;
      if (!clipboard) throw new Error('clipboard unavailable');
      await clipboard.writeText(data.source_code);
      setCopyMessage('代码已复制');
    } catch {
      setCopyMessage('复制失败，请手动选择代码');
    }
  };

  return (
    <article className="mx-auto max-w-[1100px] space-y-6">
      <header className="border-y border-[var(--border)] bg-[var(--surface)] p-6 sm:p-8">
        <Link className="inline-flex items-center gap-2 text-sm no-underline text-[var(--text-muted)] hover:text-[var(--accent)]" to="/submissions">
          <ArrowLeft aria-hidden="true" size={16} />返回提交记录
        </Link>
        <div className="mt-5 flex flex-wrap items-start justify-between gap-4">
          <div>
            <p className="eyebrow">提交详情</p>
            <h1 className="mt-1 text-2xl font-bold">{data.problem_title || data.problem_slug}</h1>
            <p className="mt-2 text-sm text-[var(--text-muted)]">{formatDate(data.created_at)} · {language?.display_name ?? data.language_key}</p>
          </div>
          <SubmissionBadge status={data.status} />
        </div>
        <dl className="mt-6 grid grid-cols-2 gap-4 border-t border-[var(--border)] pt-5 sm:grid-cols-4">
          <Metric label="通过测试" value={`${data.passed_cases}/${data.total_cases}`} />
          <Metric label="耗时" value={data.time_ms == null ? '—' : `${data.time_ms} ms`} />
          <Metric label="峰值内存" value={data.memory_kb == null ? '—' : `${(data.memory_kb / 1024).toFixed(1)} MB`} />
          <Metric label="判题完成" value={data.judged_at ? formatDate(data.judged_at) : '等待中'} />
        </dl>
      </header>

      {(data.compiler_output || data.error_message) && (
        <section className="border-y border-[var(--border)] bg-[var(--surface)] p-6">
          <h2 className="section-heading">判题信息</h2>
          {data.compiler_output && <OutputBlock label="编译输出" value={data.compiler_output} />}
          {data.error_message && <OutputBlock label="错误信息" value={data.error_message} />}
          {data.status === 'WA' && (
            <p className="mt-4 text-sm text-[var(--text-muted)]">隐藏测试的输入与输出不会公开。</p>
          )}
        </section>
      )}

      <section className="overflow-hidden border-y border-[var(--border)] bg-[var(--surface)]">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[var(--border)] px-6 py-4">
          <h2 className="inline-flex items-center gap-2 text-lg font-semibold"><Code2 aria-hidden="true" size={18} />提交代码</h2>
          <div className="flex items-center gap-3">
            {copyMessage && <span className="text-xs text-[var(--text-muted)]" role="status">{copyMessage}</span>}
            <button className="secondary-button gap-2" onClick={copyCode}><Clipboard aria-hidden="true" size={15} />复制代码</button>
          </div>
        </div>
        <div className="h-[440px] bg-[var(--editor-bg)]">
          <CodeEditor
            value={data.source_code}
            language={language?.editor_language ?? data.language_key}
            onChange={() => undefined}
            readOnly
          />
        </div>
      </section>

      <div className="flex justify-end">
        <Link
          className="primary-button gap-2 no-underline"
          to={`/problems/${data.problem_slug}`}
          state={{ submission: { languageKey: data.language_key, sourceCode: data.source_code } }}
        >
          回到题目继续修改
        </Link>
      </div>
    </article>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return <div><dt className="text-xs text-[var(--text-muted)]">{label}</dt><dd className="mt-1 text-sm font-semibold">{value}</dd></div>;
}

function OutputBlock({ label, value }: { label: string; value: string }) {
  return (
    <div className="mt-4">
      <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-[var(--text-muted)]">{label}</p>
      <pre className="max-h-[240px] overflow-auto whitespace-pre-wrap break-words border border-[var(--editor-border)] bg-[var(--editor-bg)] p-4 font-mono text-xs leading-6 text-white">{value}</pre>
    </div>
  );
}

function PageMessage({ title, children }: { title?: string; children: React.ReactNode }) {
  return (
    <div className="flex min-h-[420px] flex-col items-center justify-center border-y border-[var(--border)] bg-[var(--surface)] px-6 text-center">
      {title && <h1 className="text-2xl font-bold">{title}</h1>}
      <div className="mt-2 text-sm text-[var(--text-muted)]">{children}</div>
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
}
