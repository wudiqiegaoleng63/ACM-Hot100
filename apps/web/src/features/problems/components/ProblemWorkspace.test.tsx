import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import ProblemWorkspace from './ProblemWorkspace';

const SPLIT_RATIO_KEY = 'problem-workspace:split-ratio';
const sampleCases = [
  { id: 'sample-1', order_index: 1, input_data: '1 2\n', expected_output: '3\n', explanation_md: '' },
  { id: 'sample-2', order_index: 2, input_data: '4 5\n', expected_output: '9\n', explanation_md: '' },
];

describe('ProblemWorkspace', () => {
  beforeEach(() => {
    localStorage.clear();
    setMobileViewport(false);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('restores and clamps the persisted desktop split ratio while dragging', () => {
    localStorage.setItem(SPLIT_RATIO_KEY, '60');
    renderWorkspace();

    const workspace = screen.getByTestId('desktop-workspace');
    vi.spyOn(workspace, 'getBoundingClientRect').mockReturnValue(rect(100, 1000));
    expect(workspace).toHaveStyle({ gridTemplateColumns: '60% 6px 40%' });

    const separator = screen.getByRole('separator', { name: '调整题面和代码宽度' });
    fireEvent.mouseDown(separator, { clientX: 700 });
    fireEvent.mouseMove(window, { clientX: 10 });
    expect(workspace).toHaveStyle({ gridTemplateColumns: '32% 6px 68%' });

    fireEvent.mouseMove(window, { clientX: 1200 });
    fireEvent.mouseUp(window);
    expect(workspace).toHaveStyle({ gridTemplateColumns: '68% 6px 32%' });
    expect(localStorage.getItem(SPLIT_RATIO_KEY)).toBe('68');
  });

  it('switches among statement, code, and result tabs on mobile', () => {
    setMobileViewport(true);
    renderWorkspace();

    expect(screen.getByText('题面内容')).toBeInTheDocument();
    expect(screen.queryByText('代码内容')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('tab', { name: '代码' }));
    expect(screen.getByText('代码内容')).toBeInTheDocument();
    expect(screen.queryByText('题面内容')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('tab', { name: '结果' }));
    expect(screen.getByRole('tab', { name: '自测输入' })).toBeInTheDocument();
  });

  it('sends unauthenticated run actions to login with the current location', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/problems/two-sum-target?tab=code']}>
          <Routes>
            <Route
              path="/problems/:slug"
              element={(
                <ProblemWorkspace
                  statement={<div>题面内容</div>}
                  editor={<div>代码内容</div>}
                  previous={null}
                  next={null}
                  isAuthenticated={false}
                  problemSlug="two-sum-target"
                  sampleCases={sampleCases}
                  languageKey="cpp17"
                  sourceCode="int main() {}"
                />
              )}
            />
            <Route path="/login" element={<LocationState />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    );

    fireEvent.click(screen.getByRole('button', { name: '运行样例' }));
    expect(screen.getByTestId('login-return-location')).toHaveTextContent(
      'two-sum-target?tab=code',
    );
  });

  it('sends unauthenticated formal submissions to login with the current location', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/problems/two-sum-target?tab=code']}>
          <Routes>
            <Route
              path="/problems/:slug"
              element={(
                <ProblemWorkspace
                  statement={<div>题面内容</div>}
                  editor={<div>代码内容</div>}
                  previous={null}
                  next={null}
                  isAuthenticated={false}
                  problemSlug="two-sum-target"
                  sampleCases={sampleCases}
                  languageKey="cpp17"
                  sourceCode="int main() {}"
                />
              )}
            />
            <Route path="/login" element={<LocationState />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    );

    fireEvent.click(screen.getByRole('button', { name: '正式提交' }));
    expect(screen.getByTestId('login-return-location')).toHaveTextContent(
      'two-sum-target?tab=code',
    );
  });

  it('submits source, switches to result, and renders an AC with metrics', async () => {
    vi.stubGlobal('fetch', submissionFetch([
      submissionPayload('AC'),
    ]));
    renderWorkspace();

    fireEvent.click(screen.getByRole('button', { name: '正式提交' }));

    await waitFor(() => expect(screen.getByText('答案正确')).toBeInTheDocument());
    expect(screen.getByText(/通过 8\/8 个测试点/)).toHaveTextContent('总耗时 42 ms');
    expect(screen.getByText(/通过 8\/8 个测试点/)).toHaveTextContent('峰值内存 2 MB');
  });

  it('keeps the formal result visible when progress invalidation refetches the problem', async () => {
    vi.stubGlobal('fetch', submissionFetch([submissionPayload('AC')]));
    const { rerender } = renderWorkspace();

    fireEvent.click(screen.getByRole('button', { name: '正式提交' }));
    await waitFor(() => expect(screen.getByText('答案正确')).toBeInTheDocument());

    rerender(workspaceElement([{ ...sampleCases[0]! }, { ...sampleCases[1]! }]));
    expect(screen.getByText('答案正确')).toBeInTheDocument();
  });

  it('prevents duplicate submissions while judging and displays queued text with an icon', async () => {
    const fetchMock = submissionFetch([submissionPayload('QUEUED')]);
    vi.stubGlobal('fetch', fetchMock);
    renderWorkspace();

    const submitButton = screen.getByRole('button', { name: '正式提交' });
    fireEvent.click(submitButton);

    await waitFor(() => expect(screen.getByText('等待判题')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '判题中…' })).toBeDisabled();
    expect(screen.getByText('等待判题').previousElementSibling?.tagName).toBe('svg');
    const submitCalls = fetchMock.mock.calls.filter(([input]) => String(input).includes('/submissions'));
    expect(submitCalls).toHaveLength(2);
  });

  it('shows a failed creation and lets the user retry', async () => {
    let createCalls = 0;
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockImplementation((input, init) => {
      const url = String(input);
      if (url === '/api/v1/health') return Promise.resolve(jsonResponse(healthPayload));
      if (url.includes('/submissions') && init?.method === 'POST') {
        createCalls++;
        if (createCalls === 1) return Promise.resolve(jsonResponse({ error: { code: 'FAILED', message: '网络不可用' } }, 503));
        return Promise.resolve(jsonResponse(createResponse));
      }
      if (url === '/api/v1/submissions/submission-1') return Promise.resolve(jsonResponse(submissionPayload('AC')));
      return Promise.reject(new Error(`unexpected request: ${url}`));
    }));
    renderWorkspace();

    fireEvent.click(screen.getByRole('button', { name: '正式提交' }));
    await waitFor(() => expect(screen.getByText(/提交失败/)).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '重新提交' }));
    await waitFor(() => expect(screen.getByText('答案正确')).toBeInTheDocument());
    expect(createCalls).toBe(2);
  });

  it.each([
    ['WA', '答案错误'],
    ['TLE', '时间超限'],
    ['MLE', '内存超限'],
    ['RE', '运行错误'],
    ['CE', '编译错误'],
    ['SYSTEM_ERROR', '系统判题失败'],
  ] as const)('renders %s with text and an icon', async (status, label) => {
    vi.stubGlobal('fetch', submissionFetch([submissionPayload(status)]));
    renderWorkspace();

    fireEvent.click(screen.getByRole('button', { name: '正式提交' }));

    await waitFor(() => expect(screen.getByText(label)).toBeInTheDocument());
    expect(screen.getByText(label).previousElementSibling?.tagName).toBe('svg');
  });

  it('shows only the failed hidden case index for WA', async () => {
    vi.stubGlobal('fetch', submissionFetch([submissionPayload('WA')]));
    renderWorkspace();

    fireEvent.click(screen.getByRole('button', { name: '正式提交' }));

    await waitFor(() => expect(screen.getByText(/首个未通过测试点：#3/)).toBeInTheDocument());
    expect(screen.getByText(/首个未通过测试点：#3/)).toHaveTextContent('隐藏测试输入与输出不会公开');
    expect(screen.queryByText('secret input')).not.toBeInTheDocument();
  });

  it('copies compiler output and labels server-side truncation', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } });
    vi.stubGlobal('fetch', submissionFetch([submissionPayload('CE')]));
    renderWorkspace();

    fireEvent.click(screen.getByRole('button', { name: '正式提交' }));
    await waitFor(() => expect(screen.getByText('编译输出（已截断至 8KB）')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '复制' }));

    await waitFor(() => expect(writeText).toHaveBeenCalledWith("main.cpp:1: error\n... [truncated]"));
    expect(screen.getByText('已复制')).toBeInTheDocument();
  });

  it('shows sample case selector and custom input toggle in input tab', () => {
    renderWorkspace();

    fireEvent.click(screen.getByRole('tab', { name: '自测输入' }));
    expect(screen.getByLabelText('选择公开样例')).toBeInTheDocument();
    expect(screen.getByRole('radio', { name: '公开样例' })).toBeInTheDocument();
    expect(screen.getByRole('radio', { name: '自定义输入' })).toBeInTheDocument();
  });

  it('switches between sample and custom input modes', () => {
    renderWorkspace();

    fireEvent.click(screen.getByRole('tab', { name: '自测输入' }));
    expect(screen.getByLabelText('选择公开样例')).toBeInTheDocument();

    // Click the "自定义输入" radio button (not the textarea)
    const customRadio = screen.getByRole('radio', { name: '自定义输入' });
    fireEvent.click(customRadio);
    expect(screen.getByLabelText('自定义测试输入')).toBeInTheDocument();
    expect(screen.queryByLabelText('选择公开样例')).not.toBeInTheDocument();
  });

  it('shows custom input size error when input exceeds 16KB', () => {
    renderWorkspace();

    fireEvent.click(screen.getByRole('tab', { name: '自测输入' }));
    fireEvent.click(screen.getByRole('radio', { name: '自定义输入' }));

    const textarea = screen.getByLabelText('自定义测试输入');
    fireEvent.change(textarea, { target: { value: 'x'.repeat(16 * 1024 + 1) } });

    expect(screen.getByText('自定义输入不能超过 16KB。')).toBeInTheDocument();
  });

  it('shows Mock Judge badge when health endpoint reports mock mode', async () => {
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockImplementation((input) => {
      if (String(input) === '/api/v1/health') {
        return Promise.resolve(new Response(JSON.stringify({
          status: 'ok', services: { mysql: 'ok', redis: 'ok' }, judge_mode: 'mock',
        }), { headers: { 'Content-Type': 'application/json' } }));
      }
      return Promise.reject(new Error('unexpected'));
    }));
    renderWorkspace();

    await waitFor(() => {
      expect(screen.getByText('Mock Judge')).toBeInTheDocument();
    });
  });

  it('does not show Mock Judge badge when judge mode is judge0', async () => {
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockImplementation((input) => {
      if (String(input) === '/api/v1/health') {
        return Promise.resolve(new Response(JSON.stringify({
          status: 'ok', services: { mysql: 'ok', redis: 'ok' }, judge_mode: 'judge0',
        }), { headers: { 'Content-Type': 'application/json' } }));
      }
      return Promise.reject(new Error('unexpected'));
    }));
    renderWorkspace();

    await waitFor(() => {
      expect(screen.queryByText('Mock Judge')).not.toBeInTheDocument();
    });
  });
});

const healthPayload = { status: 'ok', services: { mysql: 'ok', redis: 'ok' }, judge_mode: 'judge0' };
const createResponse = { id: 'submission-1', status: 'QUEUED', created_at: '2026-07-14T00:00:00Z' };

function submissionPayload(status: 'QUEUED' | 'AC' | 'WA' | 'TLE' | 'MLE' | 'RE' | 'CE' | 'SYSTEM_ERROR') {
  const isTerminal = status !== 'QUEUED';
  return {
    id: 'submission-1',
    problem_slug: 'two-sum-target',
    problem_title: '两数目标和',
    language_key: 'cpp17',
    source_code: 'int main() {}',
    status,
    passed_cases: status === 'AC' ? 8 : status === 'WA' ? 2 : 0,
    total_cases: isTerminal ? 8 : 0,
    time_ms: isTerminal ? 42 : null,
    memory_kb: isTerminal ? 2048 : null,
    compiler_output: status === 'CE' ? "main.cpp:1: error\n... [truncated]" : '',
    error_message: status === 'RE' ? 'Runtime Error (SIGSEGV)' : '',
    case_results: status === 'WA'
      ? [{ case_index: 2, status: 'WA', time_ms: 10, memory_kb: 1024, is_sample: false }]
      : [],
    created_at: '2026-07-14T00:00:00Z',
    judged_at: isTerminal ? '2026-07-14T00:00:01Z' : null,
  };
}

function submissionFetch(responses: ReturnType<typeof submissionPayload>[]) {
  let detailIndex = 0;
  return vi.fn<typeof fetch>().mockImplementation((input, init) => {
    const url = String(input);
    if (url === '/api/v1/health') return Promise.resolve(jsonResponse(healthPayload));
    if (url.includes('/submissions') && init?.method === 'POST') return Promise.resolve(jsonResponse(createResponse, 202));
    if (url === '/api/v1/submissions/submission-1') {
      const response = responses[Math.min(detailIndex, responses.length - 1)];
      detailIndex++;
      return Promise.resolve(jsonResponse(response));
    }
    return Promise.reject(new Error(`unexpected request: ${url}`));
  });
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), { status, headers: { 'Content-Type': 'application/json' } });
}

function renderWorkspace() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(workspaceElement(sampleCases, queryClient));
}

function workspaceElement(cases = sampleCases, queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })) {
  return (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <ProblemWorkspace
          statement={<div>题面内容</div>}
          editor={<div>代码内容</div>}
          previous={{ slug: 'previous', title: '上一题' }}
          next={{ slug: 'next', title: '下一题' }}
          isAuthenticated
          problemSlug="two-sum-target"
          sampleCases={cases}
          languageKey="cpp17"
          sourceCode="int main() {}"
          timeLimitMs={1000}
          memoryLimitKb={262144}
        />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

function LocationState() {
  const location = useLocation();
  const from = (location.state as { from?: { pathname?: string; search?: string } } | null)?.from;
  return <output data-testid="login-return-location">{`${from?.pathname}${from?.search}`}</output>;
}

function setMobileViewport(matches: boolean) {
  const mediaQuery: MediaQueryList = {
    matches,
    media: '(max-width: 767px)',
    onchange: null,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
  };
  vi.stubGlobal('matchMedia', vi.fn(() => mediaQuery));
}

function rect(left: number, width: number): DOMRect {
  return {
    x: left,
    y: 0,
    left,
    right: left + width,
    top: 0,
    bottom: 600,
    width,
    height: 600,
    toJSON: () => ({}),
  };
}
