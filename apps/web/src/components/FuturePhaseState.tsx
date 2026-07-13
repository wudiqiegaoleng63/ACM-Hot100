import { Construction } from 'lucide-react';

export default function FuturePhaseState({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <section className="mx-auto flex min-h-[480px] max-w-[920px] items-center justify-center bg-[var(--surface)] px-6 text-center">
      <div className="max-w-lg">
        <Construction aria-hidden="true" className="mx-auto text-[var(--accent)]" size={32} strokeWidth={1.75} />
        <p className="eyebrow mt-5">后续阶段</p>
        <h1 className="mt-2 text-2xl font-bold">{title}</h1>
        <p className="mt-3 text-sm leading-7 text-[var(--text-muted)]">{description}</p>
      </div>
    </section>
  );
}
