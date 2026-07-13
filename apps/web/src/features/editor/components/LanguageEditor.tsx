import { AlertTriangle } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

import CodeEditor from '@/features/editor/components/CodeEditor';
import { useLanguages } from '@/features/problems/hooks/use-problems';
import type { Language } from '@/features/problems/lib/problems-api';

export default function LanguageEditor() {
  const languagesQuery = useLanguages();
  const initialized = useRef(false);
  const [languageKey, setLanguageKey] = useState('');
  const [sourceCode, setSourceCode] = useState('');
  const [template, setTemplate] = useState('');

  useEffect(() => {
    const firstLanguage = languagesQuery.data?.[0];
    if (!firstLanguage || initialized.current) return;

    initialized.current = true;
    setLanguageKey(firstLanguage.key);
    setSourceCode(firstLanguage.source_template);
    setTemplate(firstLanguage.source_template);
  }, [languagesQuery.data]);

  if (languagesQuery.status === 'pending') {
    return <EditorMessage>正在加载语言配置…</EditorMessage>;
  }

  if (languagesQuery.status === 'error') {
    return (
      <EditorMessage title="语言配置加载失败">
        <button className="secondary-button mt-4" onClick={() => void languagesQuery.refetch()}>
          重新加载
        </button>
      </EditorMessage>
    );
  }

  if (languagesQuery.data.length === 0) {
    return <EditorMessage title="暂无可用语言">请联系管理员启用判题语言。</EditorMessage>;
  }

  const firstLanguage = languagesQuery.data[0];
  if (!firstLanguage) return null;
  const selectedLanguage = findLanguage(languagesQuery.data, languageKey) ?? firstLanguage;

  const handleLanguageChange = (nextKey: string) => {
    if (nextKey === selectedLanguage.key) return;
    const nextLanguage = findLanguage(languagesQuery.data, nextKey);
    if (!nextLanguage) return;

    const hasModifiedSource = sourceCode !== template;
    if (hasModifiedSource && !window.confirm('切换语言会替换当前代码，是否继续？')) return;

    setLanguageKey(nextLanguage.key);
    setSourceCode(nextLanguage.source_template);
    setTemplate(nextLanguage.source_template);
  };

  return (
    <section className="flex min-h-[420px] flex-col border border-[var(--editor-border)] bg-[var(--editor-bg)]">
      <div className="flex items-center justify-between border-b border-[var(--editor-border)] bg-[var(--editor-panel)] px-4 py-3">
        <label className="flex items-center gap-3 text-sm text-neutral-300">
          <span>编程语言</span>
          <select
            aria-label="编程语言"
            className="h-9 border border-[var(--editor-border)] bg-[var(--editor-bg)] px-3 text-sm text-white"
            value={selectedLanguage.key}
            onChange={(event) => handleLanguageChange(event.target.value)}
          >
            {languagesQuery.data.map((language) => (
              <option key={language.key} value={language.key}>{language.display_name}</option>
            ))}
          </select>
        </label>
        {sourceCode !== template && <span className="text-xs text-[var(--warning)]">代码已修改</span>}
      </div>
      <div className="min-h-0 flex-1">
        <CodeEditor
          value={sourceCode}
          language={selectedLanguage.editor_language}
          onChange={setSourceCode}
          readOnly={false}
        />
      </div>
    </section>
  );
}

function findLanguage(languages: Language[], key: string) {
  return languages.find((language) => language.key === key);
}

function EditorMessage({ title, children }: { title?: string; children: React.ReactNode }) {
  return (
    <div className="flex min-h-[420px] flex-col items-center justify-center border border-[var(--editor-border)] bg-[var(--editor-bg)] px-6 text-center text-neutral-400">
      {title && (
        <>
          <AlertTriangle aria-hidden="true" className="mb-3 text-[var(--warning)]" size={24} />
          <h2 className="text-lg font-semibold text-white">{title}</h2>
        </>
      )}
      <div className="mt-1 text-sm">{children}</div>
    </div>
  );
}
