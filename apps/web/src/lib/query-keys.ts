// Centralized TanStack Query key definitions

export const authKeys = {
  all: ['auth'] as const,
  me: ['auth', 'me'] as const,
};

export const problemKeys = {
  all: ['problems'] as const,
  list: (params?: unknown) =>
    ['problems', params] as const,
  detail: (slug: string) => ['problems', slug] as const,
  navigation: (slug: string) => ['problems', slug, 'navigation'] as const,
};

export const tagKeys = {
  all: ['tags'] as const,
};

export const languageKeys = {
  all: ['languages'] as const,
};

export const draftKeys = {
  detail: (problemSlug: string, languageKey: string) =>
    ['drafts', problemSlug, languageKey] as const,
};

export const submissionKeys = {
  all: ['submissions'] as const,
  list: (params?: unknown) =>
    ['submissions', params] as const,
  detail: (id: string) => ['submissions', id] as const,
};

export const progressKeys = {
  all: ['progress'] as const,
  detail: (problemSlug: string) => ['progress', problemSlug] as const,
};
