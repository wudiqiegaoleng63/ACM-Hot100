import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import ProblemWorkspace from './ProblemWorkspace';

const SPLIT_RATIO_KEY = 'problem-workspace:split-ratio';

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
    const resultTabs = screen.getAllByRole('tab', { name: '结果' });
    expect(resultTabs).toHaveLength(2);
    fireEvent.click(resultTabs[1] as HTMLElement);
    expect(screen.getByText('样例运行后在这里显示判题结果。')).toBeInTheDocument();
  });

  it('sends unauthenticated run actions to login with the current location', () => {
    render(
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
              />
            )}
          />
          <Route path="/login" element={<LocationState />} />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole('button', { name: '运行样例' }));
    expect(screen.getByTestId('login-return-location')).toHaveTextContent(
      'two-sum-target?tab=code',
    );
  });
});

function renderWorkspace() {
  return render(
    <MemoryRouter>
      <ProblemWorkspace
        statement={<div>题面内容</div>}
        editor={<div>代码内容</div>}
        previous={{ slug: 'previous', title: '上一题' }}
        next={{ slug: 'next', title: '下一题' }}
        isAuthenticated
      />
    </MemoryRouter>,
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
