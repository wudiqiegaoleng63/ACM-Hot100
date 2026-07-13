import { Check, Circle, Clock3 } from 'lucide-react';

import type { ProgressState } from '@/features/problems/lib/problems-api';

const content = {
  NOT_STARTED: { label: '未开始', Icon: Circle, color: 'var(--text-muted)' },
  ATTEMPTED: { label: '尝试中', Icon: Clock3, color: 'var(--warning)' },
  SOLVED: { label: '已通过', Icon: Check, color: 'var(--success)' },
} as const;

export default function StatusBadge({ state }: { state: ProgressState | null }) {
  if (!state) return <span className="text-sm" style={{ color: 'var(--text-muted)' }}>—</span>;
  const { label, Icon, color } = content[state];
  return (
    <span className="inline-flex items-center gap-1 text-sm" style={{ color }}>
      <Icon aria-hidden="true" size={15} />
      {label}
    </span>
  );
}
