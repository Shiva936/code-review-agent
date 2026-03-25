import { basicAuthHeader, loadConfig } from "./config";

export type RunRow = {
  iteration: number;
  score: number;
  weakness: string;
};

export type RunsResponse = {
  runs: RunRow[];
};

export type LoopSummary = {
  iterations: number;
  sample_count: number;
  avg_scores: number[];
  weaknesses: string[];
  group_id: number;
};

export type RunGroupRun = {
  iteration: number;
  score: number;
  weakness: string;
};

export type RunGroup = {
  id: number;
  iterations: number;
  created_at: string;
  runs: RunGroupRun[];
};

export type RunGroupsResponse = {
  total: number;
  limit: number;
  offset: number;
  groups: RunGroup[];
};

async function request<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  const cfg = loadConfig();
  const url = cfg.apiBaseUrl.replace(/\/+$/, "") + path;

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
  getRuns: () => request<RunsResponse>("/runs"),
  run: (code: string, prompt: string) =>
    request<LoopSummary>("/run", {
      method: "POST",
      body: JSON.stringify({ code, prompt }),
    }),
  getRunGroups: (limit = 20, offset = 0) =>
    request<RunGroupsResponse>(`/run-groups?limit=${limit}&offset=${offset}`),
};

