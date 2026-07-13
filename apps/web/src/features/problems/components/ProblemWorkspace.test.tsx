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

function renderWorkspace() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <ProblemWorkspace
          statement={<div>题面内容</div>}
          editor={<div>代码内容</div>}
          previous={{ slug: 'previous', title: '上一题' }}
          next={{ slug: 'next', title: '下一题' }}
          isAuthenticated
          problemSlug="two-sum-target"
          sampleCases={sampleCases}
          languageKey="cpp17"
          sourceCode="int main() {}"
        />
      </MemoryRouter>
    </QueryClientProvider>,
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
