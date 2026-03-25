package router

// Fallback snippets when submitted code is empty (e.g. tests). Normal runs use the user’s code.
var hardcodedCodeSamples = []string{
	`// NOTE: This is intentionally unsafe for the demo.
func FindUserByEmail(db *sql.DB, email string) (string, error) {
	query := "SELECT id, name FROM users WHERE email = '" + email + "'"
	row := db.QueryRow(query)
	var name string
	if err := row.Scan(&name); err != nil {
		return "", err
	}
	return name, nil
}`,
	`func ParseConfig(path string) map[string]string {
	data, _ := os.ReadFile(path)
	lines := strings.Split(string(data), "\n")
	cfg := make(map[string]string)
	for _, line := range lines {
		parts := strings.Split(line, "=")
		cfg[parts[0]] = parts[1]
	}
	return cfg
}`,
	`func GenerateToken(n int) string {
	token := ""
	for i := 0; i < n; i++ {
		token += string(rune('a' + rand.Intn(26)))
	}
	return token
}`,
}
