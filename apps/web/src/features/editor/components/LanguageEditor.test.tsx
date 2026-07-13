import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
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
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockResolvedValue(jsonResponse(languages)));
  });

  afterEach(() => {
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
});

function renderEditor() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <LanguageEditor />
    </QueryClientProvider>,
  );
}

function jsonResponse(payload: unknown) {
  return new Response(JSON.stringify(payload), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  });
}
