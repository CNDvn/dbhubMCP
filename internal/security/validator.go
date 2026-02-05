package security

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Patterns for dangerous SQL keywords
	writeKeywords = []string{
		"INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
		"TRUNCATE", "REPLACE", "MERGE", "GRANT", "REVOKE",
	}

	// Regex patterns for SQL injection detection
	sqlInjectionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i);\s*(DROP|DELETE|UPDATE|INSERT|CREATE|ALTER|TRUNCATE)`),
		regexp.MustCompile(`(?i)--`),              // SQL comment
		regexp.MustCompile(`(?i)/\*.*\*/`),        // Multi-line comment
		regexp.MustCompile(`(?i)xp_cmdshell`),     // SQL Server command execution
		regexp.MustCompile(`(?i)exec\s*\(`),       // Execute statement
	}

	// Allow SELECT and EXPLAIN statements
	allowedKeywords = []string{"SELECT", "EXPLAIN", "DESCRIBE", "SHOW", "WITH"}
)

// Validator handles SQL query validation
type Validator struct {
	maxQueryLength int
}

// NewValidator creates a new query validator
func NewValidator(maxQueryLength int) *Validator {
	if maxQueryLength <= 0 {
		maxQueryLength = 10000 // default 10KB
	}
	return &Validator{
		maxQueryLength: maxQueryLength,
	}
}

// ValidateReadOnlyQuery checks if a query is read-only and safe
func (v *Validator) ValidateReadOnlyQuery(query string) error {
	// Check query length
	if len(query) > v.maxQueryLength {
		return fmt.Errorf("query exceeds maximum length of %d characters", v.maxQueryLength)
	}

	// Normalize query for checking
	normalizedQuery := strings.TrimSpace(query)
	upperQuery := strings.ToUpper(normalizedQuery)

	// Check if query is empty
	if normalizedQuery == "" {
		return fmt.Errorf("query cannot be empty")
	}

	// Check for write operations
	for _, keyword := range writeKeywords {
		if containsKeyword(upperQuery, keyword) {
			return fmt.Errorf("write operation detected: %s is not allowed", keyword)
		}
	}

	// Check if query starts with allowed keywords
	startsWithAllowed := false
	for _, keyword := range allowedKeywords {
		if strings.HasPrefix(upperQuery, keyword) {
			startsWithAllowed = true
			break
		}
	}
	if !startsWithAllowed {
		return fmt.Errorf("query must start with SELECT, EXPLAIN, DESCRIBE, SHOW, or WITH")
	}

	// Check for SQL injection patterns
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(query) {
			return fmt.Errorf("potentially dangerous SQL pattern detected")
		}
	}

	return nil
}

// containsKeyword checks if a keyword exists as a whole word in the query
func containsKeyword(query, keyword string) bool {
	// Use word boundary regex to match whole words only
	pattern := regexp.MustCompile(`\b` + keyword + `\b`)
	return pattern.MatchString(query)
}

// SanitizeTableName validates and sanitizes table names
func SanitizeTableName(tableName string) error {
	if tableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// Allow alphanumeric, underscore, dot (for schema.table), and backticks/quotes
	validPattern := regexp.MustCompile(`^[\w\.` + "`" + `"]+$`)
	if !validPattern.MatchString(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}

	// Check for SQL injection attempts
	dangerous := []string{"--", "/*", "*/", ";", "DROP", "DELETE", "UPDATE"}
	upperTable := strings.ToUpper(tableName)
	for _, d := range dangerous {
		if strings.Contains(upperTable, d) {
			return fmt.Errorf("potentially dangerous table name: %s", tableName)
		}
	}

	return nil
}
