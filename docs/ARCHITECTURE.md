# Architecture

## System overview

The service implements a **closed loop**: generate a code review, score it with a rubric judge, refine the reviewer prompt, persist state, and repeat. The design is **modular and traceable**—closer to a **single pipeline** with explicit hooks than a multi-agent orchestrator.

For **how this compares** to research-oriented autonomous loops (e.g. scoring, persistence, evaluation placement), see the learning references in `README.md` (AgentHub, karpathy/autoresearch)—those are **patterns to study**, not code this repo copies.

---

## High-level flow

```text
Client (React / API)
        ↓
Backend (Go API)
        ↓
Core loop
 ├── Generator (LLM, user code + prompt + iteration context)
 ├── Evaluator (LLM-as-judge, rubric JSON + iteration context)
 └── Refiner (rule-based prompt updates + reinforcement)
        ↓
OpenRouter (chat completions; temperature optional per call)
        ↓
SQLite (run groups, per-iteration rows, metrics JSON)
```

---

## Core components

### 1. Loop engine (`router/run_processor.go`)

Orchestrates:

```text
for iteration in 1..5:
    for each code sample (usually one: user submission):
        review = Generate(prompt, code, iteration_context)
        scores = Evaluate(review, iteration_context)
    aggregate weakest issue category (logic / performance / security / style)
    persist iteration metrics
    prompt = Refine(prompt, weakness, iteration)
```

- **User code** is the **primary** sample under review (fallback snippets exist only if code is empty).
- **Iteration context** includes: iteration index, previous review text for **that** sample (from the prior iteration only), and previous aggregate rubric total—so generation and evaluation are not identical on every pass.

---

### 2. Generator (`core/generator`)

- **Input:** config, reviewer prompt (including refined rules), code snippet, **optional** `models.IterationContext`.
- **Output:** free-form structured review text (validated for category + severity keywords).
- **LLM:** `CallLLMWithOpts` with **temperature** (~0.82), `top_p` (~0.92), and per-iteration seed to reduce repeated outputs.
- **Behavior:** after iteration 1, the prompt includes a **truncated previous review** and instructions to improve, not copy verbatim.

---

### 3. Evaluator (`core/evaluator`)

- **Input:** `IterationContext` (iteration, optional previous rubric total).
- **Output:** strict JSON (`EvalResult`) with rubric totals + issue-category scores.
- **LLM:** `CallLLMWithOpts` with **lower** sampling settings than generation (temperature ~0.55, top_p ~0.9) and per-iteration seed.
- **Prompt:** asks for **independent** dimension scores and discourages defaulting all rubric cells to the same middle value when quality differs.
- **Resilience:** tolerates extra JSON keys from the model, normalizes weakness category, and falls back only after retries.

---

### 4. Refiner (`core/refiner`)

- **Input:** weakness string from **aggregate issue-category** averages (not the rubric-only weakness in `EvalResult`).
- **Output:** prompt with appended rules; **repeating** the same base weakness adds an **escalating reinforcement** line so the prompt does not stall.

---

### 5. LLM layer (`llm/openrouter.go`)

- `CallLLM` / `CallLLMWithOpts` → OpenRouter `chat/completions`.
- **Sampling options** (`temperature`, `top_p`, `seed`) are passed when configured.
- HTTP client uses a **timeout** for long generations.

---

## Configuration surface (`backend/env/default.toml`)

Top-level runtime knobs:

- `open_router_api_key`
- `generator_model`
- `evaluator_model`
- `max_eval_retries`
- `port`, `database_path` (optional)

These can be overridden by environment variables:
`OPEN_ROUTER_API_KEY`, `GENERATOR_MODEL`, `EVALUATOR_MODEL`, `MAX_EVAL_RETRIES`, `PORT`, `DATABASE_PATH`.

---

### 6. Storage (`storage`)

- SQLite with WAL mode (see `SETUP.md` for `app.db` / `-wal` / `-shm`).
- **Run groups** hold user-submitted code and status; **run group runs** hold per-iteration score, weakness, and optional `detail_json` for metrics.

---

## API (minimal)

| Method | Path | Notes |
|--------|------|--------|
| GET | `/health` | Liveness |
| POST | `/run` | Body: `{ "code", "prompt" }`; Basic Auth; starts async loop |
| GET | `/run-groups?page=N` | Paginated groups + iterations |
| GET | `/runs` | Legacy runs list |

---

## Frontend

- React dashboard: submit code, **Recent Evaluations**, score chart, **All Evaluations** with pagination.
- Polls `/run-groups` periodically for live status.

---

## Security notes (summary)

- Treat user code as **data** in prompts; reject injection attempts in system instructions.
- API keys only on the server.
- Evaluator JSON validated with retries; fallback scores if parsing fails.

---

## Trade-offs

| Decision | Trade-off |
|----------|-----------|
| LLM judge | Flexible; not deterministic |
| Rule-based refiner | Simple; not learned policy |
| SQLite | Easy; single-node |
| Single pipeline | Clear; not a general agent framework |

---

## Summary

The architecture cleanly separates **generation**, **evaluation**, **refinement**, and **persistence**, with **explicit evaluation hooks** at each iteration and **state** carried forward so the loop can be described and debugged—similar in spirit to “close the loop and persist” systems, but scoped to **code review + rubric** only.
