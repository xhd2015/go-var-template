package template

import (
	"testing"
)

func BenchmarkCompile(b *testing.B) {
	template := "Hello ${name}, you are ${age:%d} years old and live in ${city?:Unknown}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compile(template)
	}
}

func BenchmarkTemplateExecute(b *testing.B) {
	template := "Hello ${name}, you are ${age:%d} years old and live in ${city?:Unknown}"
	tmpl := Compile(template)
	vars := map[string]string{
		"name": "John",
		"age":  "25",
		"city": "New York",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpl.Execute(vars)
	}
}

func BenchmarkTemplatePartialApply(b *testing.B) {
	template := "Hello ${name}, you are ${age:%d} years old and live in ${city?:Unknown}"
	tmpl := Compile(template)
	vars := map[string]string{
		"name": "John",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpl.PartialApply(vars)
	}
}

func BenchmarkParseVarName(b *testing.B) {
	varName := "name!?:John:%d"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseVarName(varName)
	}
}

func BenchmarkMacroExecution(b *testing.B) {
	template := "Time: ${@timestamp}, ${@timestamp_ms}, ${@timestamp_us}, ${@timestamp_ns}"
	tmpl := Compile(template)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpl.Execute(map[string]string{})
	}
}
