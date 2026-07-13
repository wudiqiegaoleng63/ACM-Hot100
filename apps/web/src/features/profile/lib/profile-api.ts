import { api } from '@/lib/api-client';

// --- Types ---

export interface ProgressOverview {
  total_problems: number;
  accepted: number;
  attempting: number;
  not_started: number;
  by_stage: StageProgress[];
}

export interface StageProgress {
  stage: string;
  total: number;
  accepted: number;
  attempting: number;
  not_started: number;
}

// --- API functions ---

export function getProgress(): Promise<ProgressOverview> {
  return api.get('/progress');
}
