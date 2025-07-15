package var_template

import (
	"strconv"
	"strings"
	"testing"
)

// TestQuickStart tests the quick start example from README
func TestQuickStart(t *testing.T) {
	// Basic usage with ${name} syntax
	tmpl := Compile("Hello ${name}")
	result, err := tmpl.Execute(map[string]string{"name": "World"})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}
	if result != "Hello World" {
		t.Errorf("Execute() = %q, want %q", result, "Hello World")
	}

	// Using $name syntax (equivalent to ${name})
	tmpl2 := Compile("Hello $name")
	result2, err := tmpl2.Execute(map[string]string{"name": "World"})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}
	if result2 != "Hello World" {
		t.Errorf("Execute() = %q, want %q", result2, "Hello World")
	}

	// Mixed syntax with separator handling
	tmpl3 := Compile("File: $name.txt, Size: ${size?:0} bytes")
	result3, err := tmpl3.Execute(map[string]string{"name": "document"})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}
	if result3 != "File: document.txt, Size: 0 bytes" {
		t.Errorf("Execute() = %q, want %q", result3, "File: document.txt, Size: 0 bytes")
	}
}

// TestBasicVariables tests basic variable syntax
func TestBasicVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		want     string
	}{
		{
			name:     "simple variable",
			template: "Hello ${name}",
			vars:     map[string]string{"name": "World"},
			want:     "Hello World",
		},
		{
			name:     "variable with spaces",
			template: "Hello ${ name }",
			vars:     map[string]string{"name": "World"},
			want:     "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := Compile(tt.template)
			result, err := tmpl.Execute(tt.vars)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if result != tt.want {
				t.Errorf("Execute() = %q, want %q", result, tt.want)
			}
		})
	}
}

// TestDollarSyntaxExamples tests the dollar syntax examples from README
func TestDollarSyntaxExamples(t *testing.T) {
	// Simple file path template
	pathTemplate := "$dir/$name.txt"
	tmpl := Compile(pathTemplate)
	path, err := tmpl.Execute(map[string]string{
		"dir":  "/var/log",
		"name": "app",
	})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}
	if path != "/var/log/app.txt" {
		t.Errorf("Execute() = %q, want %q", path, "/var/log/app.txt")
	}

	// URL template with mixed syntax
	urlTemplate := "$scheme://$host:$port/${path?:api}"
	tmpl = Compile(urlTemplate)
	url, err := tmpl.Execute(map[string]string{
		"scheme": "https",
		"host":   "api.example.com",
		"port":   "443",
	})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}
	if url != "https://api.example.com:443/api" {
		t.Errorf("Execute() = %q, want %q", url, "https://api.example.com:443/api")
	}

	// Database connection string
	dbTemplate := "$user:$password@$host:$port/$database"
	tmpl = Compile(dbTemplate)
	dsn, err := tmpl.Execute(map[string]string{
		"user":     "admin",
		"password": "secret",
		"host":     "localhost",
		"port":     "5432",
		"database": "myapp",
	})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}
	if dsn != "admin:secret@localhost:5432/myapp" {
		t.Errorf("Execute() = %q, want %q", dsn, "admin:secret@localhost:5432/myapp")
	}
}

// TestRequiredVariables tests required variable syntax
func TestRequiredVariables(t *testing.T) {
	// Required variable should work when provided
	tmpl := Compile("Hello ${name!}")
	result, err := tmpl.Execute(map[string]string{"name": "World"})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}
	if result != "Hello World" {
		t.Errorf("Execute() = %q, want %q", result, "Hello World")
	}

	// Required variable should error when not provided
	_, err = tmpl.Execute(map[string]string{})
	if err == nil {
		t.Error("Execute() should return error for missing required variable")
	}
}

// TestDefaultValues tests default value syntax
func TestDefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		want     string
	}{
		{
			name:     "simple default",
			template: "Hello ${name?:World}",
			vars:     map[string]string{},
			want:     "Hello World",
		},
		{
			name:     "default with special characters",
			template: "URL: ${url?:http://localhost:8080}",
			vars:     map[string]string{},
			want:     "URL: http://localhost:8080",
		},
		{
			name:     "default with complex values",
			template: "Config: ${config?:key=value&other=123}",
			vars:     map[string]string{},
			want:     "Config: key=value&other=123",
		},
		{
			name:     "empty default",
			template: "Value: ${value?:}",
			vars:     map[string]string{},
			want:     "Value: ",
		},
		{
			name:     "default overridden",
			template: "Hello ${name?:World}",
			vars:     map[string]string{"name": "Go"},
			want:     "Hello Go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := Compile(tt.template)
			result, err := tmpl.Execute(tt.vars)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if result != tt.want {
				t.Errorf("Execute() = %q, want %q", result, tt.want)
			}
		})
	}
}

// TestTypeHints tests type hint syntax
func TestTypeHints(t *testing.T) {
	// Number type - removes quotes in JSON contexts
	tmpl := Compile(`{"age": "${age:%d}"}`)
	result, err := tmpl.Execute(map[string]string{"age": "25"})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}
	want := `{"age": 25}`
	if result != want {
		t.Errorf("Execute() = %q, want %q", result, want)
	}
}

// TestRepeatModes tests repeat mode syntax
func TestRepeatModes(t *testing.T) {
	tests := []struct {
		name     string
		template string
		varName  string
		wantMode repeatMode
	}{
		{
			name:     "unique mode",
			template: "Items: ${items:+}",
			varName:  "items",
			wantMode: repeatMode_Uniq,
		},
		{
			name:     "any mode",
			template: "Items: ${items:*}",
			varName:  "items",
			wantMode: repeatMode_Any,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := Compile(tt.template)
			if tmpl.NumVars() != 1 {
				t.Errorf("NumVars() = %d, want 1", tmpl.NumVars())
				return
			}

			v := tmpl.varPositions[0]
			if v.varName != tt.varName {
				t.Errorf("varName = %q, want %q", v.varName, tt.varName)
			}
			if v.repeatMode != tt.wantMode {
				t.Errorf("repeatMode = %v, want %v", v.repeatMode, tt.wantMode)
			}
		})
	}
}

// TestBuiltinMacros tests built-in macro syntax
func TestBuiltinMacros(t *testing.T) {
	tests := []struct {
		name     string
		template string
		checkFn  func(string) bool
	}{
		{
			name:     "timestamp",
			template: "Time: ${@timestamp}",
			checkFn: func(result string) bool {
				if !strings.HasPrefix(result, "Time: ") {
					return false
				}
				timestampStr := strings.TrimPrefix(result, "Time: ")
				_, err := strconv.ParseInt(timestampStr, 10, 64)
				return err == nil
			},
		},
		{
			name:     "timestamp_ms",
			template: "Time: ${@timestamp_ms}",
			checkFn: func(result string) bool {
				if !strings.HasPrefix(result, "Time: ") {
					return false
				}
				timestampStr := strings.TrimPrefix(result, "Time: ")
				_, err := strconv.ParseInt(timestampStr, 10, 64)
				return err == nil
			},
		},
		{
			name:     "timestamp_us",
			template: "Time: ${@timestamp_us}",
			checkFn: func(result string) bool {
				if !strings.HasPrefix(result, "Time: ") {
					return false
				}
				timestampStr := strings.TrimPrefix(result, "Time: ")
				_, err := strconv.ParseInt(timestampStr, 10, 64)
				return err == nil
			},
		},
		{
			name:     "timestamp_ns",
			template: "Time: ${@timestamp_ns}",
			checkFn: func(result string) bool {
				if !strings.HasPrefix(result, "Time: ") {
					return false
				}
				timestampStr := strings.TrimPrefix(result, "Time: ")
				_, err := strconv.ParseInt(timestampStr, 10, 64)
				return err == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := Compile(tt.template)
			result, err := tmpl.Execute(map[string]string{})
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if !tt.checkFn(result) {
				t.Errorf("Execute() = %q, failed validation", result)
			}
		})
	}
}

// TestComplexCombinations tests complex syntax combinations
func TestComplexCombinations(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		want     string
	}{
		{
			name:     "required with default and type hint",
			template: "Age: ${age!?:25:%d}",
			vars:     map[string]string{"age": "30"},
			want:     "Age: 30",
		},
		{
			name:     "required with default and type hint - using default",
			template: "Age: ${age!?:25:%d}",
			vars:     map[string]string{},
			want:     "Age: 25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := Compile(tt.template)
			result, err := tmpl.Execute(tt.vars)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if result != tt.want {
				t.Errorf("Execute() = %q, want %q", result, tt.want)
			}
		})
	}
}

// TestAPIReference tests API reference examples
func TestAPIReference(t *testing.T) {
	// Template creation
	tmpl := Compile("Hello ${name}")

	// Get template information
	vars := tmpl.Variables()
	if len(vars) != 1 || vars[0] != "name" {
		t.Errorf("Variables() = %v, want [name]", vars)
	}

	hasVars := tmpl.HasVariables()
	if !hasVars {
		t.Error("HasVariables() = false, want true")
	}

	numVars := tmpl.NumVars()
	if numVars != 1 {
		t.Errorf("NumVars() = %d, want 1", numVars)
	}

	// Template execution
	result, err := tmpl.Execute(map[string]string{"name": "World", "age": "25"})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}
	if result != "Hello World" {
		t.Errorf("Execute() = %q, want %q", result, "Hello World")
	}

	// Partial application
	tmpl2 := Compile("Hello ${name}, you are ${age} years old")
	partial := tmpl2.PartialApply(map[string]string{"name": "World"})
	if partial.Template() != "Hello World, you are ${age} years old" {
		t.Errorf("PartialApply() = %q, want %q", partial.Template(), "Hello World, you are ${age} years old")
	}

	// Apply with options
	tmpl3 := Compile("Hello ${name?:World}")
	result3 := tmpl3.Apply(map[string]string{}, &ApplyOptions{
		ApplyDefault:     true,
		ApplyMacro:       true,
		ValidateRequired: true,
	})
	if result3.Template() != "Hello World" {
		t.Errorf("Apply() = %q, want %q", result3.Template(), "Hello World")
	}
}

// TestVariableInformation tests variable information API
func TestVariableInformation(t *testing.T) {
	tmpl := Compile("Hello ${name!?:World:%d}")

	if tmpl.NumVars() != 1 {
		t.Errorf("NumVars() = %d, want 1", tmpl.NumVars())
		return
	}

	v := tmpl.Var(0)
	if v.Name() != "name" {
		t.Errorf("Name() = %q, want %q", v.Name(), "name")
	}
	if !v.Required() {
		t.Error("Required() = false, want true")
	}
	if !v.HasDefault() {
		t.Error("HasDefault() = false, want true")
	}
	if v.IsMacro() {
		t.Error("IsMacro() = true, want false")
	}
	if !v.IsNumber() {
		t.Error("IsNumber() = false, want true")
	}
}

// TestConfigurationTemplate tests the configuration template example
func TestConfigurationTemplate(t *testing.T) {
	configTemplate := `{
    "host": "${host?:localhost}",
    "port": "${port?:8080:%d}",
    "ssl": "${ssl?:false}",
    "timeout": "${timeout?:30:%d}",
    "created_at": "${@timestamp}"
}`

	tmpl := Compile(configTemplate)
	config, err := tmpl.Execute(map[string]string{
		"host": "api.example.com",
		"port": "443",
		"ssl":  "true",
	})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}

	// Check that the result contains the expected values
	if !strings.Contains(config, `"host": "api.example.com"`) {
		t.Errorf("Config should contain host: %v", config)
	}
	if !strings.Contains(config, `"port": 443`) {
		t.Errorf("Config should contain port without quotes: %v", config)
	}
	if !strings.Contains(config, `"ssl": "true"`) {
		t.Errorf("Config should contain ssl: %v", config)
	}
	if !strings.Contains(config, `"timeout": 30`) {
		t.Errorf("Config should contain timeout without quotes: %v", config)
	}
	if !strings.Contains(config, `"created_at":`) {
		t.Errorf("Config should contain created_at: %v", config)
	}
}

// TestURLBuilder tests the URL builder example
func TestURLBuilder(t *testing.T) {
	urlTemplate := "${scheme?:https}://${host!}:${port?:80:%d}${path?:/}"

	tmpl := Compile(urlTemplate)
	url, err := tmpl.Execute(map[string]string{
		"host": "api.example.com",
		"port": "443",
		"path": "/v1/users",
	})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}

	want := "https://api.example.com:443/v1/users"
	if url != want {
		t.Errorf("Execute() = %q, want %q", url, want)
	}
}

// TestSQLQueryTemplate tests the SQL query template example
func TestSQLQueryTemplate(t *testing.T) {
	queryTemplate := `SELECT * FROM ${table!} 
WHERE ${column?:id} = ${value!} 
LIMIT ${limit?:10:%d}`

	tmpl := Compile(queryTemplate)
	query, err := tmpl.Execute(map[string]string{
		"table": "users",
		"value": "123",
		"limit": "5",
	})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}

	expectedParts := []string{
		"SELECT * FROM users",
		"WHERE id = 123",
		"LIMIT 5",
	}

	for _, part := range expectedParts {
		if !strings.Contains(query, part) {
			t.Errorf("Query should contain %q: %v", part, query)
		}
	}
}

// TestErrorHandling tests error handling examples
func TestErrorHandling(t *testing.T) {
	tmpl := Compile("Hello ${name!}")

	// This should return an error because 'name' is required
	_, err := tmpl.Execute(map[string]string{})
	if err == nil {
		t.Error("Execute() should return error for missing required variable")
	}

	// Check error message contains the variable name
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("Error message should contain variable name: %v", err)
	}
}

// TestPerformance tests performance examples
func TestPerformance(t *testing.T) {
	// Compile once, use many times
	tmpl := Compile("Hello ${name}")

	// Fast execution
	for i := 0; i < 100; i++ {
		result, err := tmpl.Execute(map[string]string{"name": "World"})
		if err != nil {
			t.Errorf("Execute() error = %v", err)
			return
		}
		if result != "Hello World" {
			t.Errorf("Execute() = %q, want %q", result, "Hello World")
		}
	}
}

// TestAllFeaturesCombined tests a comprehensive example with all features
func TestAllFeaturesCombined(t *testing.T) {
	complexTemplate := `{
    "name": "${name!}",
    "age": "${age?:25:%d}",
    "url": "${url?:http://localhost:8080}",
    "timestamp": "${@timestamp}",
    "config": "${config?:key=value&other=123}",
    "items": "${items:+}",
    "debug": "${debug?:false}"
}`

	tmpl := Compile(complexTemplate)
	result, err := tmpl.Execute(map[string]string{
		"name":  "John Doe",
		"age":   "30",
		"url":   "https://api.example.com:8443",
		"items": "item1,item2,item3",
	})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}

	// Check that the result contains expected values
	expectedParts := []string{
		`"name": "John Doe"`,
		`"age": 30`,
		`"url": "https://api.example.com:8443"`,
		`"timestamp":`,
		`"config": "key=value&other=123"`,
		`"items": "item1,item2,item3"`,
		`"debug": "false"`,
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Result should contain %q: %v", part, result)
		}
	}

	// Verify timestamp is a valid number
	timestampStart := strings.Index(result, `"timestamp": `) + len(`"timestamp": `)
	timestampEnd := strings.Index(result[timestampStart:], ",")
	if timestampEnd == -1 {
		timestampEnd = strings.Index(result[timestampStart:], "}")
	}
	timestampStr := strings.TrimSpace(result[timestampStart : timestampStart+timestampEnd])
	// Remove quotes if present
	timestampStr = strings.Trim(timestampStr, `"`)
	if _, err := strconv.ParseInt(timestampStr, 10, 64); err != nil {
		t.Errorf("Timestamp should be a valid number: %v", timestampStr)
	}
}
