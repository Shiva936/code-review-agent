package config

import (
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Env              string          `toml:"env"`
	Port             string          `toml:"port"`
	DatabasePath     string          `toml:"database_path"`
	Provider         string          `toml:"provider"`
	OpenRouterAPIKey string          `toml:"open_router_api_key"`
	GeminiAPIKey     string          `toml:"gemini_api_key"`
	GeneratorModel   string          `toml:"generator_model"`
	EvaluatorModel   string          `toml:"evaluator_model"`
	MaxEvalRetries   int             `toml:"max_eval_retries"`
	Auth             AuthConfig      `toml:"auth"`
	RateLimit        RateLimitConfig `toml:"rate_limit"`
	Refiner          RefinerConfig   `toml:"refiner"`
}

type AuthConfig struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type RateLimitConfig struct {
	Enabled     bool                     `toml:"enabled"`
	Storage     string                   `toml:"storage"` // "redis" or "memory"
	DefaultRule RateLimitRule            `toml:"default"`
	Routes      map[string]RateLimitRule `toml:"routes"`
}

type RateLimitRule struct {
	BucketSize     int           `toml:"bucket_size"`     // Maximum tokens in bucket
	RefillSize     int           `toml:"refill_size"`     // Tokens added per refill
	RefillDuration time.Duration `toml:"refill_duration"` // How often to refill
	IdentifyBy     string        `toml:"identify_by"`     // "ip" or "api_key"
	Enabled        bool          `toml:"enabled"`
}

type RefinerConfig struct {
	Mode         string  `toml:"mode"`           // rule_based | hybrid
	Model        string  `toml:"model"`          // optional; falls back to evaluator model
	Temperature  float64 `toml:"temperature"`    // LLM refiner temperature
	MaxRules     int     `toml:"max_rules"`      // active rule budget
	MaxRuleChars int     `toml:"max_rule_chars"` // per-rule size guard
	MaxDeltaOps  int     `toml:"max_delta_ops"`  // max add/remove/modify ops per iteration
	RollbackGate bool    `toml:"rollback_gate"`  // enable score-regression rollback
	RollbackDrop int     `toml:"rollback_drop"`  // score drop threshold to trigger rollback
}

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) GetConfig() *Config {
	var err error
	if cfg == nil {
		cfg, err = cfg.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	}
	return cfg
}

// LoadDefaultConfig loads default config from env/default.toml
func (cfg *Config) LoadDefaultConfig() (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile("env/default.toml", &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadConfig loads config from default TOML and overrides from environment variables
func (cfg *Config) LoadConfig() (*Config, error) {
	cfg, err := cfg.LoadDefaultConfig()
	if err != nil {
		return nil, err
	}

	cfg.UpdateEnvConfig()
	return cfg, nil
}

// UpdateEnvConfig updates config fields by checking environment variables
func (cfg *Config) UpdateEnvConfig() {
	// Use reflection to iterate through fields and update from env vars
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	// Helper function to set fields if env matches toml tag (uppercased and underscores)
	for i := 0; i < v.NumField(); i++ {
		structField := t.Field(i)
		sectionVal := v.Field(i)
		if sectionVal.Kind() == reflect.Struct {
			updateSectionFromEnv(sectionVal, structField.Name)
		} else {
			// Handle top-level fields
			tag := structField.Tag.Get("toml")
			envKey := strings.ToUpper(tag)
			if envVal, exists := os.LookupEnv(envKey); exists {
				updateFieldValue(sectionVal, envVal)
			}
		}
	}
}

func updateSectionFromEnv(val reflect.Value, sectionName string) {
	t := val.Type()
	sectionPrefix := strings.ToUpper(sectionName) + "_"

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := t.Field(i)
		tag := fieldType.Tag.Get("toml")
		envKey := sectionPrefix + strings.ToUpper(strings.ReplaceAll(tag, "-", "_"))

		envVal, exists := os.LookupEnv(envKey)
		if !exists {
			continue
		}

		updateFieldValue(field, envVal)
	}
}

func updateFieldValue(field reflect.Value, envVal string) {
	if !field.CanSet() {
		return
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(envVal)
	case reflect.Int:
		if intVal, err := strconv.Atoi(envVal); err == nil {
			field.SetInt(int64(intVal))
		} else {
			log.Printf("Warning: invalid integer for env value: %s", envVal)
		}
	case reflect.Bool:
		if boolVal, err := strconv.ParseBool(envVal); err == nil {
			field.SetBool(boolVal)
		} else {
			log.Printf("Warning: invalid boolean for env value: %s", envVal)
		}
	case reflect.Float32, reflect.Float64:
		if floatVal, err := strconv.ParseFloat(envVal, 64); err == nil {
			field.SetFloat(floatVal)
		} else {
			log.Printf("Warning: invalid float for env value: %s", envVal)
		}
	case reflect.TypeOf(time.Duration(0)).Kind():
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			if duration, err := time.ParseDuration(envVal); err == nil {
				field.Set(reflect.ValueOf(duration))
			} else {
				log.Printf("Warning: invalid duration for env value: %s", envVal)
			}
		}
	}
}
