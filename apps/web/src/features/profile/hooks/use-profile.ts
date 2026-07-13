import { useQuery } from '@tanstack/react-query';
import { progressKeys } from '@/lib/query-keys';
import * as profileApi from '@/features/profile/lib/profile-api';

// --- useProgress ---

export function useProgress() {
  return useQuery({
    queryKey: progressKeys.all,
    queryFn: () => profileApi.getProgress(),
  });
}
