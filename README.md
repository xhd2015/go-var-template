# Variable Template

A powerful Go template library for string interpolation with support for default values, type hints, and built-in macros.

## Features

- **Variable Substitution**: Replace `${variable}` placeholders with actual values
- **Default Values**: Specify default values with `${variable?:default}`
- **Required Variables**: Mark variables as required with `${variable!}`
- **Type Hints**: Specify number types with `${variable:%d}` for automatic quote removal
- **Repeat Modes**: Control variable uniqueness with `${variable:+}` (unique) or `${variable:*}` (any)
- **Built-in Macros**: Use `${@timestamp}`, `${@timestamp_ms}`, `${@timestamp_us}`, `${@timestamp_ns}`
- **Robust Parsing**: Handles complex default values including URLs and special characters

## Installation

```bash
go get github.com/xhd2015/less-gen/var_template
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/xhd2015/less-gen/var_template"
)

func main() {
    // Basic usage
    tmpl := template.Compile("Hello ${name}")
    result, err := tmpl.Execute(map[string]string{"name": "World"})
    if err != nil {
        panic(err)
    }
    fmt.Println(result) // Output: Hello World
}
```

## Syntax Reference

### Basic Variables

```go
// Simple variable
template.Compile("Hello ${name}")

// Variable with spaces (trimmed automatically)
template.Compile("Hello ${ name }")
```

### Required Variables

```go
// Required variable (will error if not provided)
template.Compile("Hello ${name!}")
```

### Default Values

```go
// Simple default
template.Compile("Hello ${name?:World}")

// Default with special characters
template.Compile("URL: ${url?:http://localhost:8080}")

// Default with complex values
template.Compile("Config: ${config?:key=value&other=123}")

// Empty default
template.Compile("Value: ${value?:}")
```

### Type Hints

```go
// Number type - removes quotes in JSON contexts
template.Compile(`{"age": "${age:%d}"}`)
// With age="25" produces: {"age": 25}
```

### Repeat Modes

```go
// Unique mode - variable should be unique
template.Compile("Items: ${items:+}")

// Any mode - variable can repeat
template.Compile("Items: ${items:*}")
```

### Built-in Macros

```go
// Unix timestamp (seconds)
template.Compile("Time: ${@timestamp}")

// Unix timestamp (milliseconds)
template.Compile("Time: ${@timestamp_ms}")

// Unix timestamp (microseconds)
template.Compile("Time: ${@timestamp_us}")

// Unix timestamp (nanoseconds)
template.Compile("Time: ${@timestamp_ns}")
```

### Complex Combinations

```go
// Required variable with default and type hint
template.Compile("Age: ${age!?:25:%d}")

// All features combined
template.Compile(`{
    "name": "${name!}",
    "age": "${age?:25:%d}",
    "url": "${url?:http://localhost:8080}",
    "timestamp": "${@timestamp}"
}`)
```

## API Reference

### Template Creation

```go
// Compile a template string
tmpl := template.Compile("Hello ${name}")

// Get template information
vars := tmpl.Variables()        // []string - list of variable names
hasVars := tmpl.HasVariables()  // bool - true if template has variables
numVars := tmpl.NumVars()       // int - number of variable positions
```

### Template Execution

```go
// Execute with all variables
result, err := tmpl.Execute(map[string]string{
    "name": "World",
    "age":  "25",
})

// Partial application (some variables remain)
partial := tmpl.PartialApply(map[string]string{
    "name": "World",
    // age is not provided, remains as ${age}
})

// Apply with options
result := tmpl.Apply(vars, &template.ApplyOptions{
    ApplyDefault:     true,  // Apply default values
    ApplyMacro:       true,  // Process macros
    ValidateRequired: true,  // Validate required variables
})
```

### Variable Information

```go
// Get variable details
for i := 0; i < tmpl.NumVars(); i++ {
    v := tmpl.Var(i)
    fmt.Printf("Name: %s\n", v.Name())
    fmt.Printf("Required: %v\n", v.Required())
    fmt.Printf("Has Default: %v\n", v.HasDefault())
    fmt.Printf("Is Macro: %v\n", v.IsMacro())
    fmt.Printf("Is Number: %v\n", v.IsNumber())
}
```

## Examples

### Configuration Template

```go
configTemplate := `{
    "host": "${host?:localhost}",
    "port": "${port?:8080:%d}",
    "ssl": "${ssl?:false}",
    "timeout": "${timeout?:30:%d}",
    "created_at": "${@timestamp}"
}`

tmpl := template.Compile(configTemplate)
config, err := tmpl.Execute(map[string]string{
    "host": "api.example.com",
    "port": "443",
    "ssl":  "true",
})
```

### URL Builder

```go
urlTemplate := "${scheme?:https}://${host!}:${port?:80:%d}${path?:/}"

tmpl := template.Compile(urlTemplate)
url, err := tmpl.Execute(map[string]string{
    "host": "api.example.com",
    "port": "443",
    "path": "/v1/users",
})
// Result: https://api.example.com:443/v1/users
```

### SQL Query Template

```go
queryTemplate := `SELECT * FROM ${table!} 
WHERE ${column?:id} = ${value!} 
LIMIT ${limit?:10:%d}`

tmpl := template.Compile(queryTemplate)
query, err := tmpl.Execute(map[string]string{
    "table": "users",
    "value": "123",
    "limit": "5",
})
```

## Error Handling

```go
tmpl := template.Compile("Hello ${name!}")

// This will return an error because 'name' is required
result, err := tmpl.Execute(map[string]string{})
if err != nil {
    fmt.Printf("Error: %v\n", err)
    // Output: Error: required variable name! is missing
}
```

## Best Practices

1. **Use descriptive variable names**: `${database_host}` instead of `${host}`
2. **Provide sensible defaults**: `${port?:8080:%d}` for common configurations
3. **Mark critical variables as required**: `${api_key!}` for essential values
4. **Use type hints for numbers**: `${count:%d}` to avoid quote issues in JSON
5. **Leverage macros for timestamps**: `${@timestamp}` instead of manual time handling

## Performance

The library is designed for high performance:
- Templates are compiled once and can be reused
- Variable parsing uses an efficient algorithm
- Memory allocations are minimized during execution

```go
// Compile once, use many times
tmpl := template.Compile("Hello ${name}")

// Fast execution
for i := 0; i < 1000; i++ {
    result, _ := tmpl.Execute(map[string]string{"name": "World"})
}
```

## License

This package is part of the less-gen project. 