package var_template

import (
	"strings"
	"testing"
)

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		want     string
		wantErr  bool
	}{
		{
			name:     "empty template",
			template: "",
			vars:     map[string]string{},
			want:     "",
		},
		{
			name:     "template with only text",
			template: "Hello World",
			vars:     map[string]string{},
			want:     "Hello World",
		},
		{
			name:     "consecutive variables",
			template: "${first}${second}",
			vars:     map[string]string{"first": "Hello", "second": "World"},
			want:     "HelloWorld",
		},
		{
			name:     "variable at start",
			template: "${greeting} World",
			vars:     map[string]string{"greeting": "Hello"},
			want:     "Hello World",
		},
		{
			name:     "variable at end",
			template: "Hello ${name}",
			vars:     map[string]string{"name": "World"},
			want:     "Hello World",
		},
		{
			name:     "multiple spaces in variable",
			template: "${  name  }",
			vars:     map[string]string{"name": "John"},
			want:     "John",
		},
		{
			name:     "special characters in variable value",
			template: "Message: ${msg}",
			vars:     map[string]string{"msg": "Hello ${world}!"},
			want:     "Message: Hello ${world}!",
		},
		{
			name:     "unicode characters",
			template: "Hello ${name}",
			vars:     map[string]string{"name": "世界"},
			want:     "Hello 世界",
		},
		{
			name:     "empty variable value",
			template: "Hello ${name}",
			vars:     map[string]string{"name": ""},
			want:     "Hello ",
		},
		{
			name:     "variable with newlines",
			template: "Message:\n${content}",
			vars:     map[string]string{"content": "Line 1\nLine 2"},
			want:     "Message:\nLine 1\nLine 2",
		},
		{
			name:     "multiple escape sequences",
			template: "\\${escaped1} and \\${escaped2}",
			vars:     map[string]string{},
			want:     "\\${escaped1} and \\${escaped2}",
		},
		{
			name:     "mixed escaped and unescaped",
			template: "\\${escaped} and ${unescaped}",
			vars:     map[string]string{"unescaped": "value"},
			want:     "\\${escaped} and value",
		},
		{
			name:     "number variable with zero",
			template: `{"count": "${count:%d}"}`,
			vars:     map[string]string{"count": "0"},
			want:     `{"count": 0}`,
		},
		{
			name:     "number variable with negative",
			template: `{"temp": "${temp:%d}"}`,
			vars:     map[string]string{"temp": "-5"},
			want:     `{"temp": -5}`,
		},
		{
			name:     "default value with special chars",
			template: "Config: ${config?:key=value&other=123}",
			vars:     map[string]string{},
			want:     "Config: key=value&other=123",
		},
		{
			name:     "complex variable name patterns",
			template: "${var1!?:default1:%d} ${var2:+} ${var3:*}",
			vars:     map[string]string{"var1": "42"},
			want:     "42 ${var2:+} ${var3:*}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := Compile(tt.template)
			got, err := tmpl.Execute(tt.vars)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Execute() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInvalidVariableNames(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantVars []string
	}{
		{
			name:     "variable with only spaces",
			template: "${   }",
			wantVars: []string{},
		},
		{
			name:     "variable with only special chars",
			template: "${!?:}",
			wantVars: []string{},
		},
		{
			name:     "malformed variable syntax",
			template: "${name!?:}",
			wantVars: []string{"name"},
		},
		{
			name:     "variable with mixed valid/invalid parts",
			template: "${name!invalid}",
			wantVars: []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := Compile(tt.template)
			if got := tmpl.Variables(); !stringSliceEqual(got, tt.wantVars) {
				t.Errorf("Variables() = %v, want %v", got, tt.wantVars)
			}
		})
	}
}

func TestComplexDefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		want     string
	}{
		{
			name:     "default with colon",
			template: "URL: ${url?:http://localhost:8080}",
			vars:     map[string]string{},
			want:     "URL: http://localhost:8080",
		},
		{
			name:     "default with spaces",
			template: "Message: ${msg?:Hello World}",
			vars:     map[string]string{},
			want:     "Message: Hello World",
		},
		{
			name:     "default with special characters",
			template: "Pattern: ${pattern?:^[a-zA-Z0-9]+$}",
			vars:     map[string]string{},
			want:     "Pattern: ^[a-zA-Z0-9]+$",
		},
		{
			name:     "empty default value",
			template: "Value: ${value?:}",
			vars:     map[string]string{},
			want:     "Value: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := Compile(tt.template)
			got, err := tmpl.Execute(tt.vars)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Execute() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMacroEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		checkFn  func(string) bool
	}{
		{
			name:     "unknown macro",
			template: "Value: ${@unknown}",
			checkFn:  func(s string) bool { return s == "Value: ${@unknown}" },
		},
		{
			name:     "macro with spaces",
			template: "Time: ${ @timestamp }",
			checkFn:  func(s string) bool { return s != "Time: ${ @timestamp }" }, // Should be processed
		},
		{
			name:     "macro in middle",
			template: "Start ${@timestamp} End",
			checkFn:  func(s string) bool { return s != "Start ${@timestamp} End" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := Compile(tt.template)
			got, err := tmpl.Execute(map[string]string{})
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if !tt.checkFn(got) {
				t.Errorf("Execute() = %q, failed check", got)
			}
		})
	}
}

func TestLargeTemplate(t *testing.T) {
	// Test with a large template to ensure performance
	template := ""
	vars := make(map[string]string)

	for i := 0; i < 100; i++ {
		template += "Variable " + string(rune('A'+i%26)) + ": ${var" + string(rune('0'+i%10)) + "} "
		vars["var"+string(rune('0'+i%10))] = "value" + string(rune('0'+i%10))
	}

	tmpl := Compile(template)
	result, err := tmpl.Execute(vars)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}

	if len(result) == 0 {
		t.Error("Execute() returned empty result for large template")
	}

	// Verify the result contains expected patterns
	if !contains(result, "Variable A: value0") {
		t.Error("Execute() result doesn't contain expected pattern")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
