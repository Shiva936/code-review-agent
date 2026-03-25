# 🏗️ ARCHITECTURE.md

## 🧠 System Overview

This system implements a **self-improving loop** where an LLM generates outputs, evaluates them, and refines its own behavior over time using persistent feedback.

The architecture is intentionally **simple, modular, and traceable**, focusing on correctness and clarity rather than over-engineering.

---

# 🔁 High-Level Flow

```text
Client (React / API)
        ↓
Backend (Go API)
        ↓
Core Loop Engine
 ├── Generator
 ├── Evaluator
 └── Refiner
        ↓
LLM (OpenRouter)
        ↓
Storage (SQLite)
```

---

# 🧩 Core Components

---

## 1. Loop Engine

The loop engine orchestrates the full pipeline:

```text
for iteration in 1..5:
    review = Generate(prompt, input_code)
    score, weakness = Evaluate(review)
    prompt = Refine(prompt, weakness)
    persist results
```

### Responsibilities:

- Controls execution flow
- Aggregates scores
- Ensures autonomous operation
- Triggers persistence

---

## 2. Generator

### Purpose:

Generate structured code reviews from input code.

### Input:

- Prompt (with refinement instructions)
- Code snippet

### Output:

- Structured review:
  - categorized comments
  - severity labels

### Behavior:

- Combines system prompt + user input
- Calls LLM via OpenRouter
- Returns formatted output

---

## 3. Evaluator

### Purpose:

Score generated reviews using a rubric.

### Input:

- Generated review

### Output:

```json
{
  "actionability": int,
  "specificity": int,
  "severity": int,
  "total": int,
  "weakness_category": "string"
}
```

### Behavior:

- Uses LLM-as-judge
- Enforces strict JSON output
- Validates score bounds (1–15)
- Retries on invalid responses

### Design Choice:

LLM-based evaluation allows flexible and scalable scoring without hardcoding rules.

---

## 4. Refiner

### Purpose:

Improve future outputs based on past weaknesses.

### Input:

- Current prompt
- Weakness category

### Output:

- Updated prompt

### Strategy:

Refinement is **category-driven**, not random.

| Category      | Instruction Added                  |
| ------------- | ---------------------------------- |
| specificity   | Reference variables and code lines |
| actionability | Provide clear fixes                |
| severity      | Assign correct severity labels     |
| structure     | Organize output into categories    |

### Key Properties:

- Adds only missing instructions
- Avoids duplication
- Maintains prompt clarity

---

## 5. Storage Layer

### Technology:

SQLite (file-based persistence)

### Stored Data:

#### Runs Table

```text
iteration
score
weakness
```

#### Prompts Table

```text
version
prompt_text
reason_for_update
```

#### Run Groups Tables

The backend also stores user-triggered run groups (one input code snippet, multiple iterations):

```text
run_groups(id, input_code, base_prompt, iterations, created_at)
run_group_runs(id, group_id, iteration, score, weakness, created_at)
```

### Responsibilities:

- Persist state across restarts
- Enable traceability
- Support debugging and analysis

---

## 6. LLM Integration Layer

### Provider:

OpenRouter

### Models:

- Generator → `mistralai/mistral-7b-instruct`
- Evaluator → `meta-llama/llama-3-8b-instruct`

### Responsibilities:

- Handle API calls
- Abstract model switching
- Manage request/response formatting

---

# 🌐 API Layer

Minimal REST API:

### POST /run

- Triggers full loop execution
- Returns summary of results
- Protected by Basic Auth (if configured)
- Requires JSON body: `{ "code": "...", "prompt": "..." }`

---

### GET /runs

- Returns all past runs

---

### GET /health

- Health check endpoint

---

### GET /run-groups

- Returns paginated run groups with per-iteration results
- Protected by Basic Auth (if configured)

---

# ⚛️ Frontend Architecture

### Purpose:

Provide visibility into system behavior.

### Components:

- Runs Table
- Score Trend Graph (Chart.js)

### Data Flow:

```text
Frontend → GET /runs, POST /run, GET /run-groups → Backend → SQLite
```

### Design Principles:

- Read-only dashboard
- Minimal UI complexity
- Focus on clarity

---

# 🔄 Data Flow (Detailed)

```text
1. User triggers /run
2. Backend initializes loop
3. Generator calls LLM → produces review
4. Evaluator calls LLM → produces score
5. Refiner updates prompt
6. Results stored in SQLite
7. Loop repeats
8. Frontend fetches results
```

---

# 🔐 Security Considerations

---

## Prompt Injection

- Input code treated strictly as data
- Explicit instruction to ignore embedded commands

---

## Output Validation

- Strict JSON parsing
- Retry mechanism
- Fallback scoring

---

## API Key Protection

- Stored in environment variables
- Never exposed to frontend

---

## Concurrency Control

- Single-run lock prevents overlapping executions

---

## Data Integrity

- Safe writes via SQLite
- Avoid partial writes

---

# ⚖️ Trade-offs

| Decision              | Trade-off                            |
| --------------------- | ------------------------------------ |
| SQLite                | Simple but not horizontally scalable |
| LLM evaluator         | Flexible but non-deterministic       |
| Rule-based refinement | Simple but not adaptive learning     |
| Minimal API           | Clear but limited extensibility      |

---

# 📈 Scalability Considerations

While not required for this assignment, the system can be extended by:

- Replacing SQLite with a managed DB
- Adding queue-based execution (e.g., workers)
- Supporting dynamic inputs instead of fixed samples
- Improving evaluator with hybrid scoring

---

# 🧠 Design Philosophy

This system prioritizes:

- **Clarity over complexity**
- **Traceability over abstraction**
- **Deterministic structure over black-box behavior**

The goal is to demonstrate a **working self-improving loop**, that can be enhanced to develop production-scale AI platform.

---

# ✅ Summary

The architecture cleanly separates:

- generation
- evaluation
- refinement
- persistence

Each component is independently understandable and testable, while working together to form a complete feedback-driven system.

---
