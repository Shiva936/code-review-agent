# ⚙️ SETUP.md

This guide will help you run the **Self-Improving Code Review Bot** locally.

---

# 🧰 Prerequisites

Make sure you have the following installed:

- Go (>= 1.25) (matches `backend/go.mod`)
- Node.js (>= 18)
- npm or yarn
- Git

---

# 🔑 1. Get OpenRouter API Key

1. Go to https://openrouter.ai
2. Create an account
3. Generate an API key

---

# 🌱 2. Clone Repository

```bash
git clone <repo-url>
cd <repo>
```

---

# 🔐 3. Set Environment Variables

### Backend

The backend loads defaults from `backend/env/default.toml`, and then overrides any fields from environment variables.

Recommended environment variables:

- `PROVIDER`: LLM provider selection (`openrouter` or `gemini`)
- `GEMINI_API_KEY`: Google Gemini API key (required when `PROVIDER=gemini`)
- `OPEN_ROUTER_API_KEY`: OpenRouter API key (overrides `open_router_api_key`)
- `GENERATOR_MODEL`: generation model (overrides `generator_model`)
- `EVALUATOR_MODEL`: evaluator/judge model (overrides `evaluator_model`)
- `MAX_EVAL_RETRIES`: evaluator retry count (overrides `max_eval_retries`)
- `PORT`: HTTP port (overrides `port`)
- `DATABASE_PATH`: SQLite path (overrides `database_path` if set in TOML)
- `AUTH_USERNAME`, `AUTH_PASSWORD`: Basic Auth credentials for protected endpoints

PowerShell example:

```powershell
$env:PROVIDER="gemini"
$env:GEMINI_API_KEY="your_gemini_key"
$env:OPEN_ROUTER_API_KEY="your_api_key"
$env:GENERATOR_MODEL="gemini-1.5-flash"
$env:EVALUATOR_MODEL="gemini-1.5-flash"
$env:MAX_EVAL_RETRIES="3"
$env:PORT="8080"
$env:AUTH_USERNAME="admin"
$env:AUTH_PASSWORD="changeme"
```

---

# 🗄️ 4. Setup Database

No manual setup required.

- SQLite DB will be auto-created on first run
- Default location (if not configured):

```text
/data/app.db
```

If running locally (from `backend/`):

```text
./data/app.db
```

SQLite may also create **`app.db-wal`** and **`app.db-shm`** next to the main file when WAL journaling is enabled—this is one database, not three separate apps.

---

# ▶️ 5. Run Backend

```bash
cd backend
go mod tidy
go run main.go
```

Expected output:

```text
Server running on port 8080
```

---

# 🧪 6. Test Backend

Health check:

```bash
curl http://localhost:8080/health
```

Fetch runs:

```bash
curl http://localhost:8080/runs
```

Trigger loop (`/run` is protected by Basic Auth; it also requires a JSON body):

```bash
curl -i -X POST "http://localhost:8080/run" ^
  -H "Content-Type: application/json" ^
  -u admin:changeme ^
  --data "{\"code\":\"package main\\n\\nfunc main() {}\\n\",\"prompt\":\"\"}"
```

Fetch run groups (`/run-groups` is protected by Basic Auth; pagination uses `page`):

```bash
curl -i "http://localhost:8080/run-groups?page=1" -u admin:changeme
```

---

# ⚛️ 7. Run Frontend

```bash
cd frontend
npm install
npm start
```

---

# 🔗 8. Connect Frontend to Backend

Create `.env` file inside `frontend/` (optional; you can also set settings in the UI):

```bash
VITE_API_URL=http://localhost:8080
VITE_AUTH_USERNAME=admin
VITE_AUTH_PASSWORD=changeme
```

Restart frontend after adding env.

---

# 📊 9. Using the App

- Open the frontend shown in your terminal (Vite default is usually `http://localhost:5173`).

- View:
  - Run history
  - Score progression graph

- Trigger runs from the UI (POST `/run`) or via curl.

---

# 🐳 10. Docker (Optional)

Build and run backend:

```bash
docker build -t self-improving-bot .
docker run -p 8080:8080 -e OPEN_ROUTER_API_KEY=your_key self-improving-bot
```

(Use the same env var name as in local setup: `OPEN_ROUTER_API_KEY` unless your image maps it differently.)

---

# 🔐 Security Notes

- API keys are stored in environment variables
- LLM calls happen only in backend
- No secrets are exposed to frontend

---

# ⚠️ Common Issues & Fixes

---

## ❌ JSON Parsing Errors

**Cause:** LLM returns invalid JSON
**Fix:**

- Retry logic already implemented
- Ensure evaluator prompt enforces strict JSON

---

## ❌ Scores Flat Across Iterations

**Cause:** Models may collapse to similar outputs; older builds ignored user code or stalled the refiner.

**Fix (current behavior):**

- Submit **non-empty** code you care about; the loop reviews **your** snippet.
- The pipeline uses **iteration-aware** generation/evaluation (previous review + prior rubric total), **non-zero temperature** on LLM calls, and **reinforcement** when the same weakness repeats.
- If scores still barely move, try another model in code or add detail in the optional **extra prompt** field.

For architectural context on agent loops and scoring hooks, see `README.md` (learning references) and `docs/ARCHITECTURE.md`.

If backend logs show **`EVALUATOR_FALLBACK`** on every sample, the judge JSON failed validation after retries—check OpenRouter errors in logs. Older builds used strict JSON decoding that rejected valid extra keys and **always** fell back to the same 9/15 score each iteration.

---

## ❌ SQLite Reset (Deployment)

**Cause:** No persistent volume
**Fix:**

- Ensure Railway volume is mounted
- DB path set to `/data/app.db`

---

## ❌ CORS Errors

**Fix:**
Allow frontend origin in backend:

```go
Access-Control-Allow-Origin: *
```

(or restrict to your domain)

---

# 🧪 Development Tips

- Start with backend only
- Verify loop + scoring works
- Then add frontend
- Then deploy

---

# ✅ Final Checklist

- [ ] Backend runs locally
- [ ] Loop executes successfully
- [ ] Scores logged across runs
- [ ] Frontend displays results
- [ ] Deployment working (optional but recommended)

---

# 💡 Notes

- Free-tier LLMs may produce slight inconsistencies
- Minor score fluctuations are expected
- Overall improvement trend is the key signal

---

# 🚀 You're Ready

Once setup is complete:

- Run the system
- Observe improvement across iterations
- Suggest Improvements

---
