# Sample code snippets for evaluation

Ten **Go** examples for pasting into the **Submit Code for Evaluation** box. They are **not** the same as the three fallback snippets in `backend/router/hardcoded_samples.go` (SQL injection demo, naive config parser, weak token loop).

Each sample is intentionally imperfect so the review loop has something to critique (security, logic, performance, or style).

**How runs use your paste:** the backend reviews **the code you submit** in each iteration (see `README.md` and `docs/ARCHITECTURE.md` for the Generate → Evaluate → Refine loop, iteration context, and scoring).

**If every iteration shows the same score:** restart the backend after updating, then check server logs for `EVALUATOR_FALLBACK` (judge output not parsed). The app now tolerates extra JSON fields from the model and uses per-iteration **seed** / **temperature** so scores are less likely to be identical plateaus.

---

## 1. HTTP GET without timeout or body handling

```go
func FetchURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}
```

---

## 2. `defer` inside a loop (resource leak)

```go
func ReadLines(paths []string) ([]string, error) {
	var out []string
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			out = append(out, sc.Text())
		}
	}
	return out, nil
}
```

---

## 3. Sensitive data in logs

```go
func Login(username, password string) error {
	log.Printf("login attempt user=%s password=%s", username, password)
	if username == "admin" && password == os.Getenv("ADMIN_PASS") {
		return nil
	}
	return fmt.Errorf("invalid credentials")
}
```

---

## 4. Unsynchronized map writes from goroutines

```go
var cache = map[string]int{}

func Hit(key string) {
	go func() {
		cache[key]++
	}()
}
```

---

## 5. Slice index without length check

```go
func FirstToken(line string) string {
	parts := strings.Fields(line)
	return parts[1]
}
```

---

## 6. Command execution with user-controlled input

```go
func RunUserScript(dir, scriptName string) error {
	cmd := exec.Command("sh", "-c", "cd "+dir+" && ./"+scriptName)
	return cmd.Run()
}
```

---

## 7. Ignored conversion error

```go
func ParsePort(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}
```

---

## 8. Busy wait / high CPU sleep misuse

```go
func WaitForFlag(path string) {
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
	}
}
```

---

## 9. `math/rand` for session identifiers

```go
func NewSessionID() string {
	return fmt.Sprintf("sess-%d", rand.Int63())
}
```

---

## 10. Float equality for money

```go
func IsPaid(balance, price float64) bool {
	return balance == price
}
```

---

## Usage

Copy **one** block at a time into the app, run **Run Evaluation**, and compare iterations. The agent reviews **your pasted snippet** for each run group.
