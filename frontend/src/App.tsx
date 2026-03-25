import React, { useEffect, useMemo, useRef, useState } from "react";
import Chart from "chart.js/auto";
import { api, type LoopSummary, type RunGroup, type RunRow } from "./api";
import { clearConfig, loadConfig, saveConfig } from "./config";

export default function App() {
  const [cfg, setCfg] = useState(() => loadConfig());
  const [runs, setRuns] = useState<RunRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [code, setCode] = useState("package main\n\nfunc main() {}\n");
  const [prompt, setPrompt] = useState("");
  const [runLoading, setRunLoading] = useState(false);
  const [runSummary, setRunSummary] = useState<LoopSummary | null>(null);

  const [groups, setGroups] = useState<RunGroup[]>([]);
  const [groupsLoading, setGroupsLoading] = useState(false);
  const [groupsError, setGroupsError] = useState<string | null>(null);

  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const chartRef = useRef<Chart | null>(null);

  const scoreSeries = useMemo(() => runs.map((r) => r.score), [runs]);
  const labels = useMemo(() => runs.map((r) => r.iteration), [runs]);

  async function refreshRuns() {
    setLoading(true);
    setError(null);
    try {
      const json = await api.getRuns();
      setRuns(json.runs || []);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      setLoading(false);
    }
  }

  async function refreshGroups() {
    setGroupsLoading(true);
    setGroupsError(null);
    try {
      const json = await api.getRunGroups(20, 0);
      setGroups(json.groups || []);
    } catch (e) {
      setGroupsError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      setGroupsLoading(false);
    }
  }

  useEffect(() => {
    refreshRuns();
  }, []);

  useEffect(() => {
    if (!canvasRef.current) return;

    const ctx = canvasRef.current.getContext("2d");
    if (!ctx) return;

    // Update chart on runs changes.
    if (chartRef.current) {
      chartRef.current.destroy();
      chartRef.current = null;
    }

    chartRef.current = new Chart(ctx, {
      type: "line",
      data: {
        labels,
        datasets: [
          {
            label: "Avg Score",
            data: scoreSeries,
            borderColor: "rgba(56, 189, 248, 1)",
            backgroundColor: "rgba(56, 189, 248, 0.15)",
            borderWidth: 2,
            pointRadius: 4,
            tension: 0.25,
          },
        ],
      },
      options: {
        responsive: true,
        plugins: {
          legend: { display: true },
          tooltip: { enabled: true },
        },
        scales: {
          y: {
            beginAtZero: true,
            suggestedMax: 15,
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
  }, [labels, scoreSeries]);

  async function onRun() {
    setRunLoading(true);
    setRunSummary(null);
    setError(null);
    try {
      const summary = await api.run(code, prompt);
      setRunSummary(summary);
      await refreshRuns();
      await refreshGroups();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      setRunLoading(false);
    }
  }

  function onSaveConfig() {
    saveConfig(cfg);
    setCfg(loadConfig());
  }

  function onResetConfig() {
    clearConfig();
    setCfg(loadConfig());
  }

  return (
    <div className="container">
      <h1>Self-Improving Code Review Bot</h1>

      <div className="card">
        <h3 style={{ marginTop: 0 }}>Backend Settings</h3>
        <div className="row">
          <div style={{ flex: "1 1 320px" }}>
            <label className="muted">API Base URL</label>
            <input
              value={cfg.apiBaseUrl}
              onChange={(e) => setCfg((c) => ({ ...c, apiBaseUrl: e.target.value }))}
              placeholder="http://localhost:8080"
              style={{ width: "100%" }}
            />
          </div>
          <div style={{ flex: "1 1 220px" }}>
            <label className="muted">Basic Auth Username</label>
            <input
              value={cfg.auth.username}
              onChange={(e) => setCfg((c) => ({ ...c, auth: { ...c.auth, username: e.target.value } }))}
              style={{ width: "100%" }}
            />
          </div>
          <div style={{ flex: "1 1 220px" }}>
            <label className="muted">Basic Auth Password</label>
            <input
              value={cfg.auth.password}
              onChange={(e) => setCfg((c) => ({ ...c, auth: { ...c.auth, password: e.target.value } }))}
              type="password"
              style={{ width: "100%" }}
            />
          </div>
        </div>
        <div className="row" style={{ marginTop: 12 }}>
          <button onClick={onSaveConfig}>Save settings</button>
          <button className="secondary" onClick={onResetConfig}>
            Reset
          </button>
          <button className="secondary" onClick={() => { refreshRuns(); refreshGroups(); }}>
            Refresh
          </button>
        </div>
        <div className="muted" style={{ marginTop: 8, fontSize: 12 }}>
          Tip: you can also set defaults via `VITE_API_URL`, `VITE_AUTH_USERNAME`, `VITE_AUTH_PASSWORD`.
        </div>
      </div>

      <div className="card">
        <h3 style={{ marginTop: 0 }}>Trigger a Run</h3>
        <div className="row" style={{ alignItems: "flex-start" }}>
          <div style={{ flex: "1 1 420px" }}>
            <label className="muted">Code</label>
            <textarea
              value={code}
              onChange={(e) => setCode(e.target.value)}
              rows={10}
              style={{ width: "100%", fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace" }}
            />
          </div>
          <div style={{ flex: "1 1 420px" }}>
            <label className="muted">Extra prompt (optional)</label>
            <textarea
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              rows={10}
              style={{ width: "100%" }}
            />
          </div>
        </div>
        <div className="row" style={{ marginTop: 12 }}>
          <button onClick={onRun} disabled={runLoading}>
            {runLoading ? "Running..." : "POST /run"}
          </button>
        </div>
        {runSummary ? (
          <div className="muted" style={{ marginTop: 8 }}>
            Run complete. Group ID: <b>{runSummary.group_id}</b>
          </div>
        ) : null}
      </div>

      <div className="card">
        <div className="row" style={{ alignItems: "flex-start" }}>
          <div style={{ flex: "1 1 320px" }}>
            <h3 style={{ marginTop: 0 }}>Run History</h3>

            {loading ? (
              <div className="muted">Loading runs...</div>
            ) : error ? (
              <div className="error">{error}</div>
            ) : runs.length === 0 ? (
              <div className="muted">No runs yet. Trigger `POST /run`.</div>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>Iteration</th>
                    <th>Score</th>
                    <th>Weakness</th>
                  </tr>
                </thead>
                <tbody>
                  {runs.map((r) => (
                    <tr key={r.iteration}>
                      <td>{r.iteration}</td>
                      <td>{r.score}</td>
                      <td>{r.weakness}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          <div style={{ flex: "1 1 360px" }}>
            <h3 style={{ marginTop: 0 }}>Score Trend</h3>
            <canvas ref={canvasRef} />
            <div className="muted" style={{ marginTop: 8, fontSize: 12 }}>
              Scores reflect the per-iteration score of the latest run.
            </div>
          </div>
        </div>
      </div>

      <div className="card">
        <div className="row" style={{ alignItems: "flex-start" }}>
          <div style={{ flex: "1 1 320px" }}>
            <h3 style={{ marginTop: 0 }}>Run Groups (Protected)</h3>
            {groupsLoading ? (
              <div className="muted">Loading run groups...</div>
            ) : groupsError ? (
              <div className="error">{groupsError}</div>
            ) : groups.length === 0 ? (
              <div className="muted">No groups yet.</div>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>Iterations</th>
                    <th>Created</th>
                    <th>Runs</th>
                  </tr>
                </thead>
                <tbody>
                  {groups.map((g) => (
                    <tr key={g.id}>
                      <td>{g.id}</td>
                      <td>{g.iterations}</td>
                      <td>{g.created_at}</td>
                      <td>{g.runs?.length ?? 0}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
            <div className="muted" style={{ marginTop: 8, fontSize: 12 }}>
              If this shows 401, set username/password above (backend `auth` config).
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

