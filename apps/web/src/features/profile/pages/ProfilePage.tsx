import { Circle, CircleCheck, Clock3 } from 'lucide-react';

import { useProgressByStage, useProgressSummary } from '@/features/profile/hooks/use-profile';

export default function ProfilePage() {
  const summary = useProgressSummary();
  const stages = useProgressByStage();

  if (summary.status === 'pending' || stages.status === 'pending') return <PageMessage>正在加载训练进度…</PageMessage>;
  if (summary.status === 'error' || stages.status === 'error') {
    return (
      <PageMessage title="训练进度加载失败">
        <button className="secondary-button mt-4" onClick={() => void Promise.all([summary.refetch(), stages.refetch()])}>重新加载</button>
      </PageMessage>
    );
  }

  const total = summary.data.total_problems;
  const percent = total === 0 ? 0 : Math.round((summary.data.solved / total) * 100);

  return (
    <section className="mx-auto max-w-[1100px] space-y-6">
      <header className="border-y border-[var(--border)] bg-[var(--surface)] p-6 sm:p-8">
        <p className="eyebrow">训练概览</p>
        <div className="mt-2 flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">个人进度</h1>
            <p className="mt-2 text-sm text-[var(--text-muted)]">所有数据均来自真实正式提交。</p>
          </div>
          <p className="text-2xl font-bold text-[var(--success)]">{summary.data.solved} / {total}</p>
        </div>
        <div className="mt-6" aria-label={`已通过 ${summary.data.solved} / ${total}`}>
          <div className="h-3 overflow-hidden rounded bg-[var(--surface-subtle)]">
            <div className="h-full bg-[var(--success)]" style={{ width: `${percent}%` }} />
          </div>
          <p className="mt-2 text-right text-xs text-[var(--text-muted)]">完成 {percent}%</p>
        </div>
      </header>

      <section className="grid gap-4 sm:grid-cols-3" aria-label="三态进度分布">
        <StateCard icon={<CircleCheck aria-hidden="true" />} label="已通过" value={summary.data.solved} color="var(--success)" />
        <StateCard icon={<Clock3 aria-hidden="true" />} label="尝试中" value={summary.data.attempted} color="var(--warning)" />
        <StateCard icon={<Circle aria-hidden="true" />} label="未开始" value={summary.data.not_started} color="var(--text-muted)" />
      </section>

      <section className="border-y border-[var(--border)] bg-[var(--surface)] p-6 sm:p-8">
        <h2 className="section-heading">训练阶段</h2>
        {stages.data.length === 0 ? (
          <p className="mt-4 text-sm text-[var(--text-muted)]">当前没有已发布题目。</p>
        ) : (
          <div className="mt-5 space-y-5">
            {stages.data.map((stage) => {
              const stagePercent = stage.total === 0 ? 0 : Math.round((stage.solved / stage.total) * 100);
              return (
                <div key={stage.stage}>
                  <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
                    <h3 className="font-semibold">{stage.stage}</h3>
                    <p className="text-xs text-[var(--text-muted)]">
                      已通过 {stage.solved} · 尝试中 {stage.attempted} · 未开始 {stage.not_started}
                    </p>
                  </div>
                  <div className="h-2 overflow-hidden rounded bg-[var(--surface-subtle)]" aria-label={`${stage.stage} 已通过 ${stage.solved} / ${stage.total}`}>
                    <div className="h-full bg-[var(--success)]" style={{ width: `${stagePercent}%` }} />
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>
    </section>
  );
}

function StateCard({ icon, label, value, color }: { icon: React.ReactNode; label: string; value: number; color: string }) {
  return (
    <div className="border-y border-[var(--border)] bg-[var(--surface)] p-5">
      <div className="flex items-center gap-2 text-sm" style={{ color }}>{icon}<span>{label}</span></div>
      <p className="mt-3 text-3xl font-bold">{value}</p>
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
