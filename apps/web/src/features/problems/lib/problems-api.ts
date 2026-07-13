import { api } from '@/lib/api-client';

// --- Types ---

export interface Tag {
  key: string;
  name: string;
}

export interface ProblemSummary {
  slug: string;
  title: string;
  difficulty: 'easy' | 'medium' | 'hard';
  tags: Tag[];
  state: 'not_started' | 'attempting' | 'accepted';
  order: number;
  last_submitted_at: string | null;
}

export interface ProblemListResponse {
  items: ProblemSummary[];
  total: number;
  page: number;
  page_size: number;
}

export interface ProblemNavigation {
  prev_slug: string | null;
  next_slug: string | null;
}

export interface SampleCase {
  input: string;
  output: string;
  explanation?: string;
}

export interface ProblemDetail {
  slug: string;
  title: string;
  difficulty: 'easy' | 'medium' | 'hard';
  tags: Tag[];
  state: 'not_started' | 'attempting' | 'accepted';
  description: string;
  input_format: string;
  output_format: string;
  constraints: string;
  hints: string;
  sample_cases: SampleCase[];
  time_limit_ms: number;
  memory_limit_mb: number;
}

export interface ProblemListParams {
  q?: string;
  difficulty?: string;
  tag?: string;
  state?: string;
  page?: number;
  page_size?: number;
}

export interface DraftData {
  source_code: string;
}

// --- API functions ---

export function getProblems(params?: ProblemListParams): Promise<ProblemListResponse> {
  return api.get('/problems', params as Record<string, string | number | undefined>);
}

export function getProblem(slug: string): Promise<ProblemDetail> {
  return api.get(`/problems/${slug}`);
}

export function getProblemNavigation(slug: string): Promise<ProblemNavigation> {
  return api.get(`/problems/${slug}/navigation`);
}

export function getTags(): Promise<Tag[]> {
  return api.get('/tags');
}

export function saveDraft(slug: string, languageKey: string, sourceCode: string): Promise<void> {
  return api.put(`/problems/${slug}/drafts/${languageKey}`, { source_code: sourceCode });
}

export function getDraft(slug: string, languageKey: string): Promise<DraftData> {
  return api.get(`/problems/${slug}/drafts/${languageKey}`);
}
