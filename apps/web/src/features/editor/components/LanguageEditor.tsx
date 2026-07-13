import { AlertTriangle } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';

import CodeEditor from '@/features/editor/components/CodeEditor';
import { useDraft, useLanguages, useSaveDraft } from '@/features/problems/hooks/use-problems';
import type { DraftData, Language } from '@/features/problems/lib/problems-api';
import { ApiError } from '@/lib/api-client';

const MAX_SOURCE_BYTES = 64 * 1024;
const SAVE_DELAY_MS = 500;

interface LocalDraft {
  source_code: string;
  updated_at: string;
}

interface PendingSave {
  timer: ReturnType<typeof setTimeout>;
  userID: string;
  slug: string;
  languageKey: string;
  sourceCode: string;
}

export default function LanguageEditor({
  problemSlug,
  userID,
  onStateChange,
}: {
  problemSlug: string;
  userID?: string;
  onStateChange?: (state: { languageKey: string; sourceCode: string }) => void;
}) {
  const languagesQuery = useLanguages();
  const saveDraft = useSaveDraft();
  const saveDraftRef = useRef(saveDraft.mutateAsync);
  const pendingSave = useRef<PendingSave | null>(null);
  const [languageKey, setLanguageKey] = useState('');
  const [sourceCode, setSourceCode] = useState('');
  const [template, setTemplate] = useState('');
  const [restoredContext, setRestoredContext] = useState('');
  const [warning, setWarning] = useState('');
  const [sizeError, setSizeError] = useState('');
  const draftQuery = useDraft(userID, problemSlug, languageKey);
  const owner = userID ?? 'guest';
  const currentContext = draftContext(owner, problemSlug, languageKey);

  saveDraftRef.current = saveDraft.mutateAsync;

  // Notify parent of current language and source code
  const onStateChangeRef = useRef(onStateChange);
  onStateChangeRef.current = onStateChange;
  useEffect(() => {
    if (languageKey && restoredContext === currentContext) {
      onStateChangeRef.current?.({ languageKey, sourceCode });
    }
  }, [languageKey, sourceCode, currentContext, restoredContext]);

  const flushPendingSave = useCallback(() => {
    const pending = pendingSave.current;
    if (!pending) return;
    clearTimeout(pending.timer);
    pendingSave.current = null;
    void saveDraftRef.current(pending).catch(() => undefined);
  }, []);

  useEffect(() => {
    const firstLanguage = languagesQuery.data?.[0];
    if (!firstLanguage || languageKey) return;
    setLanguageKey(firstLanguage.key);
  }, [languageKey, languagesQuery.data]);

  useEffect(() => {
    if (!languageKey) return;
    if (userID && draftQuery.status === 'pending') return;
    if (restoredContext === currentContext) return;

    const language = findLanguage(languagesQuery.data ?? [], languageKey);
    if (!language) return;

    const localDraft = readLocalDraft(localDraftKey(owner, problemSlug, languageKey));
    const serverDraft = draftQuery.status === 'success' ? draftQuery.data : undefined;
    const restored = newerDraft(localDraft, serverDraft) ?? {
      source_code: language.source_template,
      updated_at: '',
    };

    setTemplate(language.source_template);
    setSourceCode(restored.source_code);
    setRestoredContext(currentContext);
    setSizeError('');

    if (serverDraft && restored === serverDraft) {
      writeLocalDraft(localDraftKey(owner, problemSlug, languageKey), serverDraft);
    }
    if (userID && localDraft && restored === localDraft && newerThan(localDraft, serverDraft)) {
      scheduleServerSave(userID, problemSlug, languageKey, localDraft.source_code);
    }
    if (userID && draftQuery.status === 'error' && !isMissingDraft(draftQuery.error)) {
      setWarning('服务器草稿加载失败，已使用本地草稿。');
    }
  }, [
    currentContext,
    draftQuery.data,
    draftQuery.error,
    draftQuery.status,
    languageKey,
    languagesQuery.data,
    owner,
    problemSlug,
    restoredContext,
    userID,
  ]);

  useEffect(() => flushPendingSave, [flushPendingSave, problemSlug, userID]);

  const scheduleServerSave = (
    saveUserID: string,
    slug: string,
    saveLanguageKey: string,
    nextSourceCode: string,
  ) => {
    if (pendingSave.current) clearTimeout(pendingSave.current.timer);
    const payload = {
      userID: saveUserID,
      slug,
      languageKey: saveLanguageKey,
      sourceCode: nextSourceCode,
    };
    const timer = setTimeout(() => {
      pendingSave.current = null;
      void saveDraftRef.current(payload).catch(() => {
        if (draftContext(saveUserID, slug, saveLanguageKey) === currentContext) {
          setWarning('服务器草稿保存失败，本地草稿已保留。');
        }
      });
    }, SAVE_DELAY_MS);
    pendingSave.current = { timer, ...payload };
  };

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

  if (!languageKey || restoredContext !== currentContext) {
    return <EditorMessage>正在恢复草稿…</EditorMessage>;
  }

  const handleLanguageChange = (nextKey: string) => {
    if (nextKey === selectedLanguage.key) return;
    const nextLanguage = findLanguage(languagesQuery.data, nextKey);
    if (!nextLanguage) return;

    if (sourceCode !== template && !window.confirm('切换语言会替换当前代码，是否继续？')) return;

    flushPendingSave();
    setLanguageKey(nextLanguage.key);
    setTemplate(nextLanguage.source_template);
    setSourceCode(nextLanguage.source_template);
    setRestoredContext('');
    setWarning('');
    setSizeError('');
  };

  const handleSourceChange = (nextSourceCode: string) => {
    if (new TextEncoder().encode(nextSourceCode).byteLength > MAX_SOURCE_BYTES) {
      setSizeError('代码不能超过 64KB。');
      return;
    }

    const localDraft: LocalDraft = {
      source_code: nextSourceCode,
      updated_at: new Date().toISOString(),
    };
    setSourceCode(nextSourceCode);
    setSizeError('');
    setWarning('');
    if (!writeLocalDraft(localDraftKey(owner, problemSlug, selectedLanguage.key), localDraft)) {
      setWarning('本地草稿保存失败，请复制代码后刷新页面。');
    }
    if (userID) scheduleServerSave(userID, problemSlug, selectedLanguage.key, nextSourceCode);
  };

  return (
    <section className="flex min-h-[420px] flex-col border border-[var(--editor-border)] bg-[var(--editor-bg)]">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-[var(--editor-border)] bg-[var(--editor-panel)] px-4 py-3">
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
        <div className="text-right text-xs">
          {sizeError && <p className="text-[var(--danger)]">{sizeError}</p>}
          {warning && <p className="text-[var(--warning)]">{warning}</p>}
          {!sizeError && !warning && sourceCode !== template && <p className="text-[var(--warning)]">代码已修改</p>}
        </div>
      </div>
      <div className="min-h-0 flex-1">
        <CodeEditor
          value={sourceCode}
          language={selectedLanguage.editor_language}
          onChange={handleSourceChange}
          readOnly={false}
        />
      </div>
    </section>
  );
}

function findLanguage(languages: Language[], key: string) {
  return languages.find((language) => language.key === key);
}

function localDraftKey(owner: string, problemSlug: string, languageKey: string) {
  return `draft:${owner}:${problemSlug}:${languageKey}`;
}

function draftContext(owner: string, problemSlug: string, languageKey: string) {
  return `${owner}:${problemSlug}:${languageKey}`;
}

function readLocalDraft(key: string): LocalDraft | undefined {
  try {
    const stored = localStorage.getItem(key);
    if (!stored) return undefined;
    const draft: unknown = JSON.parse(stored);
    if (!isLocalDraft(draft)) {
      localStorage.removeItem(key);
      return undefined;
    }
    return draft;
  } catch {
    return undefined;
  }
}

function writeLocalDraft(key: string, draft: LocalDraft): boolean {
  try {
    localStorage.setItem(key, JSON.stringify(draft));
    return true;
  } catch {
    return false;
  }
}

function isLocalDraft(value: unknown): value is LocalDraft {
  if (!value || typeof value !== 'object') return false;
  const candidate = value as Record<string, unknown>;
  return typeof candidate.source_code === 'string'
    && typeof candidate.updated_at === 'string'
    && Number.isFinite(Date.parse(candidate.updated_at));
}

function newerDraft(localDraft?: LocalDraft, serverDraft?: DraftData): LocalDraft | DraftData | undefined {
  if (!localDraft) return serverDraft;
  if (!serverDraft) return localDraft;
  return newerThan(localDraft, serverDraft) ? localDraft : serverDraft;
}

function newerThan(left: LocalDraft, right?: DraftData) {
  return !right || Date.parse(left.updated_at) > Date.parse(right.updated_at);
}

function isMissingDraft(error: unknown) {
  return error instanceof ApiError && error.code === 'NOT_FOUND';
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
