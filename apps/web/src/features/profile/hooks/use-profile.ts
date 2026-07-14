import { useQuery } from '@tanstack/react-query';

import * as profileApi from '@/features/profile/lib/profile-api';
import { progressKeys } from '@/lib/query-keys';

export function useProgressSummary() {
  return useQuery({
    queryKey: progressKeys.summary,
    queryFn: profileApi.getProgressSummary,
  });
}

export function useProgressByStage() {
  return useQuery({
    queryKey: progressKeys.byStage,
    queryFn: profileApi.getProgressByStage,
  });
}
