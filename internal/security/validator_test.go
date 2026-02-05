package security

import (
	"strings"
	"testing"
)

func TestValidateReadOnlyQuery_ValidQueries(t *testing.T) {
	validator := NewValidator(10000)

	validQueries := []string{
		"SELECT * FROM users",
		"SELECT id, name FROM users WHERE status = 'active'",
		"SELECT COUNT(*) FROM orders",
		"select * from products",
		"EXPLAIN SELECT * FROM users",
		"DESCRIBE users",
		"SHOW TABLES",
		"WITH cte AS (SELECT * FROM users) SELECT * FROM cte",
		"SELECT u.id, o.total FROM users u JOIN orders o ON u.id = o.user_id",
	}

	for _, query := range validQueries {
		t.Run(query, func(t *testing.T) {
			err := validator.ValidateReadOnlyQuery(query)
			if err != nil {
				t.Errorf("Expected valid query to pass, got error: %v", err)
			}
		})
	}
}

func TestValidateReadOnlyQuery_WriteOperations(t *testing.T) {
	validator := NewValidator(10000)

	writeQueries := []struct {
		query       string
		shouldContain string
	}{
		{"INSERT INTO users (name) VALUES ('test')", "INSERT"},
		{"UPDATE users SET name = 'test'", "UPDATE"},
		{"DELETE FROM users WHERE id = 1", "DELETE"},
		{"DROP TABLE users", "DROP"},
		{"CREATE TABLE test (id INT)", "CREATE"},
		{"ALTER TABLE users ADD COLUMN age INT", "ALTER"},
		{"TRUNCATE TABLE users", "TRUNCATE"},
		{"REPLACE INTO users (id, name) VALUES (1, 'test')", "REPLACE"},
		{"GRANT SELECT ON *.* TO 'user'@'%'", "GRANT"},
		{"REVOKE SELECT ON *.* FROM 'user'@'%'", "REVOKE"},
	}

	for _, tc := range writeQueries {
		t.Run(tc.query, func(t *testing.T) {
			err := validator.ValidateReadOnlyQuery(tc.query)
			if err == nil {
				t.Errorf("Expected write query to fail, but it passed")
			}
			if !strings.Contains(err.Error(), tc.shouldContain) {
				t.Errorf("Expected error to contain '%s', got: %v", tc.shouldContain, err)
			}
		})
	}
}

func TestValidateReadOnlyQuery_SQLInjection(t *testing.T) {
	validator := NewValidator(10000)

	injectionAttempts := []string{
		"SELECT * FROM users; DROP TABLE users",
		"SELECT * FROM users -- comment",
		"SELECT * FROM users /* comment */",
		"SELECT * FROM users WHERE id = 1 OR 1=1--",
	}

	for _, query := range injectionAttempts {
		t.Run(query, func(t *testing.T) {
			err := validator.ValidateReadOnlyQuery(query)
			if err == nil {
				t.Errorf("Expected SQL injection attempt to fail, but it passed")
			}
		})
	}
}

func TestValidateReadOnlyQuery_InvalidStart(t *testing.T) {
	validator := NewValidator(10000)

	invalidQueries := []string{
		"CALL some_procedure()",
		"EXEC sp_something",
		"BEGIN TRANSACTION",
	}

	for _, query := range invalidQueries {
		t.Run(query, func(t *testing.T) {
			err := validator.ValidateReadOnlyQuery(query)
			if err == nil {
				t.Errorf("Expected query with invalid start to fail, but it passed")
			}
		})
	}
}

func TestValidateReadOnlyQuery_EmptyQuery(t *testing.T) {
	validator := NewValidator(10000)

	err := validator.ValidateReadOnlyQuery("")
	if err == nil {
		t.Error("Expected empty query to fail")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error about empty query, got: %v", err)
	}
}

func TestValidateReadOnlyQuery_TooLong(t *testing.T) {
	validator := NewValidator(100) // Small limit for testing

	longQuery := "SELECT * FROM users WHERE name = '" + strings.Repeat("a", 200) + "'"
	err := validator.ValidateReadOnlyQuery(longQuery)
	if err == nil {
		t.Error("Expected long query to fail")
	}
	if !strings.Contains(err.Error(), "maximum length") {
		t.Errorf("Expected error about length, got: %v", err)
	}
}

func TestSanitizeTableName_Valid(t *testing.T) {
	validNames := []string{
		"users",
		"user_orders",
		"users123",
		"public.users",
		"`users`",
		"\"users\"",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			err := SanitizeTableName(name)
			if err != nil {
				t.Errorf("Expected valid table name to pass, got error: %v", err)
			}
		})
	}
}

func TestSanitizeTableName_Invalid(t *testing.T) {
	invalidNames := []string{
		"",
		"users; DROP TABLE users",
		"users--",
		"users/*",
		"users*/",
		"users DROP",
		"users DELETE",
	}

	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			err := SanitizeTableName(name)
			if err == nil {
				t.Errorf("Expected invalid table name '%s' to fail, but it passed", name)
			}
		})
	}
}

func TestContainsKeyword(t *testing.T) {
	tests := []struct {
		query    string
		keyword  string
		expected bool
	}{
		{"DELETE FROM users", "DELETE", true},
		{"SELECT deleted FROM users", "DELETE", false}, // Should not match 'deleted'
		{"SELECT * FROM selection", "SELECT", true},    // Should match 'SELECT' at start
		{"UPDATE users SET name = 'test'", "UPDATE", true},
		{"SELECT * FROM updated_users", "UPDATE", false}, // Should not match 'updated'
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := containsKeyword(strings.ToUpper(tt.query), tt.keyword)
			if result != tt.expected {
				t.Errorf("containsKeyword(%q, %q) = %v, expected %v",
					tt.query, tt.keyword, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidateReadOnlyQuery(b *testing.B) {
	validator := NewValidator(10000)
	query := "SELECT id, name, email FROM users WHERE status = 'active' LIMIT 100"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateReadOnlyQuery(query)
	}
}

func BenchmarkSanitizeTableName(b *testing.B) {
	tableName := "users"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeTableName(tableName)
	}
}
