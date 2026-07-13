import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import CodeEditor from './CodeEditor';

const { editorSpy, loaderInitSpy } = vi.hoisted(() => ({
  editorSpy: vi.fn(),
  loaderInitSpy: vi.fn<() => Promise<unknown>>(() => Promise.resolve({})),
}));

vi.mock('@monaco-editor/react', () => ({
  loader: { init: loaderInitSpy },
  default: (props: {
    value: string;
    language: string;
    onChange: (value: string | undefined) => void;
    loading: React.ReactNode;
    options: { fontSize: number; lineHeight: number; readOnly: boolean };
    beforeMount: (monaco: {
      editor: { defineTheme: (name: string, theme: unknown) => void };
    }) => void;
  }) => {
    editorSpy(props);
    props.beforeMount({ editor: { defineTheme: vi.fn() } });
    return (
      <div data-testid="monaco-editor">
        {props.loading}
        <button onClick={() => props.onChange('updated source')}>模拟输入</button>
      </div>
    );
  },
}));

describe('CodeEditor', () => {
  beforeEach(() => {
    editorSpy.mockClear();
    loaderInitSpy.mockReset();
    loaderInitSpy.mockResolvedValue({});
  });

  it('passes controlled content and bounded editor options to Monaco', () => {
    const onChange = vi.fn();
    render(
      <CodeEditor
        value="int main() {}"
        language="cpp"
        onChange={onChange}
        fontSize={99}
        readOnly
      />,
    );

    expect(screen.getByText('正在加载代码编辑器…')).toBeInTheDocument();
    expect(editorSpy).toHaveBeenCalledWith(expect.objectContaining({
      value: 'int main() {}',
      language: 'cpp',
      theme: 'acm-hot100-dark',
      options: expect.objectContaining({
        fontSize: 20,
        lineHeight: 32,
        readOnly: true,
      }),
    }));

    fireEvent.click(screen.getByRole('button', { name: '模拟输入' }));
    expect(onChange).toHaveBeenCalledWith('updated source');
  });

  it('uses default font size and normalizes an undefined Monaco value', () => {
    const onChange = vi.fn();
    render(
      <CodeEditor
        value=""
        language="python"
        onChange={onChange}
        readOnly={false}
      />,
    );

    expect(editorSpy).toHaveBeenCalledWith(expect.objectContaining({
      options: expect.objectContaining({ fontSize: 14, lineHeight: 22 }),
    }));

    const props = editorSpy.mock.calls[0]?.[0] as { onChange: (value: string | undefined) => void };
    props.onChange(undefined);
    expect(onChange).toHaveBeenCalledWith('');
  });

  it('replaces the editor with a visible error state after a load failure', async () => {
    loaderInitSpy.mockRejectedValue(new Error('loader unavailable'));
    render(
      <CodeEditor
        value=""
        language="java"
        onChange={vi.fn()}
        readOnly={false}
      />,
    );

    expect(await screen.findByRole('alert')).toHaveTextContent('代码编辑器加载失败');
    expect(screen.queryByTestId('monaco-editor')).not.toBeInTheDocument();
  });
});
