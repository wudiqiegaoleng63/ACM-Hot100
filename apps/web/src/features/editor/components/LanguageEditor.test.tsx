import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import LanguageEditor from './LanguageEditor';

vi.mock('./CodeEditor', () => ({
  default: ({
    value,
    language,
    onChange,
  }: {
    value: string;
    language: string;
    onChange: (value: string) => void;
  }) => (
    <div>
      <output data-testid="editor-language">{language}</output>
      <output data-testid="editor-value">{value}</output>
      <button onClick={() => onChange('modified source')}>修改代码</button>
      <button onClick={() => onChange('a'.repeat(64 * 1024 + 1))}>写入超限代码</button>
    </div>
  ),
}));

const languages = [
  {
    key: 'cpp17',
    display_name: 'C++17',
    editor_language: 'cpp',
    source_template: 'cpp template',
  },
  {
    key: 'python3',
    display_name: 'Python 3',
    editor_language: 'python',
    source_template: 'python template',
  },
];

describe('LanguageEditor', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockImplementation((input) => {
      if (String(input) === '/api/v1/languages') return Promise.resolve(jsonResponse(languages));
      return Promise.reject(new Error(`unexpected request ${String(input)}`));
    }));
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('initializes the template and Monaco language from /languages', async () => {
    renderEditor();

    expect(screen.getByText('正在加载语言配置…')).toBeInTheDocument();
    expect(await screen.findByTestId('editor-language')).toHaveTextContent('cpp');
    expect(screen.getByTestId('editor-value')).toHaveTextContent('cpp template');
    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/languages',
      expect.objectContaining({ credentials: 'include' }),
    );
  });

  it('switches an untouched template without confirmation', async () => {
    const confirm = vi.spyOn(window, 'confirm');
    renderEditor();

    fireEvent.change(await screen.findByLabelText('编程语言'), { target: { value: 'python3' } });

    expect(confirm).not.toHaveBeenCalled();
    expect(screen.getByTestId('editor-language')).toHaveTextContent('python');
    expect(screen.getByTestId('editor-value')).toHaveTextContent('python template');
  });

  it('switches modified source only after confirmation', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    renderEditor();

    fireEvent.click(await screen.findByRole('button', { name: '修改代码' }));
    fireEvent.change(screen.getByLabelText('编程语言'), { target: { value: 'python3' } });

    expect(window.confirm).toHaveBeenCalledOnce();
    expect(screen.getByTestId('editor-language')).toHaveTextContent('python');
    expect(screen.getByTestId('editor-value')).toHaveTextContent('python template');
  });

  it('keeps language and modified source when confirmation is canceled', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false);
    renderEditor();

    fireEvent.click(await screen.findByRole('button', { name: '修改代码' }));
    fireEvent.change(screen.getByLabelText('编程语言'), { target: { value: 'python3' } });

    expect(screen.getByLabelText('编程语言')).toHaveValue('cpp17');
    expect(screen.getByTestId('editor-language')).toHaveTextContent('cpp');
    expect(screen.getByTestId('editor-value')).toHaveTextContent('modified source');
  });

  it('restores the newer local draft for a guest and never requests the server draft', async () => {
    localStorage.setItem(
      'draft:guest:two-sum-target:cpp17',
      JSON.stringify({ source_code: 'local guest source', updated_at: '2026-07-13T14:40:00Z' }),
    );
    renderEditor();

    expect(await screen.findByTestId('editor-value')).toHaveTextContent('local guest source');
    expect(fetch).not.toHaveBeenCalledWith(
      expect.stringContaining('/drafts/'),
      expect.objectContaining({ method: 'GET' }),
    );
  });

  it('chooses the newer server draft for an authenticated user', async () => {
    localStorage.setItem(
      'draft:user-1:two-sum-target:cpp17',
      JSON.stringify({ source_code: 'older local source', updated_at: '2026-07-13T14:40:00Z' }),
    );
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockImplementation((input) => {
      if (String(input) === '/api/v1/languages') return Promise.resolve(jsonResponse(languages));
      if (String(input).endsWith('/drafts/cpp17')) {
        return Promise.resolve(jsonResponse({
          source_code: 'newer server source',
          language_key: 'cpp17',
          updated_at: '2026-07-13T14:41:00Z',
        }));
      }
      return Promise.reject(new Error(`unexpected request ${String(input)}`));
    }));
    renderEditor('user-1');

    expect(await screen.findByTestId('editor-value')).toHaveTextContent('newer server source');
  });

  it('saves locally immediately and sends one server save after 500ms', async () => {
    const fetchMock = vi.fn<typeof fetch>().mockImplementation((input, init) => {
      if (String(input) === '/api/v1/languages') return Promise.resolve(jsonResponse(languages));
      if (String(input).endsWith('/drafts/cpp17') && init?.method === 'GET') {
        return Promise.resolve(jsonResponse({
          source_code: 'cpp template',
          language_key: 'cpp17',
          updated_at: '2026-07-13T14:40:00Z',
        }));
      }
      if (String(input).endsWith('/drafts/cpp17') && init?.method === 'PUT') {
        return Promise.resolve(jsonResponse({
          source_code: 'modified source',
          language_key: 'cpp17',
          updated_at: '2026-07-13T14:41:00Z',
        }));
      }
      return Promise.reject(new Error(`unexpected request ${String(input)}`));
    });
    vi.stubGlobal('fetch', fetchMock);
    renderEditor('user-1');
    expect(await screen.findByRole('button', { name: '修改代码' })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '修改代码' }));
    expect(JSON.parse(localStorage.getItem('draft:user-1:two-sum-target:cpp17') ?? '{}')).toMatchObject({
      source_code: 'modified source',
    });
    expect(putCalls(fetchMock)).toHaveLength(0);

    await act(async () => new Promise((resolve) => setTimeout(resolve, 550)));
    expect(putCalls(fetchMock)).toHaveLength(1);
  });

  it('keeps local source and shows a non-blocking warning when server save fails', async () => {
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockImplementation((input, init) => {
      if (String(input) === '/api/v1/languages') return Promise.resolve(jsonResponse(languages));
      if (String(input).endsWith('/drafts/cpp17') && init?.method === 'GET') {
        return Promise.resolve(jsonResponse({
          source_code: 'cpp template', language_key: 'cpp17', updated_at: '2026-07-13T14:40:00Z',
        }));
      }
      if (String(input).endsWith('/drafts/cpp17') && init?.method === 'PUT') {
        return Promise.resolve(jsonResponse({ error: { code: 'INTERNAL_ERROR', message: 'failed' }, request_id: '1' }, 500));
      }
      return Promise.reject(new Error(`unexpected request ${String(input)}`));
    }));
    renderEditor('user-1');
    expect(await screen.findByRole('button', { name: '修改代码' })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '修改代码' }));
    await act(async () => new Promise((resolve) => setTimeout(resolve, 550)));

    expect(await screen.findByText('服务器草稿保存失败，本地草稿已保留。')).toBeInTheDocument();
    expect(screen.getByTestId('editor-value')).toHaveTextContent('modified source');
  });

  it('flushes a pending save before switching language', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    const fetchMock = vi.fn<typeof fetch>().mockImplementation((input, init) => {
      if (String(input) === '/api/v1/languages') return Promise.resolve(jsonResponse(languages));
      if (String(input).includes('/drafts/') && init?.method === 'GET') {
        return Promise.resolve(jsonResponse({ error: { code: 'NOT_FOUND', message: 'missing' }, request_id: '1' }, 404));
      }
      if (String(input).endsWith('/drafts/cpp17') && init?.method === 'PUT') {
        return Promise.resolve(jsonResponse({ source_code: 'modified source', language_key: 'cpp17', updated_at: '2026-07-13T14:41:00Z' }));
      }
      if (String(input).endsWith('/drafts/python3') && init?.method === 'PUT') {
        return Promise.resolve(jsonResponse({ source_code: 'python template', language_key: 'python3', updated_at: '2026-07-13T14:42:00Z' }));
      }
      return Promise.reject(new Error(`unexpected request ${String(input)}`));
    });
    vi.stubGlobal('fetch', fetchMock);
    renderEditor('user-1');
    expect(await screen.findByRole('button', { name: '修改代码' })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '修改代码' }));
    fireEvent.change(screen.getByLabelText('编程语言'), { target: { value: 'python3' } });

    await waitFor(() => expect(putCalls(fetchMock)).toHaveLength(1));
    expect(String(putCalls(fetchMock)[0]?.[0])).toContain('/drafts/cpp17');
  });

  it('rejects source larger than 64KB without overwriting the saved draft', async () => {
    localStorage.setItem(
      'draft:guest:two-sum-target:cpp17',
      JSON.stringify({ source_code: 'saved source', updated_at: '2026-07-13T14:40:00Z' }),
    );
    renderEditor();
    await screen.findByText('saved source');

    fireEvent.click(screen.getByRole('button', { name: '写入超限代码' }));

    expect(screen.getByText('代码不能超过 64KB。')).toBeInTheDocument();
    expect(JSON.parse(localStorage.getItem('draft:guest:two-sum-target:cpp17') ?? '{}')).toMatchObject({
      source_code: 'saved source',
    });
  });
});

function renderEditor(userID?: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <LanguageEditor problemSlug="two-sum-target" userID={userID} />
    </QueryClientProvider>,
  );
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function putCalls(fetchMock: ReturnType<typeof vi.fn<typeof fetch>>) {
  return fetchMock.mock.calls.filter(([, init]) => init?.method === 'PUT');
}
