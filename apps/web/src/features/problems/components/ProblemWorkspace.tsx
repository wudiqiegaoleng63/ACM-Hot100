import { ArrowLeft, ArrowRight, Loader2, Play, RefreshCw, Send, Zap } from 'lucide-react';
import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from 'react';
import { Link, useLocation, useNavigate } from 'react-router';

import { useCreateSampleRun, useSampleRun, useHealth } from '@/features/problems/hooks/use-problems';
import type { ProblemNavigation, SampleCase, SampleRunStatus } from '@/features/problems/lib/problems-api';
import { isTerminalRunStatus } from '@/features/problems/lib/problems-api';

const SPLIT_RATIO_KEY = 'problem-workspace:split-ratio';
const DEFAULT_SPLIT_RATIO = 46;
const MIN_SPLIT_RATIO = 32;
const MAX_SPLIT_RATIO = 68;
const MOBILE_QUERY = '(max-width: 767px)';
const MAX_CUSTOM_INPUT_BYTES = 16 * 1024;

type MobileTab = 'statement' | 'code' | 'result';
type ResultTab = 'input' | 'output' | 'result';
type NavigationItem = NonNullable<ProblemNavigation['prev']>;
type InputMode = 'sample' | 'custom';

interface ProblemWorkspaceProps {
  statement: React.ReactNode;
  editor: React.ReactNode;
  previous: NavigationItem | null;
  next: NavigationItem | null;
  isAuthenticated: boolean;
  problemSlug: string;
  sampleCases: SampleCase[];
  languageKey: string;
  sourceCode: string;
}

export default function ProblemWorkspace({
  statement,
  editor,
  previous,
  next,
  isAuthenticated,
  problemSlug,
  sampleCases,
  languageKey,
  sourceCode,
}: ProblemWorkspaceProps) {
  const isMobile = useMediaQuery(MOBILE_QUERY);
  const [mobileTab, setMobileTab] = useState<MobileTab>('statement');
  const [resultTab, setResultTab] = useState<ResultTab>('input');
  const [splitRatio, setSplitRatio] = useState(readSplitRatio);
  const workspaceRef = useRef<HTMLDivElement>(null);
  const dragging = useRef(false);
  const location = useLocation();
  const navigate = useNavigate();

  // Sample run state
  const [inputMode, setInputMode] = useState<InputMode>('sample');
  const [selectedSampleID, setSelectedSampleID] = useState(sampleCases[0]?.id ?? '');
  const [customInput, setCustomInput] = useState('');
  const [customInputError, setCustomInputError] = useState('');
  const [activeRunID, setActiveRunID] = useState<string | null>(null);
  const [runError, setRunError] = useState('');

  const createRun = useCreateSampleRun();
  const sampleRun = useSampleRun(activeRunID);
  const health = useHealth();
  const isMockJudge = health.data?.judge_mode === 'mock';

  // Reset sample selection when sample cases change
  useEffect(() => {
    const firstID = sampleCases[0]?.id;
    if (firstID && !sampleCases.find((s) => s.id === selectedSampleID)) {
      setSelectedSampleID(firstID);
    }
  }, [sampleCases, selectedSampleID]);

  // Clear active run when problem changes
  useEffect(() => {
    setActiveRunID(null);
    setRunError('');
    setInputMode('sample');
    setSelectedSampleID(sampleCases[0]?.id ?? '');
    setCustomInput('');
    setCustomInputError('');
  }, [problemSlug, sampleCases]);

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

  const handleRun = async () => {
    if (!isAuthenticated) {
      navigate('/login', { state: { from: location } });
      return;
    }
    if (!languageKey || !sourceCode.trim()) return;

    setRunError('');
    setActiveRunID(null);

    if (inputMode === 'custom') {
      const byteLength = new TextEncoder().encode(customInput).byteLength;
      if (byteLength > MAX_CUSTOM_INPUT_BYTES) {
        setCustomInputError('自定义输入不能超过 16KB。');
        return;
      }
      setCustomInputError('');
    }

    try {
      const params = inputMode === 'sample'
        ? { language_key: languageKey, source_code: sourceCode, sample_case_id: selectedSampleID }
        : { language_key: languageKey, source_code: sourceCode, custom_input: customInput };
      const run = await createRun.mutateAsync({ slug: problemSlug, params });
      setActiveRunID(run.id);
      setResultTab('result');
      if (isMobile) setMobileTab('result');
    } catch (err) {
      setRunError(err instanceof Error ? err.message : '运行请求失败');
    }
  };

  const handleRetry = () => {
    setRunError('');
    setActiveRunID(null);
    void handleRun();
  };

  const runStatus = sampleRun.data?.status;
  const isRunning = activeRunID != null && runStatus != null && !isTerminalRunStatus(runStatus);

  const actionBar = (
    <ActionBar
      onRun={handleRun}
      onSubmit={requireLogin}
      isRunning={isRunning}
      isMockJudge={isMockJudge}
    />
  );

  const resultPanel = (
    <ResultPanel
      activeTab={resultTab}
      onTabChange={setResultTab}
      inputMode={inputMode}
      onInputModeChange={setInputMode}
      sampleCases={sampleCases}
      selectedSampleID={selectedSampleID}
      onSampleSelect={setSelectedSampleID}
      customInput={customInput}
      onCustomInputChange={(value) => {
        setCustomInput(value);
        if (new TextEncoder().encode(value).byteLength > MAX_CUSTOM_INPUT_BYTES) {
          setCustomInputError('自定义输入不能超过 16KB。');
        } else {
          setCustomInputError('');
        }
      }}
      customInputError={customInputError}
      run={sampleRun.data}
      runError={runError}
      onRetry={handleRetry}
      isMockJudge={isMockJudge}
    />
  );

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
              {actionBar}
              <div className="min-h-[420px] flex-1">{editor}</div>
            </div>
          )}
          {mobileTab === 'result' && resultPanel}
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
          {actionBar}
          <div className="min-h-0 flex-[3] overflow-y-auto">{editor}</div>
          <div className="min-h-[220px] flex-[2] overflow-y-auto border-t border-[var(--editor-border)]">
            {resultPanel}
          </div>
        </div>
      </div>
    </section>
  );
}

function ActionBar({ onRun, onSubmit, isRunning, isMockJudge }: {
  onRun: () => void;
  onSubmit: () => void;
  isRunning: boolean;
  isMockJudge: boolean;
}) {
  return (
    <div className="flex items-center justify-end gap-2 border-b border-[var(--editor-border)] bg-[var(--editor-panel)] px-3 py-2">
      {isMockJudge && (
        <span className="mr-auto inline-flex items-center gap-1 rounded bg-amber-900/40 px-2 py-0.5 text-xs font-medium text-amber-300">
          <Zap aria-hidden="true" size={12} />Mock Judge
        </span>
      )}
      <button className="secondary-button gap-2" onClick={onRun} disabled={isRunning}>
        {isRunning ? <Loader2 aria-hidden="true" size={15} className="animate-spin" /> : <Play aria-hidden="true" size={15} />}
        {isRunning ? '运行中…' : '运行样例'}
      </button>
      <button className="primary-button gap-2" onClick={onSubmit}>
        <Send aria-hidden="true" size={15} />正式提交
      </button>
    </div>
  );
}

function ResultPanel({
  activeTab,
  onTabChange,
  inputMode,
  onInputModeChange,
  sampleCases,
  selectedSampleID,
  onSampleSelect,
  customInput,
  onCustomInputChange,
  customInputError,
  run,
  runError,
  onRetry,
  isMockJudge,
}: {
  activeTab: ResultTab;
  onTabChange: (tab: ResultTab) => void;
  inputMode: InputMode;
  onInputModeChange: (mode: InputMode) => void;
  sampleCases: SampleCase[];
  selectedSampleID: string;
  onSampleSelect: (id: string) => void;
  customInput: string;
  onCustomInputChange: (value: string) => void;
  customInputError: string;
  run: import('@/features/problems/lib/problems-api').SampleRunResponse | undefined;
  runError: string;
  onRetry: () => void;
  isMockJudge: boolean;
}) {
  return (
    <section className="min-h-full bg-[var(--editor-bg)] text-neutral-300">
      <div className="flex border-b border-[var(--editor-border)]" role="tablist" aria-label="运行结果区域">
        <ResultTabButton active={activeTab === 'input'} onClick={() => onTabChange('input')}>自测输入</ResultTabButton>
        <ResultTabButton active={activeTab === 'output'} onClick={() => onTabChange('output')}>输出</ResultTabButton>
        <ResultTabButton active={activeTab === 'result'} onClick={() => onTabChange('result')}>结果</ResultTabButton>
      </div>
      <div className="p-4 text-sm">
        {activeTab === 'input' && (
          <InputPanel
            inputMode={inputMode}
            onInputModeChange={onInputModeChange}
            sampleCases={sampleCases}
            selectedSampleID={selectedSampleID}
            onSampleSelect={onSampleSelect}
            customInput={customInput}
            onCustomInputChange={onCustomInputChange}
            customInputError={customInputError}
          />
        )}
        {activeTab === 'output' && (
          run?.output_data != null && run.output_data !== ''
            ? <pre className="whitespace-pre-wrap break-words font-mono text-sm leading-6 text-neutral-200">{run.output_data}</pre>
            : <p className="text-neutral-500">程序输出将在样例运行后显示。</p>
        )}
        {activeTab === 'result' && (
          <ResultStatus run={run} runError={runError} onRetry={onRetry} isMockJudge={isMockJudge} />
        )}
      </div>
    </section>
  );
}

function InputPanel({
  inputMode,
  onInputModeChange,
  sampleCases,
  selectedSampleID,
  onSampleSelect,
  customInput,
  onCustomInputChange,
  customInputError,
}: {
  inputMode: InputMode;
  onInputModeChange: (mode: InputMode) => void;
  sampleCases: SampleCase[];
  selectedSampleID: string;
  onSampleSelect: (id: string) => void;
  customInput: string;
  onCustomInputChange: (value: string) => void;
  customInputError: string;
}) {
  return (
    <div className="space-y-3">
      <div className="flex gap-3">
        <label className="flex items-center gap-1.5 text-xs">
          <input
            type="radio"
            name="input-mode"
            checked={inputMode === 'sample'}
            onChange={() => onInputModeChange('sample')}
            className="accent-[var(--accent)]"
          />
          公开样例
        </label>
        <label className="flex items-center gap-1.5 text-xs">
          <input
            type="radio"
            name="input-mode"
            checked={inputMode === 'custom'}
            onChange={() => onInputModeChange('custom')}
            className="accent-[var(--accent)]"
          />
          自定义输入
        </label>
      </div>

      {inputMode === 'sample' ? (
        sampleCases.length > 0 ? (
          <div>
            <select
              aria-label="选择公开样例"
              className="h-9 w-full border border-[var(--editor-border)] bg-[var(--editor-bg)] px-3 text-sm text-white"
              value={selectedSampleID}
              onChange={(e) => onSampleSelect(e.target.value)}
            >
              {sampleCases.map((sample, index) => (
                <option key={sample.id} value={sample.id}>样例 {index + 1}</option>
              ))}
            </select>
            {sampleCases.find((s) => s.id === selectedSampleID) && (
              <pre className="mt-2 max-h-[200px] overflow-auto whitespace-pre-wrap break-words rounded border border-[var(--editor-border)] bg-black/20 p-3 font-mono text-xs leading-5 text-neutral-300">
                {sampleCases.find((s) => s.id === selectedSampleID)!.input_data}
              </pre>
            )}
          </div>
        ) : (
          <p className="text-neutral-500">该题目暂无公开样例。</p>
        )
      ) : (
        <div>
          <textarea
            aria-label="自定义测试输入"
            className="h-[120px] w-full resize-y border border-[var(--editor-border)] bg-black/20 p-3 font-mono text-xs leading-5 text-neutral-200 placeholder:text-neutral-600 focus:outline-none focus:border-[var(--accent)]"
            placeholder="输入自定义测试数据…"
            value={customInput}
            onChange={(e) => onCustomInputChange(e.target.value)}
          />
          {customInputError && <p className="mt-1 text-xs text-[var(--danger)]">{customInputError}</p>}
        </div>
      )}
    </div>
  );
}

const statusConfig: Record<SampleRunStatus, { label: string; color: string; icon?: string }> = {
  QUEUED: { label: '排队中', color: 'text-blue-400' },
  RUNNING: { label: '运行中', color: 'text-blue-400' },
  AC: { label: '通过', color: 'text-green-400' },
  SYSTEM_ERROR: { label: '系统错误', color: 'text-red-400' },
};

function ResultStatus({
  run,
  runError,
  onRetry,
  isMockJudge,
}: {
  run: import('@/features/problems/lib/problems-api').SampleRunResponse | undefined;
  runError: string;
  onRetry: () => void;
  isMockJudge: boolean;
}) {
  if (runError) {
    return (
      <div className="space-y-2">
        <p className="text-[var(--danger)]">{runError}</p>
        <button className="secondary-button gap-1.5 text-xs" onClick={onRetry}>
          <RefreshCw aria-hidden="true" size={12} />重试
        </button>
      </div>
    );
  }

  if (!run) {
    return <p className="text-neutral-500">点击「运行样例」开始自测。</p>;
  }

  const config = statusConfig[run.status];
  const isTerminal = isTerminalRunStatus(run.status);

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        {run.status === 'RUNNING' && <Loader2 aria-hidden="true" size={16} className="animate-spin text-blue-400" />}
        <span className={`text-sm font-semibold ${config.color}`}>{config.label}</span>
        {isMockJudge && (
          <span className="rounded bg-amber-900/40 px-1.5 py-0.5 text-[10px] font-medium text-amber-300">Mock</span>
        )}
      </div>

      {isTerminal && run.error_message && (
        <pre className="max-h-[120px] overflow-auto whitespace-pre-wrap break-words rounded border border-[var(--editor-border)] bg-black/20 p-3 font-mono text-xs leading-5 text-red-300">
          {run.error_message}
        </pre>
      )}

      {isTerminal && run.status === 'AC' && (
        <pre className="max-h-[120px] overflow-auto whitespace-pre-wrap break-words rounded border border-[var(--editor-border)] bg-black/20 p-3 font-mono text-xs leading-5 text-green-300">
          {run.output_data || '(无输出)'}
        </pre>
      )}

      {isTerminal && (
        <button className="secondary-button gap-1.5 text-xs" onClick={onRetry}>
          <RefreshCw aria-hidden="true" size={12} />再次运行
        </button>
      )}
    </div>
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
