import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { problemKeys, tagKeys, languageKeys, draftKeys } from '@/lib/query-keys';
import * as problemsApi from '@/features/problems/lib/problems-api';
import type { ProblemListParams } from '@/features/problems/lib/problems-api';

// --- useProblems ---

export function useProblems(params?: ProblemListParams) {
  return useQuery({
    queryKey: problemKeys.list(params),
    queryFn: () => problemsApi.getProblems(params),
    placeholderData: (prev) => prev,
  });
}

// --- useProblem ---

export function useProblem(slug: string) {
  return useQuery({
    queryKey: problemKeys.detail(slug),
    queryFn: () => problemsApi.getProblem(slug),
    enabled: !!slug,
  });
}

// --- useProblemNavigation ---

export function useProblemNavigation(slug: string) {
  return useQuery({
    queryKey: problemKeys.navigation(slug),
    queryFn: () => problemsApi.getProblemNavigation(slug),
    enabled: !!slug,
  });
}

// --- useTags ---

export function useTags() {
  return useQuery({
    queryKey: tagKeys.all,
    queryFn: () => problemsApi.getTags(),
    staleTime: 30 * 60 * 1000,
  });
}

// --- useLanguages ---

export function useLanguages() {
  return useQuery({
    queryKey: languageKeys.all,
    queryFn: () => problemsApi.getLanguages(),
    staleTime: 30 * 60 * 1000,
  });
}

// --- useSaveDraft (debounce handled at call site) ---

export function useSaveDraft() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ slug, languageKey, sourceCode }: { userID: string; slug: string; languageKey: string; sourceCode: string }) =>
      problemsApi.saveDraft(slug, languageKey, sourceCode),
    onSuccess: (draft, variables) => {
      queryClient.setQueryData(
        draftKeys.detail(variables.userID, variables.slug, variables.languageKey),
        draft,
      );
    },
  });
}

// --- useDraft ---

export function useDraft(userID: string | undefined, slug: string, languageKey: string) {
  return useQuery({
    queryKey: draftKeys.detail(userID ?? 'guest', slug, languageKey),
    queryFn: () => problemsApi.getDraft(slug, languageKey),
    enabled: Boolean(userID && slug && languageKey),
    retry: false,
  });
}
