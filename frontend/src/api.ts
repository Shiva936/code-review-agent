import { basicAuthHeader, loadConfig } from "./config";

export type RunStartResponse = {
  run_group_id: number;
  status: string;
};

export type SampleEvalMetrics = {
  index: number;
  total: number;
  actionability: number;
  specificity: number;
  severity: number;
  weakness_category: string;
  logic: number;
  performance: number;
  security: number;
  style: number;
};

export type RunGroupRun = {
  iteration: number;
  score: number;
  weakness: string;
  status: string;
  progress_percent?: number;
  actionability?: number;
  specificity?: number;
  severity?: number;
  structure?: number;
  samples?: SampleEvalMetrics[];
};

export type RunGroup = {
  id: number;
  input_code: string;
  status: string;
  created_at: string;
  updated_at: string;
  iterations: RunGroupRun[];
};

export type RunGroupsResponse = {
  total: number;
  page: number;
  page_size: number;
  groups: RunGroup[];
};

export type PromptVersion = {
  id: number;
  iteration: number;
  prompt_text: string;
  rules_json: string;
  source: string;
  reason: string;
  created_at: string;
};

export type PromptDelta = {
  id: number;
  iteration: number;
  weakest_issue: string;
  input_json: string;
  raw_output: string;
  delta_json: string;
  validation_status: string;
  applied: boolean;
  source: string;
  reason: string;
  created_at: string;
};

export type PromptArtifactsResponse = {
  run_group_id: number;
  versions: PromptVersion[];
  deltas: PromptDelta[];
};

async function request<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  const cfg = loadConfig();
  const base = (cfg.apiBaseUrl || "").trim();
  const sameOrigin = typeof window !== "undefined" && base === window.location.origin;
  const url = !base || sameOrigin ? path : base.replace(/\/+$/, "") + path;

  const headers = new Headers(init?.headers || undefined);
  headers.set("Content-Type", "application/json");

  const auth = basicAuthHeader(cfg.auth.username, cfg.auth.password);
  if (auth) headers.set("Authorization", auth);

  const res = await fetch(url, {
    ...init,
    headers,
  });

  if (!res.ok) {
    let detail = "";
    try {
      const t = await res.text();
      detail = t ? `: ${t}` : "";
    } catch {
      // ignore
    }
    throw new Error(`${init?.method || "GET"} ${path} failed (${res.status})${detail}`);
  }

  return (await res.json()) as T;
}

export const api = {
  run: (code: string, prompt: string) =>
    request<RunStartResponse>("/run", {
      method: "POST",
      body: JSON.stringify({ code, prompt }),
    }),
  getRunGroups: (page = 1) =>
    request<RunGroupsResponse>(`/run-groups?page=${page}`),
  getRunGroupPromptArtifacts: (groupId: number) =>
    request<PromptArtifactsResponse>(`/run-group-prompt-artifacts?group_id=${groupId}`),
};

