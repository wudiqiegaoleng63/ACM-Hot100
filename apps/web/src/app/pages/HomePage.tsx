import { ArrowRight, Braces, FileInput, Terminal } from 'lucide-react';
import { Link } from 'react-router';

import { useAuth } from '@/features/auth/contexts/auth-context';
import { useProblems } from '@/features/problems/hooks/use-problems';

export default function HomePage() {
  const { user, isLoading } = useAuth();
  const problems = useProblems({ page: 1, page_size: 100 });

  if (isLoading) {
    return <div className="flex min-h-[50vh] items-center justify-center text-sm text-[var(--text-muted)]">正在恢复训练状态…</div>;
  }

  if (user) {
    const items = problems.data?.items ?? [];
    const solved = items.filter((item) => item.progress_state === 'SOLVED').length;
    const nextProblem = items.find((item) => item.progress_state !== 'SOLVED') ?? items[0];
    return (
      <section className="mx-auto max-w-[1200px] py-8 lg:py-16">
        <p className="eyebrow">欢迎回来，{user.username}</p>
        <h1 className="mt-3 max-w-3xl text-4xl font-bold leading-tight lg:text-5xl">继续你的 ACM 训练</h1>
        {problems.status === 'pending' ? (
          <p className="mt-6 text-[var(--text-muted)]">正在读取训练进度…</p>
        ) : problems.status === 'error' ? (
          <p className="mt-6 text-[var(--danger)]">训练进度暂时无法加载，可以先打开题单。</p>
        ) : nextProblem ? (
          <div className="mt-10 border-y border-[var(--border)] bg-[var(--surface)] px-6 py-7">
            <p className="text-sm text-[var(--text-muted)]">已通过 {solved} / {problems.data.total}</p>
            <h2 className="mt-2 text-2xl font-semibold">继续训练：{nextProblem.title}</h2>
            <div className="mt-6 flex flex-wrap gap-3">
              <Link className="primary-button no-underline" to={`/problems/${nextProblem.slug}`}>继续上一题</Link>
              <Link className="secondary-button no-underline" to="/problems">打开题单</Link>
            </div>
          </div>
        ) : (
          <div className="mt-10 border-y border-[var(--border)] bg-[var(--surface)] px-6 py-8">
            <h2 className="text-xl font-semibold">题单暂时为空</h2>
            <p className="mt-2 text-sm text-[var(--text-muted)]">发布题目后，这里会显示真实训练进度。</p>
          </div>
        )}
      </section>
    );
  }

  return (
    <section className="mx-auto max-w-[1200px] py-8 lg:py-16">
      <div className="grid items-center gap-10 lg:grid-cols-[1fr_0.9fr]">
        <div>
          <p className="eyebrow">完整程序 · 标准输入输出</p>
          <h1 className="mt-4 max-w-2xl text-4xl font-bold leading-[1.12] tracking-tight sm:text-5xl lg:text-6xl">
            用 ACM 模式刷完 Hot 100
          </h1>
          <p className="mt-6 max-w-xl text-base leading-8 text-[var(--text-muted)]">
            阅读中文题面，编写完整程序，从标准输入读取数据，用真实判题反馈推进训练。
          </p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Link className="primary-button inline-flex items-center gap-2 no-underline" to="/register">
              开始训练 <ArrowRight aria-hidden="true" size={16} />
            </Link>
            <Link className="secondary-button no-underline" to="/problems">浏览题单</Link>
          </div>
        </div>
        <div className="overflow-hidden border border-[var(--editor-border)] bg-[var(--editor-bg)] text-white">
          <div className="border-b border-[var(--editor-border)] bg-[var(--editor-panel)] px-4 py-3 text-xs text-neutral-400">main.cpp</div>
          <pre className="overflow-x-auto p-5 font-mono text-sm leading-7 text-neutral-200"><code>{`#include <bits/stdc++.h>
using namespace std;

int main() {
    int n, target;
    cin >> n >> target;
    vector<int> a(n);
    for (int &x : a) cin >> x;
}`}</code></pre>
          <div className="grid grid-cols-2 border-t border-[var(--editor-border)] text-sm">
            <div className="border-r border-[var(--editor-border)] p-4"><span className="text-xs text-neutral-500">输入</span><pre className="mt-2 font-mono">4 9{`\n`}2 7 11 15</pre></div>
            <div className="p-4"><span className="text-xs text-neutral-500">输出</span><pre className="mt-2 font-mono">1 2</pre></div>
          </div>
        </div>
      </div>

      <div className="mt-20 grid border-y border-[var(--border)] bg-[var(--surface)] md:grid-cols-3">
        <Feature icon={<Braces size={20} />} title="完整程序" text="不填写函数签名，提交可直接编译运行的完整代码。" />
        <Feature icon={<FileInput size={20} />} title="真实输入输出" text="每道题明确输入格式、输出格式、约束与资源限制。" border />
        <Feature icon={<Terminal size={20} />} title="路线化训练" text="按题单顺序推进，只展示由真实提交产生的进度。" border />
      </div>
    </section>
  );
}

function Feature({ icon, title, text, border = false }: { icon: React.ReactNode; title: string; text: string; border?: boolean }) {
  return (
    <div className={`p-6 ${border ? 'border-t border-[var(--border)] md:border-l md:border-t-0' : ''}`}>
      <div className="text-[var(--accent)]">{icon}</div>
      <h2 className="mt-4 font-semibold">{title}</h2>
      <p className="mt-2 text-sm leading-6 text-[var(--text-muted)]">{text}</p>
    </div>
  );
}
