import { RotateCcw, Search } from 'lucide-react';
import { Link, useSearchParams } from 'react-router';

import StatusBadge from '@/components/StatusBadge';
import { useProblems, useTags } from '@/features/problems/hooks/use-problems';
import type {
  Difficulty,
  ProblemListParams,
  ProgressState,
} from '@/features/problems/lib/problems-api';

const difficultyLabels: Record<Difficulty, string> = {
  EASY: '简单',
  MEDIUM: '中等',
  HARD: '困难',
};

export default function ProblemListPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const params: ProblemListParams = {
    q: searchParams.get('q') || undefined,
    difficulty: (searchParams.get('difficulty') as Difficulty | null) || undefined,
    tag: searchParams.get('tag') || undefined,
    state: (searchParams.get('state') as ProgressState | null) || undefined,
    page: 1,
    page_size: 100,
  };
  const problems = useProblems(params);
  const tags = useTags();

  const updateFilter = (key: string, value: string) => {
    setSearchParams((current) => {
      const next = new URLSearchParams(current);
      if (value) next.set(key, value);
      else next.delete(key);
      return next;
    }, { replace: true });
  };

  if (problems.status === 'pending') {
    return <PageMessage>正在加载题单…</PageMessage>;
  }

  if (problems.status === 'error') {
    return (
      <PageMessage title="题单加载失败">
        <button className="secondary-button mt-4" onClick={() => void problems.refetch()}>
          重新加载
        </button>
      </PageMessage>
    );
  }

  const solved = problems.data.items.filter((item) => item.progress_state === 'SOLVED').length;
  const hasProgress = problems.data.items.some((item) => item.progress_state !== null);

  return (
    <section className="mx-auto max-w-[1200px]">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="eyebrow">训练题单</p>
          <h1 className="mt-1 text-3xl font-bold">ACM Hot 100</h1>
          <p className="mt-2 text-sm" style={{ color: 'var(--text-muted)' }}>
            {hasProgress ? `已通过 ${solved} / ${problems.data.total}` : `当前发布 ${problems.data.total} 道原创 ACM 题目`}
          </p>
        </div>
        {hasProgress && (
          <div className="w-full max-w-xs" aria-label={`已通过 ${solved} / ${problems.data.total}`}>
            <div className="h-2 overflow-hidden rounded bg-[var(--surface-subtle)]">
              <div
                className="h-full bg-[var(--success)]"
                style={{ width: `${problems.data.total ? (solved / problems.data.total) * 100 : 0}%` }}
              />
            </div>
          </div>
        )}
      </div>

      <div className="mb-4 grid gap-3 border-y border-[var(--border)] bg-[var(--surface)] p-4 md:grid-cols-[minmax(220px,1fr)_160px_180px_160px_auto]">
        <label className="relative">
          <span className="sr-only">搜索题目</span>
          <Search aria-hidden="true" className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]" size={16} />
          <input
            aria-label="搜索题目"
            className="field-control pl-9"
            placeholder="搜索题目标题"
            value={params.q ?? ''}
            onChange={(event) => updateFilter('q', event.target.value)}
          />
        </label>
        <FilterSelect
          label="难度筛选"
          value={params.difficulty ?? ''}
          onChange={(value) => updateFilter('difficulty', value)}
          options={[
            ['EASY', '简单'],
            ['MEDIUM', '中等'],
            ['HARD', '困难'],
          ]}
        />
        <FilterSelect
          label="标签筛选"
          value={params.tag ?? ''}
          onChange={(value) => updateFilter('tag', value)}
          options={(tags.data ?? []).map((tag) => [tag.slug, tag.name])}
        />
        <FilterSelect
          label="状态筛选"
          value={params.state ?? ''}
          onChange={(value) => updateFilter('state', value)}
          options={[
            ['NOT_STARTED', '未开始'],
            ['ATTEMPTED', '尝试中'],
            ['SOLVED', '已通过'],
          ]}
        />
        <button
          className="secondary-button inline-flex items-center justify-center gap-2"
          onClick={() => setSearchParams({}, { replace: true })}
        >
          <RotateCcw aria-hidden="true" size={15} />
          重置筛选
        </button>
      </div>

      {problems.data.items.length === 0 ? (
        <PageMessage title="没有符合条件的题目">调整筛选条件后再试。</PageMessage>
      ) : (
        <div className="overflow-x-auto border-y border-[var(--border)] bg-[var(--surface)]">
          <table className="w-full min-w-[720px] border-collapse text-left">
            <thead className="bg-[var(--surface-subtle)] text-xs text-[var(--text-muted)]">
              <tr>
                <th className="px-4 py-3 font-medium">状态</th>
                <th className="px-4 py-3 font-medium">序号</th>
                <th className="px-4 py-3 font-medium">题目</th>
                <th className="px-4 py-3 font-medium">难度</th>
                <th className="px-4 py-3 font-medium">标签</th>
              </tr>
            </thead>
            <tbody>
              {problems.data.items.map((problem) => (
                <tr key={problem.id} className="border-t border-[var(--border)] hover:bg-[var(--surface-subtle)]">
                  <td className="px-4 py-4"><StatusBadge state={problem.progress_state} /></td>
                  <td className="px-4 py-4 text-sm text-[var(--text-muted)]">{problem.order_index}</td>
                  <td className="px-4 py-4">
                    <Link className="font-medium no-underline hover:text-[var(--accent)]" to={`/problems/${problem.slug}`}>
                      {problem.title}
                    </Link>
                  </td>
                  <td className="px-4 py-4 text-sm">{difficultyLabels[problem.difficulty]}</td>
                  <td className="px-4 py-4">
                    <div className="flex flex-wrap gap-2">
                      {problem.tags.slice(0, 2).map((tag) => <span className="tag" key={tag.slug}>{tag.name}</span>)}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

function FilterSelect({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: string;
  options: Array<[string, string]>;
  onChange: (value: string) => void;
}) {
  return (
    <label>
      <span className="sr-only">{label}</span>
      <select aria-label={label} className="field-control" value={value} onChange={(event) => onChange(event.target.value)}>
        <option value="">全部{label.replace('筛选', '')}</option>
        {options.map(([optionValue, text]) => <option key={optionValue} value={optionValue}>{text}</option>)}
      </select>
    </label>
  );
}

function PageMessage({ title, children }: { title?: string; children: React.ReactNode }) {
  return (
    <div className="flex min-h-[320px] flex-col items-center justify-center border-y border-[var(--border)] bg-[var(--surface)] px-6 text-center">
      {title && <h2 className="mb-2 text-xl font-semibold">{title}</h2>}
      <div className="text-sm text-[var(--text-muted)]">{children}</div>
    </div>
  );
}
