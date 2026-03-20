# ⚙️ SETUP.md

This guide will help you run the **Self-Improving Code Review Bot** locally and deploy it using free-tier services.

---

# 🧰 Prerequisites

Make sure you have the following installed:

- Go (>= 1.22)
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

```bash
export OPENROUTER_API_KEY=your_api_key
```

Optional:

```bash
export PORT=8080
```

---

# 🗄️ 4. Setup Database

No manual setup required.

- SQLite DB will be auto-created on first run
- Default location:

```text
/data/app.db
```

If running locally:

```text
./data/app.db
```

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

Trigger loop:

```bash
curl -X POST http://localhost:8080/run
```

Fetch results:

```bash
curl http://localhost:8080/runs
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

Create `.env` file inside `frontend/`:

```bash
REACT_APP_API_URL=http://localhost:8080
```

Restart frontend after adding env.

---

# 📊 9. Using the App

- Open: http://localhost:3000

- View:
  - Run history
  - Score progression graph

- Trigger runs via API (or button if implemented)

---

# 🐳 10. Docker (Optional)

Build and run backend:

```bash
docker build -t self-improving-bot .
docker run -p 8080:8080 -e OPENROUTER_API_KEY=your_key self-improving-bot
```

---

# ☁️ Deployment (Free Tier)

---

## 🚂 Backend — Railway

1. Push repo to GitHub
2. Go to Railway → New Project → Deploy from GitHub
3. Add environment variables:

```text
OPENROUTER_API_KEY=your_key
PORT=8080
```

---

### ⚠️ Enable Persistent Storage

- Add volume mount
- Set DB path to:

```text
/data/app.db
```

---

## 🌐 Frontend — Vercel

1. Import repo into Vercel
2. Set environment variable:

```text
REACT_APP_API_URL=https://your-backend-url
```

3. Deploy

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

## ❌ Scores Not Improving

**Cause:** Weak refinement
**Fix:**

- Ensure correct weakness categories
- Verify prompt updates are applied

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
