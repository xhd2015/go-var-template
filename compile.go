package template

import (
	"strings"
)

const open = "${"
const close = "}"

type repeatMode int

const (
	repeatMode_Same repeatMode = 0
	repeatMode_Any  repeatMode = 1
	repeatMode_Uniq repeatMode = 2
)

// example: ${a},  ${ a.b }
// ${ a! } --> a is required
// ${a!:%d} -> a is typeof number, and is required
// ${ a ?:10} --> default 10
// valid combinations:
//
//	${a:uniq}
//
// separators:  !, ?:, :,
// accepted options:  %d, *, +
type varAndPosition struct {
	// the original raw string
	raw             string
	varName         string
	varInitContent  string
	isNumber        bool       // has :%d suffix
	repeatMode      repeatMode // :+, :*
	hasDefaultValue bool
	defaultValue    string // has ?:something
	required        bool   // has ! suffix
	isMacro         bool
	open            int // begin of ${
	close           int // position of }
	index           int // $在整个字符串中出现在第几个位置（全局唯一）
}

func (c *varAndPosition) clone() *varAndPosition {
	v := *c
	return &v
}

func (c *varAndPosition) String() string {
	return c.raw
}

func (c *varAndPosition) Name() string {
	return c.varName
}

func (c *varAndPosition) Required() bool {
	return c.required
}

func (c *varAndPosition) HasDefault() bool {
	return c.hasDefaultValue
}

func (c *varAndPosition) IsMacro() bool {
	return c.isMacro
}
func (c *varAndPosition) IsNumber() bool {
	return c.isNumber
}

var _ Var = (*varAndPosition)(nil)

type Var interface {
	Name() string
	Required() bool
	HasDefault() bool
	IsMacro() bool
	IsNumber() bool
}

func Compile(template string) *Template {
	// find all variables and positions
	var positions []*varAndPosition
	varMap := make(map[string]bool)
	s := template
	i := 0
	index := 0
	for s != "" {
		openIdx := strings.Index(s, open)
		if openIdx < 0 {
			// no more
			break
		}
		// escaped
		if openIdx > 0 && s[openIdx-1] == '\\' {
			i += openIdx + len(open)
			s = s[openIdx+len(open):]
			continue
		}

		// next begin
		openIdxEnd := openIdx + len(open)
		closeIdx := strings.Index(s[openIdxEnd:], close)
		if closeIdx < 0 {
			break
		}
		closeIdx += openIdxEnd
		varName := strings.TrimSpace(s[openIdxEnd:closeIdx])

		v := parseVarName(varName)
		if v.varName == "" {
			i += closeIdx + len(close)
			s = s[closeIdx+len(close):]
			continue
		}

		varMap[v.varName] = true
		v.open = i + openIdx
		v.close = i + closeIdx
		index++
		v.index = index
		positions = append(positions, v)
		i += closeIdx + len(close)
		s = s[closeIdx+len(close):]
	}

	return &Template{
		template:     template,
		varPositions: positions,
		vars:         getVars(varMap),
	}
}

func parseVarName(varName string) *varAndPosition {
	// New approach: split by delimiters and recognize directives
	var actualVarName string
	var required bool
	var isNumber bool
	var hasDefaultValue bool
	var defaultValue string
	var repMode = repeatMode_Same
	var isMacro bool

	// Handle macro prefix
	if strings.HasPrefix(varName, "@") {
		isMacro = true
		actualVarName = varName // Keep the @ prefix for macros
	} else {
		// Parse using the new approach
		actualVarName, required, hasDefaultValue, defaultValue, isNumber, repMode = parseVariableDefinition(varName)
	}

	return &varAndPosition{
		raw:             varName,
		varName:         strings.TrimSpace(actualVarName),
		isNumber:        isNumber,
		repeatMode:      repMode,
		hasDefaultValue: hasDefaultValue,
		defaultValue:    defaultValue,
		required:        required,
		isMacro:         isMacro,
	}
}

// parseVariableDefinition parses a variable definition using the new approach
func parseVariableDefinition(varName string) (name string, required bool, hasDefault bool, defaultVal string, isNumber bool, repMode repeatMode) {
	repMode = repeatMode_Same

	// Step 1: Find the variable name (everything before the first ?: or :)
	var nameEnd int
	if idx := strings.Index(varName, "?:"); idx != -1 {
		nameEnd = idx
		hasDefault = true
	} else if idx := strings.Index(varName, ":"); idx != -1 {
		nameEnd = idx
	} else {
		nameEnd = len(varName)
	}

	// Extract variable name and check for required flag
	namePart := varName[:nameEnd]
	name, required = parseVariableNameAndRequired(namePart)

	// Step 2: Process the rest of the string
	remainder := varName[nameEnd:]

	if hasDefault {
		// We have a default value, extract it
		remainder = remainder[2:] // Skip "?:"
		defaultVal, remainder = extractDefaultValue(remainder)
	}

	// Step 3: Process any remaining directives
	if remainder != "" && strings.HasPrefix(remainder, ":") {
		remainder = remainder[1:] // Skip ":"

		// Check for directives
		if remainder == "%d" {
			isNumber = true
		} else if remainder == "+" {
			repMode = repeatMode_Uniq
		} else if remainder == "*" {
			repMode = repeatMode_Any
		}
	}

	return
}

// parseVariableNameAndRequired extracts variable name and required flag, handling invalid characters
func parseVariableNameAndRequired(segment string) (string, bool) {
	segment = strings.TrimSpace(segment)

	// Find the actual variable name (alphanumeric + underscore)
	var nameBytes []byte
	var foundRequired bool

	for i, r := range segment {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			nameBytes = append(nameBytes, segment[i])
		} else if r == '!' {
			foundRequired = true
			// Stop processing after finding the required flag
			break
		} else {
			// Invalid character, stop processing
			break
		}
	}

	return string(nameBytes), foundRequired
}

// extractDefaultValue extracts the default value from the remainder, stopping at directive markers
func extractDefaultValue(remainder string) (defaultVal string, remaining string) {
	// Look for the next directive marker
	for i := 0; i < len(remainder); i++ {
		if remainder[i] == ':' {
			// Check if this is followed by a directive
			if i+1 < len(remainder) {
				next := remainder[i+1:]
				if next == "%d" || next == "+" || next == "*" {
					// This is a directive marker
					return remainder[:i], remainder[i:]
				}
			}
		}
	}
	// No directive found, the entire remainder is the default value
	return remainder, ""
}
