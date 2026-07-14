import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { authKeys, problemKeys, progressKeys, submissionKeys } from '@/lib/query-keys';
import * as authApi from '@/features/auth/lib/auth-api';
import type { LoginRequest, RegisterRequest } from '@/features/auth/lib/auth-api';

// --- useCurrentUser ---

export function useCurrentUser() {
  return useQuery({
    queryKey: authKeys.me,
    queryFn: authApi.getCurrentUser,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });
}

// --- useLogin ---

export function useLogin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: LoginRequest) => authApi.login(data),
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: submissionKeys.all });
      queryClient.removeQueries({ queryKey: progressKeys.all });
      queryClient.removeQueries({ queryKey: problemKeys.all });
      void queryClient.invalidateQueries({ queryKey: authKeys.me });
    },
  });
}

// --- useRegister ---

export function useRegister() {
  return useMutation({
    mutationFn: (data: RegisterRequest) => authApi.register(data),
  });
}

// --- useLogout ---

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => authApi.logout(),
    onSuccess: () => {
      queryClient.setQueryData(authKeys.me, null);
      queryClient.clear();
    },
  });
}
