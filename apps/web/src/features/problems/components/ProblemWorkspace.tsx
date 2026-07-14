import { ArrowLeft, ArrowRight, CheckCircle2, Clock3, Copy, Loader2, Play, RefreshCw, Send, XCircle, Zap } from 'lucide-react';
import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from 'react';
import { Link, useLocation, useNavigate } from 'react-router';

import { useCreateSampleRun, useSampleRun, useHealth, useCreateSubmission, useSubmission } from '@/features/problems/hooks/use-problems';
import type { ProblemNavigation, SampleCase, SampleRunStatus, SubmissionDetail, SubmissionStatus } from '@/features/problems/lib/problems-api';
import { isTerminalRunStatus, isTerminalSubmissionStatus } from '@/features/problems/lib/problems-api';

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
  userID?: string;
  problemSlug: string;
  sampleCases: SampleCase[];
  languageKey: string;
  sourceCode: string;
  timeLimitMs?: number;
  memoryLimitKb?: number;
}

export default function ProblemWorkspace({
  statement,
  editor,
  previous,
  next,
  isAuthenticated,
  userID,
  problemSlug,
  sampleCases,
  languageKey,
  sourceCode,
  timeLimitMs,
  memoryLimitKb,
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
  const [submissionID, setSubmissionID] = useState<string | null>(null);
  const [submissionMode, setSubmissionMode] = useState(false);
  const [submissionError, setSubmissionError] = useState('');

  const createRun = useCreateSampleRun();
  const sampleRun = useSampleRun(activeRunID);
  const createSubmission = useCreateSubmission();
  const submissionQuery = useSubmission(submissionID, userID);
  const health = useHealth();
  const isMockJudge = health.data?.judge_mode === 'mock';

  // Reset sample selection when sample cases change
  useEffect(() => {
    const firstID = sampleCases[0]?.id;
    if (firstID && !sampleCases.find((sample) => sample.id === selectedSampleID)) {
      setSelectedSampleID(firstID);
    }
  }, [sampleCases, selectedSampleID]);

  // Clear run and submission state only when navigating to a different problem.
  useEffect(() => {
    setActiveRunID(null);
    setRunError('');
    setInputMode('sample');
    setCustomInput('');
    setCustomInputError('');
    setSubmissionID(null);
    setSubmissionMode(false);
    setSubmissionError('');
  }, [problemSlug]);

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

  const handleSubmit = async () => {
    if (!isAuthenticated) {
      navigate('/login', { state: { from: location } });
      return;
    }
    if (!languageKey || !sourceCode.trim()) return;

    setSubmissionMode(true);
    setSubmissionError('');
    setSubmissionID(null);
    setResultTab('result');
    if (isMobile) setMobileTab('result');

    try {
      const result = await createSubmission.mutateAsync({
        slug: problemSlug,
        languageKey,
        sourceCode,
      });
      setSubmissionID(result.id);
    } catch (err) {
      setSubmissionError(err instanceof Error ? err.message : '提交请求失败');
    }
  };

  const runStatus = sampleRun.data?.status;
  const isRunning = createRun.isPending || (activeRunID != null && (!runStatus || !isTerminalRunStatus(runStatus)));
  const submissionStatus = submissionQuery.data?.status;
  const isSubmitting = createSubmission.isPending || (submissionID != null && (!submissionStatus || !isTerminalSubmissionStatus(submissionStatus)));

  const actionBar = (
    <ActionBar
      onRun={handleRun}
      onSubmit={handleSubmit}
      isRunning={isRunning}
      isSubmitting={isSubmitting}
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
      hasSubmission={submissionMode}
      submission={submissionQuery.data}
      submissionCreated={submissionID != null}
      submissionError={submissionError}
      submissionRefreshError={submissionQuery.error instanceof Error ? submissionQuery.error.message : ''}
      submissionPollTimedOut={submissionQuery.pollTimedOut}
      onRefreshSubmission={() => void submissionQuery.refresh()}
      onRetrySubmission={() => void handleSubmit()}
      timeLimitMs={timeLimitMs}
      memoryLimitKb={memoryLimitKb}
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

function ActionBar({ onRun, onSubmit, isRunning, isSubmitting, isMockJudge }: {
  onRun: () => void;
  onSubmit: () => void;
  isRunning: boolean;
  isSubmitting: boolean;
  isMockJudge: boolean;
}) {
  return (
    <div className="flex items-center justify-end gap-2 border-b border-[var(--editor-border)] bg-[var(--editor-panel)] px-3 py-2">
      {isMockJudge && (
        <span className="mr-auto inline-flex items-center gap-1 rounded bg-amber-900/40 px-2 py-0.5 text-xs font-medium text-amber-300">
          <Zap aria-hidden="true" size={12} />Mock Judge
        </span>
      )}
      <button className="secondary-button gap-2 disabled:cursor-not-allowed disabled:opacity-60" onClick={onRun} disabled={isRunning || isSubmitting}>
        {isRunning ? <Loader2 aria-hidden="true" size={15} className="animate-spin" /> : <Play aria-hidden="true" size={15} />}
        {isRunning ? '运行中…' : '运行样例'}
      </button>
      <button className="primary-button gap-2 disabled:cursor-not-allowed disabled:opacity-60" onClick={onSubmit} disabled={isRunning || isSubmitting}>
        {isSubmitting ? <Loader2 aria-hidden="true" size={15} className="animate-spin" /> : <Send aria-hidden="true" size={15} />}
        {isSubmitting ? '判题中…' : '正式提交'}
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
  hasSubmission,
  submission,
  submissionCreated,
  submissionError,
  submissionRefreshError,
  submissionPollTimedOut,
  onRefreshSubmission,
  onRetrySubmission,
  timeLimitMs,
  memoryLimitKb,
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
  hasSubmission: boolean;
  submission: import('@/features/problems/lib/problems-api').SubmissionDetail | undefined;
  submissionCreated: boolean;
  submissionError: string;
  submissionRefreshError: string;
  submissionPollTimedOut: boolean;
  onRefreshSubmission: () => void;
  onRetrySubmission: () => void;
  timeLimitMs?: number;
  memoryLimitKb?: number;
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
          run?.output_data
            ? <pre className="max-h-[200px] overflow-auto whitespace-pre-wrap break-words font-mono text-sm leading-6 text-neutral-200">{run.output_data}</pre>
            : <p className="text-neutral-500">样例或自测输出将在运行后显示。</p>
        )}
        {activeTab === 'result' && (
          hasSubmission
            ? <SubmissionResultStatus
                submission={submission}
                submissionCreated={submissionCreated}
                submissionError={submissionError}
                refreshError={submissionRefreshError}
                pollTimedOut={submissionPollTimedOut}
                onRefresh={onRefreshSubmission}
                onRetry={onRetrySubmission}
                timeLimitMs={timeLimitMs}
                memoryLimitKb={memoryLimitKb}
                isMockJudge={isMockJudge}
              />
            : <ResultStatus run={run} runError={runError} onRetry={onRetry} isMockJudge={isMockJudge} />
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

const submissionStatusConfig: Record<SubmissionStatus, { label: string; color: string; icon: 'clock' | 'active' | 'success' | 'failure' }> = {
  QUEUED: { label: '等待判题', color: 'text-neutral-300', icon: 'clock' },
  COMPILING: { label: '正在编译', color: 'text-blue-400', icon: 'active' },
  RUNNING: { label: '正在运行', color: 'text-blue-400', icon: 'active' },
  AC: { label: '答案正确', color: 'text-green-400', icon: 'success' },
  WA: { label: '答案错误', color: 'text-red-400', icon: 'failure' },
  TLE: { label: '时间超限', color: 'text-red-400', icon: 'failure' },
  MLE: { label: '内存超限', color: 'text-red-400', icon: 'failure' },
  RE: { label: '运行错误', color: 'text-red-400', icon: 'failure' },
  CE: { label: '编译错误', color: 'text-red-400', icon: 'failure' },
  SYSTEM_ERROR: { label: '系统判题失败', color: 'text-red-400', icon: 'failure' },
};

function SubmissionResultStatus({
  submission,
  submissionCreated,
  submissionError,
  refreshError,
  pollTimedOut,
  onRefresh,
  onRetry,
  timeLimitMs,
  memoryLimitKb,
  isMockJudge,
}: {
  submission: SubmissionDetail | undefined;
  submissionCreated: boolean;
  submissionError: string;
  refreshError: string;
  pollTimedOut: boolean;
  onRefresh: () => void;
  onRetry: () => void;
  timeLimitMs?: number;
  memoryLimitKb?: number;
  isMockJudge: boolean;
}) {
  if (submissionError && !submissionCreated) {
    return (
      <div className="space-y-2" role="alert">
        <p className="text-red-400">提交失败：{submissionError}</p>
        <button className="secondary-button gap-1.5 text-xs" onClick={onRetry}>
          <RefreshCw aria-hidden="true" size={12} />重新提交
        </button>
      </div>
    );
  }

  if (!submission && refreshError) {
    return (
      <div className="space-y-2" role="alert">
        <p className="text-red-400">提交已创建，但判题状态加载失败：{refreshError}</p>
        <button className="secondary-button gap-1.5 text-xs" onClick={onRefresh}>
          <RefreshCw aria-hidden="true" size={12} />重新加载状态
        </button>
      </div>
    );
  }

  if (!submission) {
    return (
      <div className="flex items-center gap-2" role="status">
        <Loader2 aria-hidden="true" size={16} className="animate-spin text-blue-400" />
        <span className="text-sm text-blue-400">正在创建提交…</span>
      </div>
    );
  }

  const config = submissionStatusConfig[submission.status];
  const isTerminal = isTerminalSubmissionStatus(submission.status);
  const firstFailedCase = submission.case_results.find((result) => result.status !== 'AC');

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-2" role="status">
        <SubmissionStatusIcon icon={config.icon} color={config.color} />
        <span className={`text-sm font-semibold ${config.color}`}>{config.label}</span>
        {isMockJudge && (
          <span className="rounded bg-amber-900/40 px-1.5 py-0.5 text-[10px] font-medium text-amber-300">Mock</span>
        )}
      </div>

      {submission.status === 'QUEUED' && (
        <p className="text-xs text-neutral-400">提交已入队，正在等待可用的判题 Worker。</p>
      )}
      {submission.status === 'RUNNING' && (
        <p className="text-xs text-neutral-300">
          当前进度：{submission.passed_cases}/{submission.total_cases || '—'} 个测试点
        </p>
      )}
      {isTerminal && submission.total_cases > 0 && (
        <p className="text-xs text-neutral-300">
          通过 {submission.passed_cases}/{submission.total_cases} 个测试点
          {submission.time_ms != null && ` · 总耗时 ${submission.time_ms} ms`}
          {submission.memory_kb != null && ` · 峰值内存 ${formatMemory(submission.memory_kb)}`}
        </p>
      )}

      {submission.status === 'WA' && firstFailedCase && (
        <p className="text-xs text-red-300">
          首个未通过测试点：#{firstFailedCase.case_index + 1}。隐藏测试输入与输出不会公开。
        </p>
      )}
      {submission.status === 'TLE' && (
        <p className="text-xs text-red-300">
          程序超过时间限制{timeLimitMs != null ? ` ${timeLimitMs} ms` : ''}
          {submission.time_ms != null ? `，本次累计耗时 ${submission.time_ms} ms。` : '。'}
        </p>
      )}
      {submission.status === 'MLE' && (
        <p className="text-xs text-red-300">
          程序超过内存限制{memoryLimitKb != null ? ` ${formatMemory(memoryLimitKb)}` : ''}
          {submission.memory_kb != null ? `，本次峰值 ${formatMemory(submission.memory_kb)}。` : '。'}
        </p>
      )}
      {submission.status === 'RE' && !submission.error_message && (
        <p className="text-xs text-red-300">程序运行时异常；错误信息中的宿主机路径已清理。</p>
      )}
      {submission.status === 'SYSTEM_ERROR' && (
        <p className="text-xs text-red-300">系统判题失败，可重新提交。此次系统错误不代表代码答案错误。</p>
      )}

      {submission.compiler_output && (
        <CopyableOutput label="编译输出" value={submission.compiler_output} />
      )}
      {submission.error_message && (
        <CopyableOutput label="错误信息" value={submission.error_message} />
      )}

      {refreshError && (
        <div className="flex flex-wrap items-center gap-2 text-xs text-red-300" role="alert">
          <span>刷新判题状态失败：{refreshError}</span>
          <button className="secondary-button gap-1.5 text-xs" onClick={onRefresh}>
            <RefreshCw aria-hidden="true" size={12} />重试刷新
          </button>
        </div>
      )}
      {pollTimedOut && !isTerminal && (
        <div className="flex flex-wrap items-center gap-2 text-xs text-neutral-300" role="alert">
          <span>已等待 60 秒，自动刷新已停止。</span>
          <button className="secondary-button gap-1.5 text-xs" onClick={onRefresh}>
            <RefreshCw aria-hidden="true" size={12} />刷新状态
          </button>
        </div>
      )}

      {isTerminal && (
        <button className="secondary-button gap-1.5 text-xs" onClick={onRetry}>
          <RefreshCw aria-hidden="true" size={12} />再次提交
        </button>
      )}
    </div>
  );
}

function SubmissionStatusIcon({ icon, color }: { icon: 'clock' | 'active' | 'success' | 'failure'; color: string }) {
  const className = icon === 'active' ? `animate-spin ${color}` : color;
  if (icon === 'clock') return <Clock3 aria-hidden="true" size={17} className={className} />;
  if (icon === 'active') return <Loader2 aria-hidden="true" size={17} className={className} />;
  if (icon === 'success') return <CheckCircle2 aria-hidden="true" size={17} className={className} />;
  return <XCircle aria-hidden="true" size={17} className={className} />;
}

function CopyableOutput({ label, value }: { label: string; value: string }) {
  const [copyMessage, setCopyMessage] = useState('');
  const truncated = value.includes('[truncated]');

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopyMessage('已复制');
    } catch {
      setCopyMessage('复制失败，请手动选择文本');
    }
  };

  return (
    <section className="overflow-hidden rounded border border-[var(--editor-border)] bg-black/20">
      <div className="flex items-center justify-between gap-3 border-b border-[var(--editor-border)] px-3 py-2">
        <span className="text-xs font-medium text-neutral-300">
          {label}{truncated ? '（已截断至 8KB）' : ''}
        </span>
        <div className="flex items-center gap-2">
          {copyMessage && <span className="text-[10px] text-neutral-400">{copyMessage}</span>}
          <button className="inline-flex items-center gap-1 text-xs text-neutral-300 hover:text-white" onClick={copy}>
            <Copy aria-hidden="true" size={12} />复制
          </button>
        </div>
      </div>
      <pre className="max-h-[160px] overflow-auto whitespace-pre-wrap break-words p-3 font-mono text-xs leading-5 text-red-300">
        {value}
      </pre>
    </section>
  );
}

function formatMemory(memoryKb: number) {
  return `${Math.max(0, memoryKb / 1024).toFixed(memoryKb % 1024 === 0 ? 0 : 1)} MB`;
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
