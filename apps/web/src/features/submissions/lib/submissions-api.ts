import { api } from '@/lib/api-client';

// --- Types ---

export interface SubmissionSummary {
  id: string;
  problem_slug: string;
  problem_title: string;
  language: string;
  status: SubmissionStatus;
  time_ms: number | null;
  memory_kb: number | null;
  created_at: string;
}

export type SubmissionStatus =
  | 'pending'
  | 'running'
  | 'accepted'
  | 'wrong_answer'
  | 'time_limit_exceeded'
  | 'memory_limit_exceeded'
  | 'runtime_error'
  | 'compilation_error'
  | 'system_error';

export interface SubmissionListResponse {
  items: SubmissionSummary[];
  total: number;
  page: number;
  page_size: number;
}

export interface SubmissionListParams {
  problem_slug?: string;
  status?: string;
  language?: string;
  page?: number;
  page_size?: number;
}

export interface TestCaseResult {
  status: SubmissionStatus;
  time_ms: number;
  memory_kb: number;
}

export interface SubmissionDetail {
  id: string;
  problem_slug: string;
  problem_title: string;
  language: string;
  status: SubmissionStatus;
  time_ms: number | null;
  memory_kb: number | null;
  source_code: string;
  compiler_output: string | null;
  first_failure: {
    test_case: number;
    input: string;
    expected_output: string;
    actual_output: string;
  } | null;
  test_cases: TestCaseResult[];
  passed_count: number;
  total_count: number;
  created_at: string;
}

// --- API functions ---

export function getSubmissions(params?: SubmissionListParams): Promise<SubmissionListResponse> {
  return api.get('/submissions', params as Record<string, string | number | undefined>);
}

export function getSubmission(id: string): Promise<SubmissionDetail> {
  return api.get(`/submissions/${id}`);
}
