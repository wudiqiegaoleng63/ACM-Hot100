import Editor, { loader, type BeforeMount } from '@monaco-editor/react';
import { AlertTriangle } from 'lucide-react';
import { useEffect, useState } from 'react';

const MIN_FONT_SIZE = 13;
const MAX_FONT_SIZE = 20;
const DEFAULT_FONT_SIZE = 14;
const EDITOR_THEME = 'acm-hot100-dark';

export interface CodeEditorProps {
  value: string;
  language: string;
  onChange: (value: string) => void;
  fontSize?: number;
  readOnly: boolean;
}

export default function CodeEditor({
  value,
  language,
  onChange,
  fontSize = DEFAULT_FONT_SIZE,
  readOnly,
}: CodeEditorProps) {
  const [loadFailed, setLoadFailed] = useState(false);
  const requestedFontSize = Number.isFinite(fontSize) ? fontSize : DEFAULT_FONT_SIZE;
  const safeFontSize = Math.min(MAX_FONT_SIZE, Math.max(MIN_FONT_SIZE, requestedFontSize));

  useEffect(() => {
    let active = true;
    loader.init().catch(() => {
      if (active) setLoadFailed(true);
    });
    return () => {
      active = false;
    };
  }, []);

  if (loadFailed) {
    return (
      <div
        role="alert"
        className="flex min-h-64 flex-col items-center justify-center border border-[var(--editor-border)] bg-[var(--editor-bg)] px-6 text-center text-neutral-300"
      >
        <AlertTriangle aria-hidden="true" className="text-[var(--warning)]" size={24} />
        <p className="mt-3 font-semibold text-white">代码编辑器加载失败</p>
        <p className="mt-1 text-sm text-neutral-400">请检查网络后刷新页面重试。</p>
      </div>
    );
  }

  return (
    <div className="min-h-64 overflow-hidden border border-[var(--editor-border)] bg-[var(--editor-bg)]">
      <Editor
        height="100%"
        width="100%"
        value={value}
        language={language}
        theme={EDITOR_THEME}
        beforeMount={defineEditorTheme}
        onChange={(nextValue) => onChange(nextValue ?? '')}
        loading={(
          <div className="flex min-h-64 items-center justify-center bg-[var(--editor-bg)] text-sm text-neutral-400">
            正在加载代码编辑器…
          </div>
        )}
        options={{
          automaticLayout: true,
          fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
          fontSize: safeFontSize,
          lineHeight: Math.round(safeFontSize * 1.6),
          minimap: { enabled: false },
          padding: { top: 16, bottom: 16 },
          readOnly,
          renderLineHighlight: 'line',
          scrollBeyondLastLine: false,
          tabSize: 4,
          wordWrap: 'off',
        }}
      />
    </div>
  );
}

const defineEditorTheme: BeforeMount = (monaco) => {
  monaco.editor.defineTheme(EDITOR_THEME, {
    base: 'vs-dark',
    inherit: true,
    rules: [],
    colors: {
      'editor.background': '#181a1b',
      'editorGutter.background': '#181a1b',
      'editorLineNumber.foreground': '#777a7d',
      'editorLineNumber.activeForeground': '#deddd7',
      'editor.lineHighlightBackground': '#202224',
      'editorCursor.foreground': '#c45724',
      'editor.selectionBackground': '#7c3d2255',
      'editorWidget.background': '#202224',
      'editorWidget.border': '#343638',
    },
  });
};
