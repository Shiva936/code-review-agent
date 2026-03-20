import React, { useEffect, useMemo, useRef, useState } from "react";
import Chart from "chart.js/auto";

type RunRow = {
  iteration: number;
  score: number;
  weakness: string;
};

type RunsResponse = {
  runs: RunRow[];
};

export default function App() {
  const [runs, setRuns] = useState<RunRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const chartRef = useRef<Chart | null>(null);

  const scoreSeries = useMemo(() => runs.map((r) => r.score), [runs]);
  const labels = useMemo(() => runs.map((r) => r.iteration), [runs]);

  useEffect(() => {
    let cancelled = false;

    async function fetchRuns() {
      setLoading(true);
      setError(null);

      try {
        const res = await fetch("/runs");
        if (!res.ok) {
          throw new Error(`GET /runs failed (${res.status})`);
        }
        const json = (await res.json()) as RunsResponse;
        if (cancelled) return;
        setRuns(json.runs || []);
      } catch (e) {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : "Unknown error");
      } finally {
        if (cancelled) return;
        setLoading(false);
      }
    }

    fetchRuns();

    return () => {
      cancelled = true;
    };
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

  return (
    <div className="container">
      <h1>Self-Improving Code Review Bot</h1>

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
              Scores reflect the average across 3 code samples per iteration.
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

