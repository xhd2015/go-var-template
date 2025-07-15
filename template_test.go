package var_template

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCompile(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantVars []string
		wantNum  int
	}{
		{
			name:     "simple variable",
			template: "Hello ${name}",
			wantVars: []string{"name"},
			wantNum:  1,
		},
		{
			name:     "multiple variables",
			template: "Hello ${name}, you are ${age} years old",
			wantVars: []string{"age", "name"},
			wantNum:  2,
		},
		{
			name:     "no variables",
			template: "Hello world",
			wantVars: []string{},
			wantNum:  0,
		},
		{
			name:     "escaped variable",
			template: "Hello \\${name}",
			wantVars: []string{},
			wantNum:  0,
		},
		{
			name:     "variable with spaces",
			template: "Hello ${ name }",
			wantVars: []string{"name"},
			wantNum:  1,
		},
		{
			name:     "empty variable name",
			template: "Hello ${}",
			wantVars: []string{},
			wantNum:  0,
		},
		{
			name:     "unclosed variable",
			template: "Hello ${name",
			wantVars: []string{},
			wantNum:  0,
		},
		{
			name:     "duplicate variables",
			template: "Hello ${name} and ${name}",
			wantVars: []string{"name"},
			wantNum:  2,
		},
		// New tests for $name syntax
		{
			name:     "simple dollar variable",
			template: "Hello $name",
			wantVars: []string{"name"},
			wantNum:  1,
		},
		{
			name:     "dollar variable with separator",
			template: "Hello $name.txt",
			wantVars: []string{"name"},
			wantNum:  1,
		},
		{
			name:     "dollar variable with underscore",
			template: "Hello $name_suffix",
			wantVars: []string{"name_suffix"},
			wantNum:  1,
		},
		{
			name:     "mixed syntax",
			template: "Hello $name and ${age}",
			wantVars: []string{"age", "name"},
			wantNum:  2,
		},
		{
			name:     "multiple dollar variables",
			template: "$host:$port/$path",
			wantVars: []string{"host", "path", "port"},
			wantNum:  3,
		},
		{
			name:     "escaped dollar variable",
			template: "Hello \\$name",
			wantVars: []string{},
			wantNum:  0,
		},
		{
			name:     "dollar at end",
			template: "Hello $name",
			wantVars: []string{"name"},
			wantNum:  1,
		},
		{
			name:     "dollar macro",
			template: "Time: $@timestamp",
			wantVars: []string{"@timestamp"},
			wantNum:  1,
		},
		{
			name:     "invalid dollar patterns",
			template: "$ $1name $-invalid",
			wantVars: []string{},
			wantNum:  0,
		},
		{
			name:     "dollar followed by brace",
			template: "Test $name and ${other}",
			wantVars: []string{"name", "other"},
			wantNum:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := Compile(tt.template)
			if got := tmpl.Variables(); !stringSliceEqual(got, tt.wantVars) {
				t.Errorf("Variables() = %v, want %v", got, tt.wantVars)
			}
			if got := tmpl.NumVars(); got != tt.wantNum {
				t.Errorf("NumVars() = %v, want %v", got, tt.wantNum)
			}
		})
	}
}

func TestParseVarName(t *testing.T) {
	tests := []struct {
		name           string
		varName        string
		wantVarName    string
		wantRequired   bool
		wantHasDefault bool
		wantDefaultVal string
		wantIsNumber   bool
		wantIsMacro    bool
		wantRepeatMode repeatMode
	}{
		{
			name:        "simple variable",
			varName:     "name",
			wantVarName: "name",
		},
		{
			name:         "required variable",
			varName:      "name!",
			wantVarName:  "name",
			wantRequired: true,
		},
		{
			name:           "variable with default",
			varName:        "name?:John",
			wantVarName:    "name",
			wantHasDefault: true,
			wantDefaultVal: "John",
		},
		{
			name:         "number variable",
			varName:      "age:%d",
			wantVarName:  "age",
			wantIsNumber: true,
		},
		{
			name:           "required number with default",
			varName:        "age!?:25:%d",
			wantVarName:    "age",
			wantRequired:   true,
			wantHasDefault: true,
			wantDefaultVal: "25",
			wantIsNumber:   true,
		},
		{
			name:           "repeat mode uniq",
			varName:        "items:+",
			wantVarName:    "items",
			wantRepeatMode: repeatMode_Uniq,
		},
		{
			name:           "repeat mode any",
			varName:        "items:*",
			wantVarName:    "items",
			wantRepeatMode: repeatMode_Any,
		},
		{
			name:        "macro variable",
			varName:     "@timestamp",
			wantVarName: "@timestamp",
			wantIsMacro: true,
		},
		{
			name:        "variable with spaces",
			varName:     " name ",
			wantVarName: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := parseVarName(tt.varName)
			if v.varName != tt.wantVarName {
				t.Errorf("varName = %v, want %v", v.varName, tt.wantVarName)
			}
			if v.required != tt.wantRequired {
				t.Errorf("required = %v, want %v", v.required, tt.wantRequired)
			}
			if v.hasDefaultValue != tt.wantHasDefault {
				t.Errorf("hasDefaultValue = %v, want %v", v.hasDefaultValue, tt.wantHasDefault)
			}
			if v.defaultValue != tt.wantDefaultVal {
				t.Errorf("defaultValue = %v, want %v", v.defaultValue, tt.wantDefaultVal)
			}
			if v.isNumber != tt.wantIsNumber {
				t.Errorf("isNumber = %v, want %v", v.isNumber, tt.wantIsNumber)
			}
			if v.isMacro != tt.wantIsMacro {
				t.Errorf("isMacro = %v, want %v", v.isMacro, tt.wantIsMacro)
			}
			if v.repeatMode != tt.wantRepeatMode {
				t.Errorf("repeatMode = %v, want %v", v.repeatMode, tt.wantRepeatMode)
			}
		})
	}
}

func TestTemplateExecute(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		want     string
		wantErr  bool
	}{
		{
			name:     "simple substitution",
			template: "Hello ${name}",
			vars:     map[string]string{"name": "John"},
			want:     "Hello John",
		},
		{
			name:     "multiple substitutions",
			template: "Hello ${name}, you are ${age} years old",
			vars:     map[string]string{"name": "John", "age": "25"},
			want:     "Hello John, you are 25 years old",
		},
		{
			name:     "no variables",
			template: "Hello world",
			vars:     map[string]string{},
			want:     "Hello world",
		},
		{
			name:     "variable with default",
			template: "Hello ${name?:World}",
			vars:     map[string]string{},
			want:     "Hello World",
		},
		{
			name:     "variable with default overridden",
			template: "Hello ${name?:World}",
			vars:     map[string]string{"name": "John"},
			want:     "Hello John",
		},
		{
			name:     "required variable missing",
			template: "Hello ${name!}",
			vars:     map[string]string{},
			wantErr:  true,
		},
		{
			name:     "number variable with quotes",
			template: `{"age": "${age:%d}"}`,
			vars:     map[string]string{"age": "25"},
			want:     `{"age": 25}`,
		},
		{
			name:     "number variable without quotes",
			template: "Age: ${age:%d}",
			vars:     map[string]string{"age": "25"},
			want:     "Age: 25",
		},
		// New tests for $name syntax
		{
			name:     "simple dollar substitution",
			template: "Hello $name",
			vars:     map[string]string{"name": "John"},
			want:     "Hello John",
		},
		{
			name:     "dollar variable with separator",
			template: "File: $name.txt",
			vars:     map[string]string{"name": "test"},
			want:     "File: test.txt",
		},
		{
			name:     "dollar variable with underscore",
			template: "Var: $name_suffix",
			vars:     map[string]string{"name_suffix": "value"},
			want:     "Var: value",
		},
		{
			name:     "mixed syntax execution",
			template: "Hello $name, you are ${age} years old",
			vars:     map[string]string{"name": "John", "age": "25"},
			want:     "Hello John, you are 25 years old",
		},
		{
			name:     "multiple dollar variables",
			template: "$host:$port/$path",
			vars:     map[string]string{"host": "localhost", "port": "8080", "path": "api"},
			want:     "localhost:8080/api",
		},
		{
			name:     "dollar at end",
			template: "Hello $name",
			vars:     map[string]string{"name": "World"},
			want:     "Hello World",
		},
		{
			name:     "dollar macro",
			template: "Time: $@timestamp",
			vars:     map[string]string{},
			want:     "Time: " + strconv.FormatInt(time.Now().Unix(), 10),
		},
		{
			name:     "consecutive dollar variables",
			template: "$first$second",
			vars:     map[string]string{"first": "Hello", "second": "World"},
			want:     "HelloWorld",
		},
		{
			name:     "dollar variable with spaces around",
			template: "Hello $name !",
			vars:     map[string]string{"name": "John"},
			want:     "Hello John !",
		},
		{
			name:     "dollar variable in quotes",
			template: `"$name"`,
			vars:     map[string]string{"name": "value"},
			want:     `"value"`,
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
			if !tt.wantErr {
				// For timestamp tests, we need to be more flexible
				if strings.Contains(tt.template, "@timestamp") {
					if !strings.HasPrefix(got, "Time: ") {
						t.Errorf("Execute() = %v, should start with 'Time: '", got)
					}
					// Check if the timestamp part is a valid number
					timestampStr := strings.TrimPrefix(got, "Time: ")
					if _, err := strconv.ParseInt(timestampStr, 10, 64); err != nil {
						t.Errorf("Execute() timestamp = %v, should be a valid number", timestampStr)
					}
				} else if got != tt.want {
					t.Errorf("Execute() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestTemplatePartialApply(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		vars        map[string]string
		want        string
		wantVars    []string
		wantNumVars int
	}{
		{
			name:        "partial application",
			template:    "Hello ${name}, you are ${age} years old",
			vars:        map[string]string{"name": "John"},
			want:        "Hello John, you are ${age} years old",
			wantVars:    []string{"age"},
			wantNumVars: 1,
		},
		{
			name:        "no variables provided",
			template:    "Hello ${name}",
			vars:        map[string]string{},
			want:        "Hello ${name}",
			wantVars:    []string{"name"},
			wantNumVars: 1,
		},
		{
			name:        "all variables provided",
			template:    "Hello ${name}",
			vars:        map[string]string{"name": "John"},
			want:        "Hello John",
			wantVars:    []string{},
			wantNumVars: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := Compile(tt.template)
			result := tmpl.PartialApply(tt.vars)
			if result.Template() != tt.want {
				t.Errorf("PartialApply() template = %v, want %v", result.Template(), tt.want)
			}
			if got := result.Variables(); !stringSliceEqual(got, tt.wantVars) {
				t.Errorf("PartialApply() variables = %v, want %v", got, tt.wantVars)
			}
			if got := result.NumVars(); got != tt.wantNumVars {
				t.Errorf("PartialApply() numVars = %v, want %v", got, tt.wantNumVars)
			}
		})
	}
}

func TestTemplateApply(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		opts     *ApplyOptions
		want     string
		wantVars []string
	}{
		{
			name:     "apply with defaults",
			template: "Hello ${name?:World}",
			vars:     nil,
			opts:     &ApplyOptions{ApplyDefault: true},
			want:     "Hello World",
			wantVars: []string{},
		},
		{
			name:     "apply without defaults",
			template: "Hello ${name?:World}",
			vars:     nil,
			opts:     &ApplyOptions{ApplyDefault: false},
			want:     "Hello ${name?:World}",
			wantVars: []string{"name"},
		},
		{
			name:     "apply with macros",
			template: "Time: ${@timestamp}",
			vars:     nil,
			opts:     &ApplyOptions{ApplyMacro: true},
			want:     "Time: " + strconv.FormatInt(time.Now().Unix(), 10),
			wantVars: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := Compile(tt.template)
			result := tmpl.Apply(tt.vars, tt.opts)

			// For timestamp tests, we need to be more flexible
			if strings.Contains(tt.template, "@timestamp") && tt.opts.ApplyMacro {
				if !strings.HasPrefix(result.Template(), "Time: ") {
					t.Errorf("Apply() template = %v, should start with 'Time: '", result.Template())
				}
				// Check if the timestamp part is a valid number
				timestampStr := strings.TrimPrefix(result.Template(), "Time: ")
				if _, err := strconv.ParseInt(timestampStr, 10, 64); err != nil {
					t.Errorf("Apply() timestamp = %v, should be a valid number", timestampStr)
				}
			} else {
				if result.Template() != tt.want {
					t.Errorf("Apply() template = %v, want %v", result.Template(), tt.want)
				}
			}

			if got := result.Variables(); !stringSliceEqual(got, tt.wantVars) {
				t.Errorf("Apply() variables = %v, want %v", got, tt.wantVars)
			}
		})
	}
}

func TestTemplateMacros(t *testing.T) {
	tests := []struct {
		name     string
		template string
		macro    string
	}{
		{
			name:     "timestamp macro",
			template: "Time: ${@timestamp}",
			macro:    "timestamp",
		},
		{
			name:     "timestamp_ms macro",
			template: "Time: ${@timestamp_ms}",
			macro:    "timestamp_ms",
		},
		{
			name:     "timestamp_us macro",
			template: "Time: ${@timestamp_us}",
			macro:    "timestamp_us",
		},
		{
			name:     "timestamp_ns macro",
			template: "Time: ${@timestamp_ns}",
			macro:    "timestamp_ns",
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

			// Check if the result contains a timestamp
			if !strings.HasPrefix(result, "Time: ") {
				t.Errorf("Execute() result = %v, should start with 'Time: '", result)
			}

			timestampStr := strings.TrimPrefix(result, "Time: ")
			if _, err := strconv.ParseInt(timestampStr, 10, 64); err != nil {
				t.Errorf("Execute() timestamp = %v, should be a valid number", timestampStr)
			}
		})
	}
}

func TestTemplateHelperMethods(t *testing.T) {
	tmpl := Compile("Hello ${name}, you are ${age:%d} years old")

	if !tmpl.HasVariables() {
		t.Error("HasVariables() should return true")
	}

	if !tmpl.HasNonMacroVariables() {
		t.Error("HasNonMacroVariables() should return true")
	}

	if tmpl.GetGetRaw("name") != "name" {
		t.Errorf("GetGetRaw('name') = %v, want 'name'", tmpl.GetGetRaw("name"))
	}

	if tmpl.GetGetRaw("age") != "age:%d" {
		t.Errorf("GetGetRaw('age') = %v, want 'age:%%d'", tmpl.GetGetRaw("age"))
	}

	if tmpl.GetGetRaw("nonexistent") != "" {
		t.Errorf("GetGetRaw('nonexistent') = %v, want ''", tmpl.GetGetRaw("nonexistent"))
	}

	// Test Var interface
	if tmpl.NumVars() != 2 {
		t.Errorf("NumVars() = %v, want 2", tmpl.NumVars())
	}

	var0 := tmpl.Var(0)
	if var0.Name() != "name" { // Variables appear in order they are found in template
		t.Errorf("Var(0).Name() = %v, want 'name'", var0.Name())
	}

	if var0.IsNumber() != false {
		t.Errorf("Var(0).IsNumber() = %v, want false", var0.IsNumber())
	}

	var1 := tmpl.Var(1)
	if var1.Name() != "age" {
		t.Errorf("Var(1).Name() = %v, want 'age'", var1.Name())
	}

	if var1.IsNumber() != true {
		t.Errorf("Var(1).IsNumber() = %v, want true", var1.IsNumber())
	}
}

func TestTemplateMacroOnly(t *testing.T) {
	tmpl := Compile("Time: ${@timestamp}")

	if !tmpl.HasVariables() {
		t.Error("HasVariables() should return true for macro variables")
	}

	if tmpl.HasNonMacroVariables() {
		t.Error("HasNonMacroVariables() should return false for macro-only template")
	}

	var0 := tmpl.Var(0)
	if !var0.IsMacro() {
		t.Error("Var(0).IsMacro() should return true")
	}
}

func TestTemplateUpdateVars(t *testing.T) {
	tmpl := Compile("Hello ${name}")

	originalVars := tmpl.Variables()
	if !stringSliceEqual(originalVars, []string{"name"}) {
		t.Errorf("Original variables = %v, want ['name']", originalVars)
	}

	newVars := []string{"name", "age"}
	tmpl.UpdateVars(newVars)

	updatedVars := tmpl.Variables()
	if !stringSliceEqual(updatedVars, newVars) {
		t.Errorf("Updated variables = %v, want %v", updatedVars, newVars)
	}
}

func TestComplexTemplate(t *testing.T) {
	template := `{
		"name": "${name!}",
		"age": "${age:%d}",
		"city": "${city?:Unknown}",
		"timestamp": "${@timestamp}"
	}`

	tmpl := Compile(template)

	// Test with all variables provided
	vars := map[string]string{
		"name": "John",
		"age":  "25",
		"city": "New York",
	}

	result, err := tmpl.Execute(vars)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}

	// Check that the result contains the expected values
	if !strings.Contains(result, `"name": "John"`) {
		t.Errorf("Result should contain name: %v", result)
	}

	if !strings.Contains(result, `"age": 25`) { // Note: no quotes for numbers
		t.Errorf("Result should contain age without quotes: %v", result)
	}

	if !strings.Contains(result, `"city": "New York"`) {
		t.Errorf("Result should contain city: %v", result)
	}

	if !strings.Contains(result, `"timestamp":`) {
		t.Errorf("Result should contain timestamp: %v", result)
	}
}

// TestDollarSyntaxSeparators tests the separator handling for $name syntax
func TestDollarSyntaxSeparators(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		want     string
	}{
		{
			name:     "dot separator",
			template: "$name.txt",
			vars:     map[string]string{"name": "file"},
			want:     "file.txt",
		},
		{
			name:     "underscore in name",
			template: "$name_suffix",
			vars:     map[string]string{"name_suffix": "value"},
			want:     "value",
		},
		{
			name:     "multiple separators",
			template: "$name.ext and $other-file",
			vars:     map[string]string{"name": "test", "other": "another"},
			want:     "test.ext and another-file",
		},
		{
			name:     "slash separator",
			template: "$path/file",
			vars:     map[string]string{"path": "/home/user"},
			want:     "/home/user/file",
		},
		{
			name:     "colon separator",
			template: "$host:$port",
			vars:     map[string]string{"host": "localhost", "port": "8080"},
			want:     "localhost:8080",
		},
		{
			name:     "space separator",
			template: "$first $second",
			vars:     map[string]string{"first": "Hello", "second": "World"},
			want:     "Hello World",
		},
		{
			name:     "comma separator",
			template: "$first,$second",
			vars:     map[string]string{"first": "one", "second": "two"},
			want:     "one,two",
		},
		{
			name:     "semicolon separator",
			template: "$first;$second",
			vars:     map[string]string{"first": "cmd1", "second": "cmd2"},
			want:     "cmd1;cmd2",
		},
		{
			name:     "parentheses separator",
			template: "$func($arg)",
			vars:     map[string]string{"func": "print", "arg": "hello"},
			want:     "print(hello)",
		},
		{
			name:     "brackets separator",
			template: "$array[$index]",
			vars:     map[string]string{"array": "items", "index": "0"},
			want:     "items[0]",
		},
		{
			name:     "underscore vs separator",
			template: "$name_var.ext",
			vars:     map[string]string{"name_var": "test"},
			want:     "test.ext",
		},
		{
			name:     "mixed with brace syntax",
			template: "$name.txt and ${other}.log",
			vars:     map[string]string{"name": "file", "other": "debug"},
			want:     "file.txt and debug.log",
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
				t.Errorf("Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to compare string slices
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
