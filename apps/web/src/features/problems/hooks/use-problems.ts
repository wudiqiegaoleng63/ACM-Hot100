import { useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { problemKeys, tagKeys, languageKeys, draftKeys, runKeys, healthKeys } from '@/lib/query-keys';
import * as problemsApi from '@/features/problems/lib/problems-api';
import type { ProblemListParams, CreateSampleRunParams } from '@/features/problems/lib/problems-api';
import { isTerminalRunStatus } from '@/features/problems/lib/problems-api';

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

// --- useCreateSampleRun ---

export function useCreateSampleRun() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ slug, params }: { slug: string; params: CreateSampleRunParams }) =>
      problemsApi.createSampleRun(slug, params),
    onSuccess: (run) => {
      queryClient.setQueryData(runKeys.detail(run.id), run);
    },
  });
}

// --- useSampleRun (polling) ---

const FAST_POLL_MS = 800;
const SLOW_POLL_MS = 2000;
const FAST_POLL_DURATION_MS = 10_000;

export function useSampleRun(runID: string | null) {
  const createdAt = useRef<number | null>(null);

  return useQuery({
    queryKey: runKeys.detail(runID ?? ''),
    queryFn: () => problemsApi.getSampleRun(runID!),
    enabled: Boolean(runID),
    refetchInterval: (query) => {
      if (!query.state.data) return FAST_POLL_MS;
      if (isTerminalRunStatus(query.state.data.status)) return false;
      if (createdAt.current === null) createdAt.current = Date.now();
      const elapsed = Date.now() - createdAt.current;
      return elapsed < FAST_POLL_DURATION_MS ? FAST_POLL_MS : SLOW_POLL_MS;
    },
    refetchIntervalInBackground: false,
  });
}

// --- useHealth (judge mode) ---

export function useHealth() {
  return useQuery({
    queryKey: healthKeys.all,
    queryFn: () => problemsApi.getHealth(),
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}
