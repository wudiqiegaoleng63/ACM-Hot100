import 'katex/dist/katex.min.css';

import { ArrowLeft, ArrowRight, Clock3, Database } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import { Link, useParams } from 'react-router';
import rehypeKatex from 'rehype-katex';
import remarkGfm from 'remark-gfm';
import remarkMath from 'remark-math';

import { useProblem, useProblemNavigation } from '@/features/problems/hooks/use-problems';
import type { Difficulty } from '@/features/problems/lib/problems-api';

const difficultyLabels: Record<Difficulty, string> = {
  EASY: '简单',
  MEDIUM: '中等',
  HARD: '困难',
};

export default function ProblemDetailPage() {
  const { slug = '' } = useParams<{ slug: string }>();
  const problem = useProblem(slug);
  const navigation = useProblemNavigation(slug);

  if (problem.status === 'pending') {
    return <PageMessage>正在加载题目…</PageMessage>;
  }

  if (problem.status === 'error') {
    return (
      <PageMessage title="题目加载失败">
        <button className="secondary-button mt-4" onClick={() => void problem.refetch()}>
          重新加载
        </button>
      </PageMessage>
    );
  }

  const data = problem.data;
  const memoryMb = Math.round(data.memory_limit_kb / 1024);

  return (
    <article className="mx-auto max-w-[980px] bg-[var(--surface)] px-5 py-6 sm:px-8 lg:px-12 lg:py-10">
      <header className="border-b border-[var(--border)] pb-6">
        <p className="eyebrow">第 {data.order_index} 题 · {data.stage}</p>
        <div className="mt-2 flex flex-wrap items-center gap-3">
          <h1 className="text-3xl font-bold tracking-tight">{data.title}</h1>
          <span className={`difficulty difficulty-${data.difficulty.toLowerCase()}`}>
            {difficultyLabels[data.difficulty]}
          </span>
        </div>
        <div className="mt-4 flex flex-wrap items-center gap-2">
          {data.tags.map((tag) => <span className="tag" key={tag.slug}>{tag.name}</span>)}
        </div>
        <div className="mt-5 flex flex-wrap gap-5 text-sm text-[var(--text-muted)]">
          <span className="inline-flex items-center gap-2"><Clock3 aria-hidden="true" size={16} />{data.time_limit_ms} ms</span>
          <span className="inline-flex items-center gap-2"><Database aria-hidden="true" size={16} />{memoryMb} MB</span>
        </div>
      </header>

      <div className="problem-markdown mt-8">
        <Markdown>{data.statement_md}</Markdown>
        <MarkdownSection title="输入格式" content={data.input_format_md} />
        <MarkdownSection title="输出格式" content={data.output_format_md} />
        <MarkdownSection title="数据范围" content={data.constraints_md} />
      </div>

      <section className="mt-10" aria-labelledby="samples-title">
        <h2 id="samples-title" className="section-heading">公开样例</h2>
        <div className="mt-4 space-y-6">
          {data.sample_cases.map((sample, index) => (
            <div className="border border-[var(--border)]" key={sample.id}>
              <div className="border-b border-[var(--border)] bg-[var(--surface-subtle)] px-4 py-3 text-sm font-semibold">
                样例 {index + 1}
              </div>
              <div className="grid md:grid-cols-2">
                <SampleBlock label="输入" value={sample.input_data} />
                <SampleBlock label="输出" value={sample.expected_output} border />
              </div>
              {sample.explanation_md && (
                <div className="border-t border-[var(--border)] px-4 py-4">
                  <p className="mb-2 text-sm font-semibold">解释</p>
                  <div className="problem-markdown text-sm"><Markdown>{sample.explanation_md}</Markdown></div>
                </div>
              )}
            </div>
          ))}
        </div>
      </section>

      {data.hints_md && (
        <section className="problem-markdown mt-10">
          <MarkdownSection title="提示" content={data.hints_md} />
        </section>
      )}

      <nav className="mt-12 flex items-center justify-between border-t border-[var(--border)] pt-6" aria-label="题目导航">
        {navigation.data?.prev ? (
          <Link className="secondary-button inline-flex items-center gap-2 no-underline" to={`/problems/${navigation.data.prev.slug}`}>
            <ArrowLeft aria-hidden="true" size={16} />
            <span className="hidden sm:inline">{navigation.data.prev.title}</span>
            <span className="sm:hidden">上一题</span>
          </Link>
        ) : <span />}
        {navigation.data?.next ? (
          <Link className="secondary-button inline-flex items-center gap-2 no-underline" to={`/problems/${navigation.data.next.slug}`}>
            <span className="hidden sm:inline">{navigation.data.next.title}</span>
            <span className="sm:hidden">下一题</span>
            <ArrowRight aria-hidden="true" size={16} />
          </Link>
        ) : <span />}
      </nav>
    </article>
  );
}

function Markdown({ children }: { children: string }) {
  return (
    <ReactMarkdown
      skipHtml
      remarkPlugins={[remarkGfm, remarkMath]}
      rehypePlugins={[rehypeKatex]}
    >
      {children}
    </ReactMarkdown>
  );
}

function MarkdownSection({ title, content }: { title: string; content: string }) {
  return (
    <section className="mt-8">
      <h2 className="section-heading">{title}</h2>
      <div className="mt-3"><Markdown>{content}</Markdown></div>
    </section>
  );
}

function SampleBlock({ label, value, border = false }: { label: string; value: string; border?: boolean }) {
  return (
    <div className={`min-w-0 p-4 ${border ? 'border-t border-[var(--border)] md:border-l md:border-t-0' : ''}`}>
      <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-[var(--text-muted)]">{label}</p>
      <pre className="overflow-x-auto whitespace-pre-wrap break-words font-mono text-sm leading-6">{value}</pre>
    </div>
  );
}

function PageMessage({ title, children }: { title?: string; children: React.ReactNode }) {
  return (
    <div className="flex min-h-[480px] flex-col items-center justify-center bg-[var(--surface)] px-6 text-center">
      {title && <h1 className="mb-2 text-2xl font-bold">{title}</h1>}
      <div className="text-sm text-[var(--text-muted)]">{children}</div>
    </div>
  );
}
