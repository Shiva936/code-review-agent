import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import Chart from "chart.js/auto";
import {
  api,
  type PromptArtifactsResponse,
  type RunGroup,
  type RunGroupRun,
  type RunStartResponse,
} from "./api";

const POLL_MS = 2500;

function truncate(s: string, max: number) {
  if (!s) return "";
  const t = s.replace(/\s+/g, " ").trim();
  return t.length <= max ? t : t.slice(0, max) + "…";
}

/** Last completed iteration's score, or undefined if none */
function finalScoreForGroup(g: RunGroup): number | undefined {
  const done = (g.iterations || []).filter((r) => r.status === "completed");
  if (done.length === 0) return undefined;
  const last = done.reduce((a, b) => (a.iteration > b.iteration ? a : b));
  return last.score;
}

/** Rubric 1–5 → mockup-style labels (numeric still available via title). */
const METRIC_LABELS = {
  specificity: ["Very low", "Low", "Medium", "High", "Very high"],
  actionability: ["Very weak", "Weak", "Moderate", "Strong", "Very strong"],
  severity: ["Poor", "Fair", "Good", "Strong", "Excellent"],
  structure: ["Poor", "Fair", "Good", "Strong", "Excellent"],
} as const;

type MetricKind = keyof typeof METRIC_LABELS;

function fmtMetricQual(v: number | undefined, kind: MetricKind): string {
  if (v === undefined || v === null || Number.isNaN(v)) return "—";
  const n = Math.round(Number(v));
  const labels = METRIC_LABELS[kind];
  if (n >= 1 && n <= 5) return labels[n - 1];
  return String(v);
}

function metricTitle(v: number | undefined, kind: MetricKind): string | undefined {
  if (v === undefined || v === null || Number.isNaN(v)) return undefined;
  return `Rubric: ${Math.round(Number(v))}/5`;
}

function sortedIterations(iterations: RunGroupRun[] | undefined): RunGroupRun[] {
  return [...(iterations || [])].sort((a, b) => a.iteration - b.iteration);
}

/** Previous completed iteration in sort order (for score trend). */
function lastCompletedBefore(sorted: RunGroupRun[], currentIndex: number): RunGroupRun | undefined {
  for (let i = currentIndex - 1; i >= 0; i--) {
    if (sorted[i].status === "completed") return sorted[i];
  }
  return undefined;
}

/** Bar + caption while a run group is executing (polling updates this). */
function runGroupProgressUi(g: RunGroup): { barPct: number; caption: string } {
  const rows = sortedIterations(g.iterations);
  const n = Math.max(1, rows.length);
  const completed = rows.filter((r) => r.status === "completed").length;
  const running = rows.find((r) => r.status === "running");
  let barPct = Math.round((completed / n) * 100);
  if (running) {
    barPct = Math.min(99, Math.round(((completed + 0.5) / n) * 100));
  }
  if (g.status === "completed") {
    barPct = 100;
  }
  let caption: string;
  if (g.status === "completed") {
    caption = `All ${n} iterations finished`;
  } else if (running) {
    caption = `Running iteration ${running.iteration} of ${n} · ${completed} completed so far`;
  } else if (g.status === "running" || g.status === "pending") {
    caption = `In progress · ${completed}/${n} iterations complete`;
  } else {
    caption = `${completed}/${n} iterations complete`;
  }
  return { barPct, caption };
}

export default function App() {
  const [code, setCode] = useState("package main\n\nfunc main() {}\n");
  const [prompt, setPrompt] = useState("");
  const [runLoading, setRunLoading] = useState(false);
  const [runSummary, setRunSummary] = useState<RunStartResponse | null>(null);

  /** Newest run group (always from API page 1). */
  const [recentRunGroup, setRecentRunGroup] = useState<RunGroup | null>(null);
  /** Paginated list for “All Evaluations”. */
  const [listRunGroups, setListRunGroups] = useState<RunGroup[]>([]);
  const [listTotal, setListTotal] = useState(0);
  const [listPageSize, setListPageSize] = useState(5);
  const [evalPage, setEvalPage] = useState(1);

  const [groupsLoading, setGroupsLoading] = useState(true);
  const [groupsError, setGroupsError] = useState<string | null>(null);
  const [selectedRunGroupId, setSelectedRunGroupId] = useState<number | null>(null);
  const [promptArtifactsByGroup, setPromptArtifactsByGroup] = useState<Record<number, PromptArtifactsResponse>>({});
  const [promptArtifactsLoading, setPromptArtifactsLoading] = useState<Record<number, boolean>>({});

  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const chartRef = useRef<Chart | null>(null);

  const refreshGroups = useCallback(async (silent = false, listPageOverride?: number) => {
    const listPage = listPageOverride ?? evalPage;
    if (!silent) {
      setGroupsLoading(true);
      setGroupsError(null);
    }
    try {
      const [firstPage, listResp] = await Promise.all([api.getRunGroups(1), api.getRunGroups(listPage)]);
      setRecentRunGroup(firstPage.groups?.[0] ?? null);
      setListRunGroups(listResp.groups || []);
      setListTotal(listResp.total);
      setListPageSize(listResp.page_size || 5);
    } catch (e) {
      setGroupsError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      if (!silent) {
        setGroupsLoading(false);
      }
    }
  }, [evalPage]);

  useEffect(() => {
    void refreshGroups(false);
  }, [refreshGroups]);

  useEffect(() => {
    if (listTotal === 0) return;
    const max = Math.max(1, Math.ceil(listTotal / Math.max(1, listPageSize)));
    if (evalPage > max) setEvalPage(max);
  }, [listTotal, listPageSize, evalPage]);

  useEffect(() => {
    let stopped = false;
    const t = window.setInterval(() => {
      if (stopped) return;
      void refreshGroups(true);
    }, POLL_MS);
    return () => {
      stopped = true;
      window.clearInterval(t);
    };
  }, [refreshGroups]);

  const chartLabels = useMemo(() => [1, 2, 3, 4, 5], []);
  const chartScores = useMemo(() => {
    if (!recentRunGroup?.iterations?.length) return [null, null, null, null, null] as (number | null)[];
    const out: (number | null)[] = [null, null, null, null, null];
    for (const it of recentRunGroup.iterations) {
      const i = it.iteration;
      if (i >= 1 && i <= 5 && it.status === "completed") {
        out[i - 1] = it.score;
      }
    }
    return out;
  }, [recentRunGroup]);

  const totalListPages = Math.max(1, Math.ceil(listTotal / Math.max(1, listPageSize)));

  useEffect(() => {
    if (!canvasRef.current) return;
    const ctx = canvasRef.current.getContext("2d");
    if (!ctx) return;

    if (chartRef.current) {
      chartRef.current.destroy();
      chartRef.current = null;
    }

    const data = chartScores.map((v) => (v === null ? null : v)) as (number | null)[];

    chartRef.current = new Chart(ctx, {
      type: "line",
      data: {
        labels: chartLabels.map(String),
        datasets: [
          {
            label: "Score",
            data,
            borderColor: "rgba(14, 99, 156, 1)",
            backgroundColor: "rgba(14, 99, 156, 0.12)",
            borderWidth: 2,
            pointRadius: 4,
            tension: 0.2,
            spanGaps: false,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: true },
        },
        scales: {
          x: { title: { display: true, text: "Iteration" } },
          y: {
            beginAtZero: true,
            suggestedMax: 15,
            title: { display: true, text: "Score" },
          },
        },
      },
    });

    return () => {
      if (chartRef.current) {
        chartRef.current.destroy();
        chartRef.current = null;
      }
    };
  }, [chartLabels, chartScores]);

  async function onSubmit() {
    setRunLoading(true);
    setRunSummary(null);
    setGroupsError(null);
    try {
      const summary = await api.run(code, prompt);
      setRunSummary(summary);
      setEvalPage(1);
      await refreshGroups(true, 1);
    } catch (e) {
      setGroupsError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      setRunLoading(false);
    }
  }

  async function toggleExpand(id: number) {
    setSelectedRunGroupId((prev) => (prev === id ? null : id));
    if (!promptArtifactsByGroup[id] && !promptArtifactsLoading[id]) {
      setPromptArtifactsLoading((prev) => ({ ...prev, [id]: true }));
      try {
        const artifacts = await api.getRunGroupPromptArtifacts(id);
        setPromptArtifactsByGroup((prev) => ({ ...prev, [id]: artifacts }));
      } catch {
        // keep panel usable even if artifacts fail to load
      } finally {
        setPromptArtifactsLoading((prev) => ({ ...prev, [id]: false }));
      }
    }
  }

  function renderIterationRows(iterations: RunGroupRun[] | undefined) {
    const rows = sortedIterations(iterations);
    const dash = "—";
    return rows.map((it, idx) => {
      const prev = lastCompletedBefore(rows, idx);
      let trend: React.ReactNode = "—";
      if (it.status === "completed" && prev) {
        const d = it.score - prev.score;
        if (d > 0) {
          trend = (
            <span className="trend-up" title={`Improved vs iteration ${prev.iteration}`}>
              Good · +{d}
            </span>
          );
        } else if (d < 0) {
          trend = (
            <span className="trend-down" title={`Lower than iteration ${prev.iteration}`}>
              {d}
            </span>
          );
        } else {
          trend = (
            <span className="trend-same" title="Same score as previous completed iteration">
              Same
            </span>
          );
        }
      } else if (it.status === "completed" && !prev) {
        trend = <span className="muted">Baseline</span>;
      }

      return (
        <tr key={it.iteration}>
          <td>{it.iteration}</td>
          <td>{it.status}</td>
          <td title={metricTitle(it.specificity, "specificity")}>{fmtMetricQual(it.specificity, "specificity")}</td>
          <td title={metricTitle(it.actionability, "actionability")}>{fmtMetricQual(it.actionability, "actionability")}</td>
          <td title={metricTitle(it.severity, "severity")}>{fmtMetricQual(it.severity, "severity")}</td>
          <td title={metricTitle(it.structure, "structure")}>{fmtMetricQual(it.structure, "structure")}</td>
          <td title={it.status === "completed" ? `Score: ${it.score}` : undefined}>
            {it.status === "completed" ? it.score : dash}
          </td>
          <td>{trend}</td>
        </tr>
      );
    });
  }

  function renderRunProgressBlock(g: RunGroup) {
    const { barPct, caption } = runGroupProgressUi(g);
    return (
      <div className="run-progress" role="status" aria-live="polite">
        <div className="run-progress-track" aria-hidden>
          <div className="run-progress-fill" style={{ width: `${barPct}%` }} />
        </div>
        <div className="run-progress-caption muted">{caption}</div>
      </div>
    );
  }

  function renderPromptArtifacts(groupId: number) {
    const loading = promptArtifactsLoading[groupId];
    const data = promptArtifactsByGroup[groupId];
    if (loading) {
      return <div className="muted">Loading prompt artifacts…</div>;
    }
    if (!data) {
      return <div className="muted">Prompt artifacts unavailable.</div>;
    }
    return (
      <div className="artifacts-wrap">
        <div className="artifact-section">
          <strong>Prompt Versions</strong>
          {data.versions.length === 0 ? (
            <div className="muted">No prompt versions saved.</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>Iter</th>
                  <th>Source</th>
                  <th>Reason</th>
                </tr>
              </thead>
              <tbody>
                {data.versions.map((v) => (
                  <tr key={v.id}>
                    <td>{v.iteration}</td>
                    <td>{v.source}</td>
                    <td title={v.reason}>{truncate(v.reason || "—", 80)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
        <div className="artifact-section">
          <strong>Prompt Deltas</strong>
          {data.deltas.length === 0 ? (
            <div className="muted">No prompt deltas saved.</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>Iter</th>
                  <th>Source</th>
                  <th>Status</th>
                  <th>Reason</th>
                </tr>
              </thead>
              <tbody>
                {data.deltas.map((d) => (
                  <tr key={d.id}>
                    <td>{d.iteration}</td>
                    <td>{d.source}</td>
                    <td>{d.validation_status}</td>
                    <td title={d.reason}>{truncate(d.reason || "—", 80)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    );
  }

  /** Compact page list: 1 … 4 5 6 … 10 */
  function pageNumbersToShow(current: number, total: number): (number | "ellipsis")[] {
    if (total <= 7) {
      return Array.from({ length: total }, (_, i) => i + 1);
    }
    const out: (number | "ellipsis")[] = [];
    const edge = new Set([1, 2, total - 1, total, current - 1, current, current + 1].filter((n) => n >= 1 && n <= total));
    const sorted = [...edge].sort((a, b) => a - b);
    let prev = 0;
    for (const n of sorted) {
      if (prev && n - prev > 1) out.push("ellipsis");
      out.push(n);
      prev = n;
    }
    return out;
  }

  function renderPagination() {
    if (listTotal === 0 || totalListPages <= 1) return null;
    const pages = pageNumbersToShow(evalPage, totalListPages);
    return (
      <div className="pagination-bar" role="navigation" aria-label="Evaluation list pages">
        <button type="button" disabled={evalPage <= 1 || groupsLoading} onClick={() => setEvalPage((p) => Math.max(1, p - 1))}>
          Previous
        </button>
        {pages.map((p, i) =>
          p === "ellipsis" ? (
            <span key={`e-${i}`} className="page-ellipsis">
              …
            </span>
          ) : (
            <button
              key={p}
              type="button"
              className={p === evalPage ? "page-active" : undefined}
              disabled={groupsLoading}
              onClick={() => setEvalPage(p)}
              aria-current={p === evalPage ? "page" : undefined}
            >
              {p}
            </button>
          ),
        )}
        <button
          type="button"
          disabled={evalPage >= totalListPages || groupsLoading}
          onClick={() => setEvalPage((p) => Math.min(totalListPages, p + 1))}
        >
          Next
        </button>
      </div>
    );
  }

  return (
    <div className="container">
      <h1>Self-Improving Code Review Bot</h1>

      {/* Part 1: Top — submit + latest evaluation */}
      <div className="grid-two">
        <div className="card">
          <h2>Submit Code for Evaluation</h2>
          <textarea
            className="code-input"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            spellCheck={false}
          />
          <div style={{ marginTop: 10 }}>
            <label className="muted">Extra prompt (optional)</label>
            <textarea
              className="prompt-input"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
            />
          </div>
          <button className="btn-primary" type="button" disabled={runLoading} onClick={() => void onSubmit()}>
            {runLoading ? "Submitting…" : "Run Evaluation"}
          </button>
          {runSummary ? (
            <div className="muted" style={{ marginTop: 10 }}>
              Started run group <strong>{runSummary.run_group_id}</strong> ({runSummary.status})
            </div>
          ) : null}
        </div>

        <div className="card">
          <h2>Recent Evaluations</h2>
          {groupsLoading && !recentRunGroup ? (
            <div className="muted">Loading…</div>
          ) : groupsError && !recentRunGroup ? (
            <div className="error">{groupsError}</div>
          ) : !recentRunGroup ? (
            <div className="muted">No evaluation yet. Submit code to start.</div>
          ) : (
            <>
              <div className="muted" style={{ marginBottom: 6 }}>
                Run group #{recentRunGroup.id} · {recentRunGroup.status}
              </div>
              {renderRunProgressBlock(recentRunGroup)}
              <div>
                <strong>Input code</strong>
                <div className="input-snippet">{truncate(recentRunGroup.input_code, 400)}</div>
              </div>
              <div style={{ overflowX: "auto" }}>
                <table>
                  <thead>
                    <tr>
                      <th>Iteration</th>
                      <th>Status</th>
                      <th>Specificity</th>
                      <th>Actionability</th>
                      <th>Severity</th>
                      <th>Structure</th>
                      <th>Score</th>
                      <th>vs prev</th>
                    </tr>
                  </thead>
                  <tbody>{renderIterationRows(recentRunGroup.iterations)}</tbody>
                </table>
              </div>
            </>
          )}
        </div>
      </div>

      {/* Part 2: Mid — score trend */}
      <div className="card">
        <h2>Recent Score Trend</h2>
        {!recentRunGroup ? (
          <div className="muted">No data for chart yet.</div>
        ) : (
          <div className="chart-wrap">
            <canvas ref={canvasRef} />
          </div>
        )}
      </div>

      {/* Part 3: Bottom — all evaluations */}
      <div className="card">
        <h2>All Evaluations</h2>
        {groupsLoading && listRunGroups.length === 0 ? (
          <div className="muted">Loading…</div>
        ) : groupsError ? (
          <div className="error">{groupsError}</div>
        ) : listRunGroups.length === 0 ? (
          <div className="muted">No run groups yet.</div>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Run Group ID</th>
                <th>Status</th>
                <th>Runs Completed</th>
                <th>Final Score</th>
              </tr>
            </thead>
            <tbody>
              {listRunGroups.map((g) => {
                const completed = g.iterations?.filter((r) => r.status === "completed").length ?? 0;
                const total = g.iterations?.length ?? 0;
                const finalScore = finalScoreForGroup(g);
                const expanded = selectedRunGroupId === g.id;
                return (
                  <React.Fragment key={g.id}>
                    <tr
                      className="table-row-expand"
                      role="button"
                      tabIndex={0}
                      aria-expanded={expanded}
                      aria-label={`Run group ${g.id}, ${expanded ? "expanded" : "collapsed"}. Press Enter or Space to toggle.`}
                      onClick={() => void toggleExpand(g.id)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter" || e.key === " ") {
                          e.preventDefault();
                          void toggleExpand(g.id);
                        }
                      }}
                    >
                      <td>{g.id}</td>
                      <td>{g.status}</td>
                      <td>
                        {completed} / {total}
                      </td>
                      <td>{finalScore !== undefined ? finalScore : "—"}</td>
                    </tr>
                    {expanded ? (
                      <tr>
                        <td colSpan={4} style={{ padding: 0, borderBottom: "none" }}>
                          <div className="expand-panel">
                            <div className="muted" style={{ marginBottom: 8 }}>
                              <strong>Input code</strong>
                            </div>
                            <pre
                              style={{
                                margin: "0 0 16px",
                                padding: 12,
                                background: "#fff",
                                border: "1px solid #e2e5ea",
                                borderRadius: 4,
                                fontSize: 12,
                                overflow: "auto",
                                maxHeight: 200,
                              }}
                            >
                              {g.input_code}
                            </pre>
                            <strong>Iterations</strong>
                            <table>
                              <thead>
                                <tr>
                                  <th>Iteration</th>
                                  <th>Status</th>
                                  <th>Specificity</th>
                                  <th>Actionability</th>
                                  <th>Severity</th>
                                  <th>Structure</th>
                                  <th>Score</th>
                                  <th>vs prev</th>
                                </tr>
                              </thead>
                              <tbody>{renderIterationRows(g.iterations)}</tbody>
                            </table>
                            <div style={{ marginTop: 16 }}>
                              <strong>Prompt Evolution</strong>
                              {renderPromptArtifacts(g.id)}
                            </div>
                          </div>
                        </td>
                      </tr>
                    ) : null}
                  </React.Fragment>
                );
              })}
            </tbody>
          </table>
        )}
        {listTotal > 0 ? <div className="pagination-footer">{renderPagination()}</div> : null}
      </div>
    </div>
  );
}
