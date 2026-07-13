import { z } from 'zod';

import { api } from '@/lib/api-client';

export const difficultySchema = z.enum(['EASY', 'MEDIUM', 'HARD']);
export const progressStateSchema = z.enum(['NOT_STARTED', 'ATTEMPTED', 'SOLVED']);

const tagSchema = z.object({
  slug: z.string(),
  name: z.string(),
});

const problemSummarySchema = z.object({
  id: z.string(),
  slug: z.string(),
  order_index: z.number().int(),
  title: z.string(),
  difficulty: difficultySchema,
  tags: z.array(tagSchema),
  progress_state: progressStateSchema.nullable(),
});

const problemListResponseSchema = z.object({
  items: z.array(problemSummarySchema),
  total: z.number().int().nonnegative(),
  page: z.number().int().positive(),
  page_size: z.number().int().positive(),
});

const sampleCaseSchema = z.object({
  id: z.string(),
  order_index: z.number().int(),
  input_data: z.string(),
  expected_output: z.string(),
  explanation_md: z.string(),
});

const problemDetailSchema = z.object({
  id: z.string(),
  slug: z.string(),
  order_index: z.number().int(),
  title: z.string(),
  difficulty: difficultySchema,
  stage: z.string(),
  tags: z.array(tagSchema),
  progress_state: progressStateSchema.nullable(),
  statement_md: z.string(),
  input_format_md: z.string(),
  output_format_md: z.string(),
  constraints_md: z.string(),
  hints_md: z.string(),
  time_limit_ms: z.number().int().positive(),
  memory_limit_kb: z.number().int().positive(),
  sample_cases: z.array(sampleCaseSchema),
});

const navigationItemSchema = z.object({
  slug: z.string(),
  title: z.string(),
});

const problemNavigationSchema = z.object({
  prev: navigationItemSchema.nullable(),
  next: navigationItemSchema.nullable(),
});

const languageSchema = z.object({
  key: z.string(),
  display_name: z.string(),
  editor_language: z.string(),
  source_template: z.string(),
});

const languagesSchema = z.array(languageSchema);
const tagsSchema = z.array(tagSchema);
const draftSchema = z.object({
  source_code: z.string(),
  language_key: z.string(),
  updated_at: z.string().datetime({ offset: true }),
});

export type Difficulty = z.infer<typeof difficultySchema>;
export type ProgressState = z.infer<typeof progressStateSchema>;
export type Tag = z.infer<typeof tagSchema>;
export type ProblemSummary = z.infer<typeof problemSummarySchema>;
export type ProblemListResponse = z.infer<typeof problemListResponseSchema>;
export type SampleCase = z.infer<typeof sampleCaseSchema>;
export type ProblemDetail = z.infer<typeof problemDetailSchema>;
export type ProblemNavigation = z.infer<typeof problemNavigationSchema>;
export type Language = z.infer<typeof languageSchema>;

export interface ProblemListParams {
  q?: string;
  difficulty?: Difficulty;
  tag?: string;
  state?: ProgressState;
  page?: number;
  page_size?: number;
}

export type DraftData = z.infer<typeof draftSchema>;

export async function getProblems(params?: ProblemListParams): Promise<ProblemListResponse> {
  const response = await api.get<unknown>(
    '/problems',
    params as Record<string, string | number | undefined>,
  );
  return problemListResponseSchema.parse(response);
}

export async function getProblem(slug: string): Promise<ProblemDetail> {
  return problemDetailSchema.parse(await api.get<unknown>(`/problems/${slug}`));
}

export async function getProblemNavigation(slug: string): Promise<ProblemNavigation> {
  return problemNavigationSchema.parse(
    await api.get<unknown>(`/problems/${slug}/navigation`),
  );
}

export async function getTags(): Promise<Tag[]> {
  return tagsSchema.parse(await api.get<unknown>('/tags'));
}

export async function getLanguages(): Promise<Language[]> {
  return languagesSchema.parse(await api.get<unknown>('/languages'));
}

export async function saveDraft(slug: string, languageKey: string, sourceCode: string): Promise<DraftData> {
  return draftSchema.parse(
    await api.put<unknown>(`/problems/${slug}/drafts/${languageKey}`, { source_code: sourceCode }),
  );
}

export async function getDraft(slug: string, languageKey: string): Promise<DraftData> {
  return draftSchema.parse(await api.get<unknown>(`/problems/${slug}/drafts/${languageKey}`));
}
