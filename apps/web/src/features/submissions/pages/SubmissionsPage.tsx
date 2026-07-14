import { ChevronLeft, ChevronRight, RotateCcw } from 'lucide-react';
import { Link, useSearchParams } from 'react-router';

import { useAuth } from '@/features/auth/contexts/auth-context';
import { useLanguages } from '@/features/problems/hooks/use-problems';
import { useSubmissions } from '@/features/submissions/hooks/use-submissions';
import type { SubmissionListParams, SubmissionStatus } from '@/features/submissions/lib/submissions-api';

const PAGE_SIZE = 20;

const statusLabels: Record<SubmissionStatus, string> = {
  QUEUED: '等待判题',
  COMPILING: '正在编译',
  RUNNING: '正在运行',
  AC: '答案正确',
  WA: '答案错误',
  TLE: '时间超限',
  MLE: '内存超限',
  RE: '运行错误',
  CE: '编译错误',
  SYSTEM_ERROR: '系统错误',
};

export default function SubmissionsPage() {
  const { user } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const page = positiveInteger(searchParams.get('page'));
  const params: SubmissionListParams = {
    problem: searchParams.get('problem') || undefined,
    status: (searchParams.get('status') as SubmissionStatus | null) || undefined,
    language: searchParams.get('language') || undefined,
    page,
    page_size: PAGE_SIZE,
  };
  const submissions = useSubmissions(params, user?.id);
  const languages = useLanguages();

  const updateFilter = (key: string, value: string) => {
    setSearchParams((current) => {
      const next = new URLSearchParams(current);
      if (value) next.set(key, value);
      else next.delete(key);
      next.delete('page');
      return next;
    }, { replace: true });
  };

  if (submissions.status === 'pending') return <PageMessage>正在加载提交记录…</PageMessage>;
  if (submissions.status === 'error') {
    return (
      <PageMessage title="提交记录加载失败">
        <button className="secondary-button mt-4" onClick={() => void submissions.refetch()}>重新加载</button>
      </PageMessage>
    );
  }

  const totalPages = Math.max(1, Math.ceil(submissions.data.total / PAGE_SIZE));

  return (
    <section className="mx-auto max-w-[1200px]">
      <header className="mb-6">
        <p className="eyebrow">判题记录</p>
        <h1 className="mt-1 text-3xl font-bold">我的提交</h1>
        <p className="mt-2 text-sm text-[var(--text-muted)]">共 {submissions.data.total} 次正式提交</p>
      </header>

      <div className="mb-4 grid gap-3 border-y border-[var(--border)] bg-[var(--surface)] p-4 md:grid-cols-[minmax(220px,1fr)_180px_180px_auto]">
        <label>
          <span className="sr-only">按题目 Slug 筛选</span>
          <input
            className="field-control"
            aria-label="按题目 Slug 筛选"
            placeholder="题目 Slug"
            value={params.problem ?? ''}
            onChange={(event) => updateFilter('problem', event.target.value)}
          />
        </label>
        <FilterSelect
          label="状态筛选"
          value={params.status ?? ''}
          onChange={(value) => updateFilter('status', value)}
          options={Object.entries(statusLabels)}
        />
        <FilterSelect
          label="语言筛选"
          value={params.language ?? ''}
          onChange={(value) => updateFilter('language', value)}
          options={(languages.data ?? []).map((language) => [language.key, language.display_name])}
        />
        <button className="secondary-button gap-2" onClick={() => setSearchParams({}, { replace: true })}>
          <RotateCcw aria-hidden="true" size={15} />重置筛选
        </button>
      </div>

      {submissions.data.items.length === 0 ? (
        <PageMessage title="暂无提交记录">完成一次正式提交后，记录会显示在这里。</PageMessage>
      ) : (
        <>
          <div className="overflow-x-auto border-y border-[var(--border)] bg-[var(--surface)]">
            <table className="w-full min-w-[820px] border-collapse text-left">
              <thead className="bg-[var(--surface-subtle)] text-xs text-[var(--text-muted)]">
                <tr>
                  <th className="px-4 py-3 font-medium">提交时间</th>
                  <th className="px-4 py-3 font-medium">题目</th>
                  <th className="px-4 py-3 font-medium">语言</th>
                  <th className="px-4 py-3 font-medium">状态</th>
                  <th className="px-4 py-3 font-medium">耗时</th>
                  <th className="px-4 py-3 font-medium">内存</th>
                </tr>
              </thead>
              <tbody>
                {submissions.data.items.map((submission) => (
                  <tr className="border-t border-[var(--border)] hover:bg-[var(--surface-subtle)]" key={submission.id}>
                    <td className="px-4 py-4 text-sm text-[var(--text-muted)]">{formatDate(submission.created_at)}</td>
                    <td className="px-4 py-4">
                      <Link className="font-medium no-underline hover:text-[var(--accent)]" to={`/submissions/${submission.id}`}>
                        {submission.problem_title || submission.problem_slug}
                      </Link>
                    </td>
                    <td className="px-4 py-4 text-sm">{languageName(languages.data, submission.language_key)}</td>
                    <td className="px-4 py-4"><SubmissionBadge status={submission.status} /></td>
                    <td className="px-4 py-4 text-sm">{submission.time_ms == null ? '—' : `${submission.time_ms} ms`}</td>
                    <td className="px-4 py-4 text-sm">{submission.memory_kb == null ? '—' : formatMemory(submission.memory_kb)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <Pagination page={page} totalPages={totalPages} onPageChange={(nextPage) => {
            setSearchParams((current) => {
              const next = new URLSearchParams(current);
              if (nextPage === 1) next.delete('page');
              else next.set('page', String(nextPage));
              return next;
            });
          }} />
        </>
      )}
    </section>
  );
}

export function SubmissionBadge({ status }: { status: SubmissionStatus }) {
  const color = status === 'AC' ? 'var(--success)' : ['QUEUED', 'COMPILING', 'RUNNING'].includes(status) ? 'var(--info)' : 'var(--danger)';
  return <span className="text-sm font-semibold" style={{ color }}>{statusLabels[status]}</span>;
}

function FilterSelect({ label, value, options, onChange }: {
  label: string;
  value: string;
  options: Array<[string, string]>;
  onChange: (value: string) => void;
}) {
  return (
    <label>
      <span className="sr-only">{label}</span>
      <select className="field-control" aria-label={label} value={value} onChange={(event) => onChange(event.target.value)}>
        <option value="">全部{label.replace('筛选', '')}</option>
        {options.map(([optionValue, text]) => <option key={optionValue} value={optionValue}>{text}</option>)}
      </select>
    </label>
  );
}

function Pagination({ page, totalPages, onPageChange }: { page: number; totalPages: number; onPageChange: (page: number) => void }) {
  return (
    <nav className="mt-5 flex items-center justify-end gap-3" aria-label="提交记录分页">
      <button className="secondary-button gap-1 disabled:cursor-not-allowed disabled:opacity-50" disabled={page <= 1} onClick={() => onPageChange(page - 1)}>
        <ChevronLeft aria-hidden="true" size={15} />上一页
      </button>
      <span className="text-sm text-[var(--text-muted)]">第 {page} / {totalPages} 页</span>
      <button className="secondary-button gap-1 disabled:cursor-not-allowed disabled:opacity-50" disabled={page >= totalPages} onClick={() => onPageChange(page + 1)}>
        下一页<ChevronRight aria-hidden="true" size={15} />
      </button>
    </nav>
  );
}

function PageMessage({ title, children }: { title?: string; children: React.ReactNode }) {
  return (
    <div className="flex min-h-[320px] flex-col items-center justify-center border-y border-[var(--border)] bg-[var(--surface)] px-6 text-center">
      {title && <h2 className="text-xl font-semibold">{title}</h2>}
      <div className="mt-2 text-sm text-[var(--text-muted)]">{children}</div>
    </div>
  );
}

function positiveInteger(value: string | null) {
  const parsed = Number(value);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : 1;
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
}

function formatMemory(memoryKb: number) {
  return `${(memoryKb / 1024).toFixed(memoryKb % 1024 === 0 ? 0 : 1)} MB`;
}

function languageName(languages: Array<{ key: string; display_name: string }> | undefined, key: string) {
  return languages?.find((language) => language.key === key)?.display_name ?? key;
}
