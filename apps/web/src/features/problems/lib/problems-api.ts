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

// --- Sample Run types and API ---

export const sampleRunStatusSchema = z.enum(['QUEUED', 'RUNNING', 'AC', 'SYSTEM_ERROR']);
export type SampleRunStatus = z.infer<typeof sampleRunStatusSchema>;

const sampleRunResponseSchema = z.object({
  id: z.string(),
  language_key: z.string(),
  sample_case_id: z.string().nullable(),
  input_data: z.string(),
  status: sampleRunStatusSchema,
  output_data: z.string(),
  error_message: z.string(),
  created_at: z.string().datetime({ offset: true }),
  updated_at: z.string().datetime({ offset: true }),
  started_at: z.string().datetime({ offset: true }).nullable(),
  finished_at: z.string().datetime({ offset: true }).nullable(),
  expires_at: z.string().datetime({ offset: true }),
});

export type SampleRunResponse = z.infer<typeof sampleRunResponseSchema>;

export interface CreateSampleRunParams {
  language_key: string;
  source_code: string;
  sample_case_id?: string;
  custom_input?: string;
}

export function isTerminalRunStatus(status: SampleRunStatus): boolean {
  return status === 'AC' || status === 'SYSTEM_ERROR';
}

export async function createSampleRun(slug: string, params: CreateSampleRunParams): Promise<SampleRunResponse> {
  return sampleRunResponseSchema.parse(
    await api.post<unknown>(`/problems/${slug}/run`, params),
  );
}

export async function getSampleRun(runID: string): Promise<SampleRunResponse> {
  return sampleRunResponseSchema.parse(await api.get<unknown>(`/runs/${runID}`));
}

// --- Health / Judge Mode ---

const healthResponseSchema = z.object({
  status: z.string(),
  services: z.record(z.string()),
  judge_mode: z.enum(['mock', 'judge0']),
});

export type HealthResponse = z.infer<typeof healthResponseSchema>;

export async function getHealth(): Promise<HealthResponse> {
  return healthResponseSchema.parse(await api.get<unknown>('/health'));
}

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

// --- Submission types and API ---

export const submissionStatusSchema = z.enum([
  'QUEUED', 'COMPILING', 'RUNNING',
  'AC', 'WA', 'TLE', 'MLE', 'RE', 'CE', 'SYSTEM_ERROR',
]);
export type SubmissionStatus = z.infer<typeof submissionStatusSchema>;

export function isTerminalSubmissionStatus(status: SubmissionStatus): boolean {
  return ['AC', 'WA', 'TLE', 'MLE', 'RE', 'CE', 'SYSTEM_ERROR'].includes(status);
}

const caseResultSchema = z.object({
  case_index: z.number().int(),
  status: submissionStatusSchema,
  time_ms: z.number().int().nullable(),
  memory_kb: z.number().int().nullable(),
  actual_output: z.string().optional(),
  is_sample: z.boolean(),
});

const submissionDetailSchema = z.object({
  id: z.string(),
  problem_slug: z.string(),
  language_key: z.string(),
  source_code: z.string(),
  status: submissionStatusSchema,
  passed_cases: z.number().int(),
  total_cases: z.number().int(),
  time_ms: z.number().int().nullable(),
  memory_kb: z.number().int().nullable(),
  compiler_output: z.string(),
  error_message: z.string(),
  case_results: z.array(caseResultSchema),
  created_at: z.string().datetime({ offset: true }),
  judged_at: z.string().datetime({ offset: true }).nullable(),
});

export type SubmissionDetail = z.infer<typeof submissionDetailSchema>;

const submissionSummarySchema = z.object({
  id: z.string(),
  problem_slug: z.string(),
  language_key: z.string(),
  status: submissionStatusSchema,
  passed_cases: z.number().int(),
  total_cases: z.number().int(),
  time_ms: z.number().int().nullable(),
  memory_kb: z.number().int().nullable(),
  created_at: z.string().datetime({ offset: true }),
});

export type SubmissionSummary = z.infer<typeof submissionSummarySchema>;

const submissionListResponseSchema = z.object({
  items: z.array(submissionSummarySchema),
  total: z.number().int().nonnegative(),
  page: z.number().int().positive(),
  page_size: z.number().int().positive(),
});

export type SubmissionListResponse = z.infer<typeof submissionListResponseSchema>;

export interface SubmissionListParams {
  problem?: string;
  status?: SubmissionStatus;
  language?: string;
  page?: number;
  page_size?: number;
}

const createSubmissionResponseSchema = z.object({
  id: z.string(),
  status: z.literal('QUEUED'),
  created_at: z.string().datetime({ offset: true }),
});

export type CreateSubmissionResponse = z.infer<typeof createSubmissionResponseSchema>;

export async function createSubmission(slug: string, languageKey: string, sourceCode: string): Promise<CreateSubmissionResponse> {
  return createSubmissionResponseSchema.parse(
    await api.post<unknown>(`/problems/${slug}/submissions`, {
      language_key: languageKey,
      source_code: sourceCode,
    }),
  );
}

export async function getSubmission(id: string): Promise<SubmissionDetail> {
  return submissionDetailSchema.parse(await api.get<unknown>(`/submissions/${id}`));
}

export async function listSubmissions(params?: SubmissionListParams): Promise<SubmissionListResponse> {
  return submissionListResponseSchema.parse(
    await api.get<unknown>('/submissions', params as Record<string, string | number | undefined>),
  );
}
