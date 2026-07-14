import { useQuery } from '@tanstack/react-query';
import { submissionKeys } from '@/lib/query-keys';
import * as submissionsApi from '@/features/submissions/lib/submissions-api';
import type { SubmissionListParams } from '@/features/submissions/lib/submissions-api';

// --- useSubmissions ---

export function useSubmissions(params?: SubmissionListParams, userID = 'anonymous') {
  return useQuery({
    queryKey: submissionKeys.list(userID, params),
    queryFn: () => submissionsApi.getSubmissions(params),
    placeholderData: (prev) => prev,
  });
}

// --- useSubmission ---

export function useSubmission(id: string, userID = 'anonymous') {
  return useQuery({
    queryKey: submissionKeys.detail(userID, id),
    queryFn: () => submissionsApi.getSubmission(id),
    enabled: !!id,
  });
}
