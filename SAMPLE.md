# 🧪 SAMPLE.md

This file contains **sample code snippets** used by the system for generating and evaluating code reviews.

These samples are intentionally designed to include **security, logic, and performance issues** to test the effectiveness of the Generate → Evaluate → Refine loop.

---

# 📌 Sample 1 — SQL Injection Risk (Security)

```go
// NOTE: This is intentionally unsafe for the demo.
func FindUserByEmail(db *sql.DB, email string) (string, error) {
	// Vulnerability: string concatenation in query construction
	query := "SELECT id, name FROM users WHERE email = '" + email + "'"
	row := db.QueryRow(query)

	var name string
	if err := row.Scan(&name); err != nil {
		return "", err
	}
	return name, nil
}
```

### 🔍 Expected Issues

- SQL injection vulnerability
- Unsafe query construction
- Missing parameterized query usage

---

# 📌 Sample 2 — Panic & Weak Error Handling (Logic + Severity)

```go
func ParseConfig(path string) map[string]string {
	data, _ := os.ReadFile(path) // Ignoring error is a bug
	lines := strings.Split(string(data), "\n")

	cfg := make(map[string]string)
	for _, line := range lines {
		parts := strings.Split(line, "=")
		// Vulnerability: parts[1] panics when "=" is missing
		cfg[parts[0]] = parts[1]
	}
	return cfg
}
```

### 🔍 Expected Issues

- Ignored error from file read
- Potential runtime panic (`index out of range`)
- Missing input validation
- Weak error handling

---

# 📌 Sample 3 — Inefficient Performance + Insecure Randomness

```go
func GenerateToken(n int) string {
	// Insecure: math/rand is not suitable for tokens.
	// Inefficient: repeated string concatenation in a loop.
	token := ""
	for i := 0; i < n; i++ {
		token += string(rune('a' + rand.Intn(26)))
	}
	return token
}
```

### 🔍 Expected Issues

- Use of `math/rand` for security-sensitive token generation
- Inefficient string concatenation (O(n²))
- Lack of cryptographic randomness (`crypto/rand`)

---

# 🎯 Purpose

These samples are used to:

- Test **generation quality** (does the LLM identify real issues?)
- Test **evaluation accuracy** (are issues scored correctly?)
- Drive **refinement** (does the system improve over iterations?)

---

# ⚠️ Notes

- All samples are intentionally flawed for demonstration
- They are not production-safe implementations
- The system should progressively improve its ability to detect these issues

---
