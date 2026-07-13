import { ArrowLeft, ArrowRight, Play, Send } from 'lucide-react';
import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from 'react';
import { Link, useLocation, useNavigate } from 'react-router';

import type { ProblemNavigation } from '@/features/problems/lib/problems-api';

const SPLIT_RATIO_KEY = 'problem-workspace:split-ratio';
const DEFAULT_SPLIT_RATIO = 46;
const MIN_SPLIT_RATIO = 32;
const MAX_SPLIT_RATIO = 68;
const MOBILE_QUERY = '(max-width: 767px)';

type MobileTab = 'statement' | 'code' | 'result';
type ResultTab = 'input' | 'output' | 'result';
type NavigationItem = NonNullable<ProblemNavigation['prev']>;

interface ProblemWorkspaceProps {
  statement: React.ReactNode;
  editor: React.ReactNode;
  previous: NavigationItem | null;
  next: NavigationItem | null;
  isAuthenticated: boolean;
}

export default function ProblemWorkspace({
  statement,
  editor,
  previous,
  next,
  isAuthenticated,
}: ProblemWorkspaceProps) {
  const isMobile = useMediaQuery(MOBILE_QUERY);
  const [mobileTab, setMobileTab] = useState<MobileTab>('statement');
  const [resultTab, setResultTab] = useState<ResultTab>('input');
  const [splitRatio, setSplitRatio] = useState(readSplitRatio);
  const workspaceRef = useRef<HTMLDivElement>(null);
  const dragging = useRef(false);
  const location = useLocation();
  const navigate = useNavigate();

  useEffect(() => {
    const stopDragging = () => {
      dragging.current = false;
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    };
    const resize = (event: MouseEvent) => {
      const workspace = workspaceRef.current;
      if (!dragging.current || !workspace) return;
      const bounds = workspace.getBoundingClientRect();
      const ratio = clamp(((event.clientX - bounds.left) / bounds.width) * 100);
      setSplitRatio(ratio);
      localStorage.setItem(SPLIT_RATIO_KEY, String(ratio));
    };

    window.addEventListener('mousemove', resize);
    window.addEventListener('mouseup', stopDragging);
    return () => {
      window.removeEventListener('mousemove', resize);
      window.removeEventListener('mouseup', stopDragging);
      stopDragging();
    };
  }, []);

  const beginDragging = () => {
    dragging.current = true;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
  };

  const requireLogin = () => {
    if (isAuthenticated) return;
    navigate('/login', { state: { from: location } });
  };

  if (isMobile) {
    return (
      <section className="min-h-[calc(100vh-8rem)] bg-[var(--surface)]">
        <div className="grid grid-cols-3 border-b border-[var(--border)]" role="tablist" aria-label="做题区域">
          <MobileTabButton active={mobileTab === 'statement'} onClick={() => setMobileTab('statement')}>题面</MobileTabButton>
          <MobileTabButton active={mobileTab === 'code'} onClick={() => setMobileTab('code')}>代码</MobileTabButton>
          <MobileTabButton active={mobileTab === 'result'} onClick={() => setMobileTab('result')}>结果</MobileTabButton>
        </div>
        <div className="min-h-[calc(100vh-11rem)] overflow-auto">
          {mobileTab === 'statement' && statement}
          {mobileTab === 'code' && (
            <div className="flex min-h-[calc(100vh-11rem)] flex-col">
              <ActionBar onRun={requireLogin} onSubmit={requireLogin} />
              <div className="min-h-[420px] flex-1">{editor}</div>
            </div>
          )}
          {mobileTab === 'result' && <ResultPanel activeTab={resultTab} onTabChange={setResultTab} />}
        </div>
      </section>
    );
  }

  return (
    <section className="mx-auto max-w-[1600px] bg-[var(--surface)]">
      <div className="flex items-center justify-between border-b border-[var(--border)] px-4 py-3">
        <ProblemNavigationLink direction="previous" item={previous} />
        <ProblemNavigationLink direction="next" item={next} />
      </div>
      <div
        ref={workspaceRef}
        data-testid="desktop-workspace"
        className="grid h-[calc(100vh-10rem)] min-h-[620px]"
        style={{ gridTemplateColumns: `${splitRatio}% 6px ${100 - splitRatio}%` }}
      >
        <div className="min-w-0 overflow-y-auto">{statement}</div>
        <button
          type="button"
          role="separator"
          aria-label="调整题面和代码宽度"
          aria-orientation="vertical"
          aria-valuemin={MIN_SPLIT_RATIO}
          aria-valuemax={MAX_SPLIT_RATIO}
          aria-valuenow={splitRatio}
          className="cursor-col-resize border-0 border-x border-[var(--editor-border)] bg-[var(--editor-panel)] hover:bg-[var(--accent)]"
          onMouseDown={beginDragging}
        />
        <div className="flex min-w-0 flex-col overflow-hidden bg-[var(--editor-bg)]">
          <ActionBar onRun={requireLogin} onSubmit={requireLogin} />
          <div className="min-h-0 flex-[3] overflow-y-auto">{editor}</div>
          <div className="min-h-[220px] flex-[2] overflow-y-auto border-t border-[var(--editor-border)]">
            <ResultPanel activeTab={resultTab} onTabChange={setResultTab} />
          </div>
        </div>
      </div>
    </section>
  );
}

function ActionBar({ onRun, onSubmit }: { onRun: () => void; onSubmit: () => void }) {
  return (
    <div className="flex items-center justify-end gap-2 border-b border-[var(--editor-border)] bg-[var(--editor-panel)] px-3 py-2">
      <button className="secondary-button gap-2" onClick={onRun}>
        <Play aria-hidden="true" size={15} />运行样例
      </button>
      <button className="primary-button gap-2" onClick={onSubmit}>
        <Send aria-hidden="true" size={15} />正式提交
      </button>
    </div>
  );
}

function ResultPanel({ activeTab, onTabChange }: { activeTab: ResultTab; onTabChange: (tab: ResultTab) => void }) {
  return (
    <section className="min-h-full bg-[var(--editor-bg)] text-neutral-300">
      <div className="flex border-b border-[var(--editor-border)]" role="tablist" aria-label="运行结果区域">
        <ResultTabButton active={activeTab === 'input'} onClick={() => onTabChange('input')}>自测输入</ResultTabButton>
        <ResultTabButton active={activeTab === 'output'} onClick={() => onTabChange('output')}>输出</ResultTabButton>
        <ResultTabButton active={activeTab === 'result'} onClick={() => onTabChange('result')}>结果</ResultTabButton>
      </div>
      <div className="p-5 text-sm text-neutral-400">
        {activeTab === 'input' && '自定义输入将在样例运行阶段开放。'}
        {activeTab === 'output' && '程序输出将在样例运行后显示。'}
        {activeTab === 'result' && '样例运行后在这里显示判题结果。'}
      </div>
    </section>
  );
}

function MobileTabButton({ active, onClick, children }: TabButtonProps) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      className={`border-0 border-b-2 px-3 py-3 text-sm font-semibold ${active ? 'border-[var(--accent)] text-[var(--accent)]' : 'border-transparent text-[var(--text-muted)]'}`}
      onClick={onClick}
    >
      {children}
    </button>
  );
}

function ResultTabButton({ active, onClick, children }: TabButtonProps) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      className={`border-0 border-b-2 px-4 py-3 text-xs font-semibold ${active ? 'border-[var(--accent)] text-white' : 'border-transparent text-neutral-400'}`}
      onClick={onClick}
    >
      {children}
    </button>
  );
}

interface TabButtonProps {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}

function ProblemNavigationLink({ direction, item }: { direction: 'previous' | 'next'; item: NavigationItem | null }) {
  if (!item) return <span />;
  const isPrevious = direction === 'previous';
  return (
    <Link className="inline-flex items-center gap-2 text-sm no-underline hover:text-[var(--accent)]" to={`/problems/${item.slug}`}>
      {isPrevious && <ArrowLeft aria-hidden="true" size={16} />}
      <span>{item.title}</span>
      {!isPrevious && <ArrowRight aria-hidden="true" size={16} />}
    </Link>
  );
}

function readSplitRatio() {
  const stored = Number(localStorage.getItem(SPLIT_RATIO_KEY));
  return Number.isFinite(stored) && stored > 0 ? clamp(stored) : DEFAULT_SPLIT_RATIO;
}

function clamp(ratio: number) {
  return Math.round(Math.min(MAX_SPLIT_RATIO, Math.max(MIN_SPLIT_RATIO, ratio)));
}

function useMediaQuery(query: string) {
  const subscribe = useCallback((notify: () => void) => {
    const mediaQuery = window.matchMedia(query);
    mediaQuery.addEventListener('change', notify);
    return () => mediaQuery.removeEventListener('change', notify);
  }, [query]);
  const getSnapshot = useCallback(() => window.matchMedia(query).matches, [query]);
  return useSyncExternalStore(subscribe, getSnapshot, () => false);
}
