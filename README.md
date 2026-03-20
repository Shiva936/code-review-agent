# 🧠 Self-Improving Code Review Bot

A feedback-driven system that improves its own outputs across runs using a **Generate → Evaluate → Refine** loop powered by LLMs.

This project implements a minimal, fully traceable pipeline where a code review agent iteratively enhances its performance by learning from past weaknesses — with persistent state and measurable improvement.

---

# 🚀 Overview

This project solves the problem of **Code Review's** by using a feedback-driven system.

Given a code snippet, the system:

1. **Generates** a structured code review
2. **Evaluates** the review using a scoring rubric
3. **Refines** its prompt based on detected weaknesses
4. Repeats this loop across multiple runs

The objective is to demonstrate **score improvement over iterations**.

---

# 🔁 Core Loop

```text
Generate → Evaluate → Refine → Persist → Repeat
```

### Generate

* Produces structured code reviews
* Includes:

  * categorized comments (logic, performance, security, style)
  * severity labels (critical, minor, suggestion)

### Evaluate

* Uses LLM-as-judge to score output:

| Metric        | Description                            |
| ------------- | -------------------------------------- |
| Actionability | Are fixes clearly suggested? (1–5)     |
| Specificity   | References variables/lines? (1–5)      |
| Severity      | Correct severity classification? (1–5) |

Total score: **/15**

### Refine

* Identifies weakest category
* Updates prompt with targeted instructions
* Maintains persistent weakness patterns

### Persist

* Stores:

  * run results
  * prompt versions
  * weakness history

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
  /core
    generator/
    evaluator/
    refiner/
  /llm
  /storage
  /models
  main.go

/frontend
  /src
    /components
    /pages

/data (SQLite or JSON persistence)
```

---

# ⚙️ Tech Stack

* **Backend:** Go
* **Frontend:** React (minimal dashboard)
* **LLM Provider:** OpenRouter (free tier)
* **Models:**

  * Generation → `mistralai/mistral-7b-instruct`
  * Evaluation → `meta-llama/llama-3-8b-instruct`
* **Database:** SQLite

---

# ▶️ How It Works

1. User (or API call) triggers the loop
2. System processes 3 predefined code snippets
3. Each snippet is:

   * reviewed
   * scored
   * used for refinement
4. After each iteration:

   * average score is calculated
   * prompt is updated
   * results are persisted

The loop runs **autonomously for 5 iterations**.

---

# 📊 Example Output

```text
Run 1 → Score: 8 | Weakness: specificity
Run 2 → Score: 10 | Weakness: severity
Run 3 → Score: 12 | Weakness: actionability
Run 4 → Score: 13 | Weakness: minor
Run 5 → Score: 14 | Weakness: none
```

Scores improve as the system adapts its prompt based on previous failures.

---

# 🧠 Key Design Decisions

### LLM-as-Judge

Evaluation is handled by an LLM to simulate flexible, human-like scoring.

---

### Targeted Refinement

Instead of blindly modifying prompts, refinement is **category-driven**:

* specificity → enforce references to code elements
* actionability → enforce clear fixes
* severity → enforce correct classification

---

### Persistence (Required)

All state is stored:

* run history
* prompt versions
* weakness patterns

This ensures learning survives restarts.

---

### Controlled Loop Execution

* Fixed number of iterations (5)
* Fully autonomous after trigger
* No manual intervention

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
**Mitigation:** Single-run lock to prevent concurrent execution

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

# 📈 Observations

* Scores improve across iterations due to targeted refinement
* Minor fluctuations may occur due to model randomness
* Structured constraints significantly improve output quality

---

# ⚠️ Limitations

* LLM-based evaluation is heuristic, not deterministic
* Free-tier models may produce inconsistent formatting
* Refinement logic is rule-based, not learned

---

# ✅ Requirements Checklist

* [x] Generate → Evaluate → Refine loop
* [x] Persistent state across runs
* [x] Scores logged across iterations
* [x] Autonomous execution
* [x] No external agent frameworks

---

# 💡 Summary

This project demonstrates a simple but powerful idea:

> Systems can improve their outputs by identifying weaknesses, adapting behavior, and repeating the loop.

---
