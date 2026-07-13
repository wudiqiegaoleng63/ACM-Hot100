import { useQuery } from '@tanstack/react-query';
import { submissionKeys } from '@/lib/query-keys';
import * as submissionsApi from '@/features/submissions/lib/submissions-api';
import type { SubmissionListParams } from '@/features/submissions/lib/submissions-api';

// --- useSubmissions ---

export function useSubmissions(params?: SubmissionListParams) {
  return useQuery({
    queryKey: submissionKeys.list(params),
    queryFn: () => submissionsApi.getSubmissions(params),
    placeholderData: (prev) => prev,
  });
}

// --- useSubmission ---

export function useSubmission(id: string) {
  return useQuery({
    queryKey: submissionKeys.detail(id),
    queryFn: () => submissionsApi.getSubmission(id),
    enabled: !!id,
  });
}
