import { z } from 'zod';

import { api } from '@/lib/api-client';

const progressSummarySchema = z.object({
  total_problems: z.number().int().nonnegative(),
  solved: z.number().int().nonnegative(),
  attempted: z.number().int().nonnegative(),
  not_started: z.number().int().nonnegative(),
});

const stageProgressSchema = z.object({
  stage: z.string(),
  total: z.number().int().nonnegative(),
  solved: z.number().int().nonnegative(),
  attempted: z.number().int().nonnegative(),
  not_started: z.number().int().nonnegative(),
});

export type ProgressSummary = z.infer<typeof progressSummarySchema>;
export type StageProgress = z.infer<typeof stageProgressSchema>;

export async function getProgressSummary(): Promise<ProgressSummary> {
  return progressSummarySchema.parse(await api.get<unknown>('/profile/summary'));
}

export async function getProgressByStage(): Promise<StageProgress[]> {
  return z.array(stageProgressSchema).parse(await api.get<unknown>('/profile/progress-by-stage'));
}
