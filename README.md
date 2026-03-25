# 🧠 Self-Improving Code Review Bot

A feedback-driven system that improves its own outputs across runs using a **Generate → Evaluate → Refine** loop powered by LLMs.

This project implements a minimal, fully traceable pipeline where a code review agent iteratively enhances its performance by learning from past weaknesses — with persistent state and measurable improvement.

![UI Design](docs/images/ui.jpeg)

---

# 🚀 Overview

This project explores **automated code review** using a feedback-driven system.

Given a **user-submitted** code snippet (see `docs/SAMPLE.md` for extra examples), the system:

1. **Generates** a structured code review for that snippet
2. **Evaluates** the review using a scoring rubric (LLM-as-judge)
3. **Refines** its prompt based on detected weaknesses (category-driven rules + reinforcement when a weakness repeats)
4. Repeats for **5 iterations** and persists each step

The objective is to demonstrate **measurable change** across iterations (scores may rise, dip slightly, or plateau depending on the model; the pipeline is designed so prompts and reviews are not identical every time).

---

# 🔁 Core Loop

```text
Generate → Evaluate → Refine → Persist → Repeat
```

### Generate

- Produces structured code reviews
- Includes:
  - categorized comments (logic, performance, security, style)
  - severity labels (critical, minor, suggestion)

### Evaluate

- Uses LLM-as-judge to score output. The judge receives **iteration context** (e.g. iteration 3 of 5, previous rubric total) so it can differentiate runs instead of collapsing to identical middle scores.
- **Sampling:** generation uses higher **temperature** than evaluation to reduce mode-collapse (same text/scores every call).


| Metric        | Description                            |
| ------------- | -------------------------------------- |
| Actionability | Are fixes clearly suggested? (1–5)     |
| Specificity   | References variables/lines? (1–5)      |
| Severity      | Correct severity classification? (1–5) |


Total score: **/15** (sum of the three rubric dimensions)

### Refine

- Aggregates weakest **issue category** (logic / performance / security / style) across samples
- Appends a targeted rule to the reviewer prompt; if the same weakness repeats, an **escalating reinforcement** line is added so the prompt keeps changing

### Persist

- Stores:
  - run results
  - prompt versions
  - weakness history

---

# 🏗️ Architecture

```text
Frontend (React)
        ↓
Backend API (Go)
        ↓
Core Loop (Generator / Evaluator / Refiner)
        ↓
LLM (OpenRouter)
        ↓
Storage (SQLite)
```

---

# 📁 Project Structure

```text
/backend
  /config
  /core
    generator/
    evaluator/
    refiner/
  /llm
  /storage
  /router
  /models
  main.go

/frontend
  /src
    api.ts
    config.ts

docs/SAMPLE.md (optional paste examples for the UI)

backend/data (SQLite; WAL sidecar files may appear next to `app.db`)
```

---

# ⚙️ Tech Stack

- **Backend:** Go
- **Frontend:** React (minimal dashboard)
- **LLM Provider:** OpenRouter
- **Models (defaults in config):** `generator_model` and `evaluator_model` in `backend/env/default.toml` (override via `GENERATOR_MODEL` / `EVALUATOR_MODEL`).
- **Evaluation retries:** `max_eval_retries` in `backend/env/default.toml` (override via `MAX_EVAL_RETRIES`).
- **Database:** SQLite (WAL mode; see `docs/SETUP.md` for `app.db` / `-wal` / `-shm`)

---

# ▶️ How It Works

1. User (or frontend) triggers `POST /run` with a code snippet (and optional extra prompt)
2. The system runs **5 iterations** on **that snippet** (stored for the UI; the loop uses it as the code under review)
3. Each iteration:
  - generates a review (with **previous iteration’s review** as context after iteration 1)
  - evaluates the review (with **iteration index** and **previous rubric total**)
  - refines the prompt for the next iteration
  - persists results

---

## Problem, where each loop phase lives, and five-run scores

### Problem we picked

**Automated, feedback-driven code review:** for a **user-submitted** snippet (or optional demo samples if code is empty), the system runs a fixed **five-iteration** loop. Each iteration **generates** a structured review, **evaluates** it with an LLM-as-judge rubric (actionability, specificity, severity), **refines** the reviewer prompt from the weakest issue category, and **persists** scores and metrics so improvement is visible across runs—not a one-shot review.

### Where each loop phase lives in the code


| Phase                                | What it does                                                                  | Where in the repo                                                                                     |
| ------------------------------------ | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| **Orchestrate** (loop + aggregation) | Runs iterations 1–5, wires context between steps, aggregates weakest category | `backend/router/run_processor.go` (`processRunGroup`; also `samplesForRun`, `aggregateAcrossSamples`) |
| **Generate**                         | Produces the review text from prompt + code + iteration context               | `backend/core/generator/generator.go` (called from `run_processor.go`)                                |
| **Evaluate**                         | Parses rubric JSON, scores review, yields weakness category                   | `backend/core/evaluator/evaluator.go`                                                                 |
| **Refine**                           | Appends category rules + reinforcement to the prompt for the next iteration   | `backend/core/refiner/refiner.go`                                                                     |
| **Persist**                          | Saves per-iteration rows, metrics JSON, run-group status                      | `run_processor.go` → `backend/storage/` (e.g. `UpdateRunGroupRun`, `SaveRun`, …)                      |
| **LLM calls**                        | OpenRouter HTTP client                                                        | `backend/llm/openrouter.go`                                                                           |
| **HTTP entry**                       | Starts a run group                                                            | `backend/router/` (e.g. `handler_run.go`, `routes.go`)                                                |


### Score table across five iterations

Below is an example with numbers from a sample run (UI **Recent Evaluations** / **All Evaluations**, or SQLite). **Total** is out of **15** (actionability + specificity + severity, each 1–5). **Weakness (refine target)** is the **aggregate issue category** with the lowest mean across samples—always one of **logic**, **performance**, **security**, or **style** (see `aggregateAcrossSamples` in `backend/router/iteration_aggregate.go`).


| Iteration | Total (/15) | Actionability | Specificity | Severity | Weakness (refine target) |
| --------- | ----------- | ------------- | ----------- | -------- | ------------------------ |
| 1         | 8           | 3             | 2           | 3        | style                    |
| 2         | 10          | 3             | 4           | 3        | security                 |
| 3         | 12          | 4             | 4           | 4        | performance              |
| 4         | 13          | 4             | 5           | 4        | logic                    |
| 5         | 14          | 5             | 5           | 4        | logic                    |


---

# 🧠 Key Design Decisions

### LLM-as-Judge

Evaluation is handled by an LLM to simulate flexible, human-like scoring.

---

### Targeted Refinement

Instead of blindly modifying prompts, refinement is **category-driven**:

- specificity → enforce references to code elements
- actionability → enforce clear fixes
- severity → enforce correct classification

---

### Persistence (Required)

All state is stored:

- run history
- prompt versions
- weakness patterns

This ensures learning survives restarts.

---

### Controlled Loop Execution

- Fixed number of iterations (5)
- Fully autonomous after trigger
- No manual intervention

---

# 🔐 Threat Model & Mitigations

### 1. Prompt Injection

**Risk:** Malicious code influencing model behavior
**Mitigation:** Treat input strictly as data, enforce system instructions

---

### 2. Output Format Injection

**Risk:** Invalid JSON responses
**Mitigation:** Strict JSON enforcement + retry + fallback scoring

---

### 3. API Key Leakage

**Risk:** Exposure of OpenRouter credentials
**Mitigation:** Backend-only usage via environment variables

---

### 4. Denial of Service

**Risk:** Repeated triggering of loop
**Mitigation:** Single-run lock to prevent concurrent execution, plus configurable rate limiting

---

### 5. Data Integrity

**Risk:** Corrupted or partial writes
**Mitigation:** Safe persistence using SQLite / atomic writes

---

### 6. Model Unreliability

**Risk:** Inconsistent outputs
**Mitigation:** Validation, retries, bounded scoring

---

### 7. Frontend Injection

**Risk:** Unsafe rendering of model output
**Mitigation:** No raw HTML rendering in React

---

# 📚 Learning references (patterns, not copied code)

These are useful for how **agentic loops** are structured—task decomposition, evaluation hooks, persistence—not as code to paste into this repo.

- **[AgentHub](https://www.agenthub.dev)** — Browse real agent pipelines; notice where evaluation and branching happen in the loop.
- **[karpathy/autoresearch](https://github.com/karpathy/autoresearch)** — Research-style autonomous loop with scoring and state; closest public spirit to “close the loop and persist.”

**How this project differs (short):**


|               | This repo                                              | autoresearch-style systems        |
| ------------- | ------------------------------------------------------ | --------------------------------- |
| Domain        | Code review + rubric judge                             | Research / training experiments   |
| State         | SQLite run groups + per-iteration metrics              | Varies (checkpoints, logs, etc.)  |
| “Improvement” | Prompt rules + reinforcement + iteration-aware scoring | Task-specific metrics and tooling |


You should be able to **say in your own words** how your loop closes (generate → evaluate → refine → persist) and where the **evaluation hook** sits versus open-ended agent steps.

---

# 📈 Observations

- Iterations are **designed** to diverge (temperature, prior review context, refiner, judge instructions); **monotonic score increase is not guaranteed** on free or small models.
- Minor fluctuations are normal; compare **trends** and qualitative review text, not only a single number.

---

# ⚠️ Limitations

- LLM-based evaluation is heuristic, not deterministic
- Free-tier models may produce inconsistent formatting
- Refinement logic is rule-based, not learned

---

# ✅ Requirements Checklist

- Generate → Evaluate → Refine loop
- Persistent state across runs
- Scores logged across iterations
- Autonomous execution
- No external agent frameworks

---

# 💡 Summary

This project demonstrates a simple but powerful idea:

> Systems can improve their outputs by identifying weaknesses, adapting behavior, and repeating the loop.

---

