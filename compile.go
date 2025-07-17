package var_template

import (
	"fmt"
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
// accepted options:  %d, *, +, :file, :bash, :shell_quote
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
	// New directive fields
	isFile       bool // has :file suffix
	isBash       bool // has :bash suffix
	isShellQuote bool // has :shell_quote suffix
	open         int  // begin of ${
	close        int  // position of }
	index        int  // $'s position in the string (global unique)
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

// findNextDollarVar finds the next $name pattern in the string
// Returns -1 if no valid $name pattern is found
func findNextDollarVar(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '$' {
			// Check if this is a ${ pattern (skip it)
			if i+1 < len(s) && s[i+1] == '{' {
				continue
			}
			// Check if this is a valid $name pattern
			if i+1 < len(s) && isValidVarStart(s[i+1]) {
				return i
			}
		}
	}
	return -1
}

// extractDollarVarName extracts the variable name from a $name pattern
// Returns the variable name and the end position (exclusive)
func extractDollarVarName(s string) (string, int) {
	if len(s) == 0 || s[0] != '$' {
		return "", 0
	}

	if len(s) == 1 {
		return "", 0
	}

	// Skip the $
	i := 1

	// Check if first character is valid for variable name
	if !isValidVarStart(s[i]) {
		return "", 0
	}

	// Find the end of the variable name
	start := i

	// Handle macro case: $@timestamp
	if s[i] == '@' {
		i++ // Skip the @
		// Continue with normal variable name characters
		for i < len(s) && isValidVarChar(s[i]) {
			i++
		}
	} else {
		// Normal variable name
		for i < len(s) && isValidVarChar(s[i]) {
			i++
		}
	}

	// Handle separator logic: $name.s -> ${name}.s,  $name_s -> ${name_s}
	// If we hit a separator that's not underscore, stop
	if i < len(s) && !isValidVarChar(s[i]) && s[i] != '_' {
		// This is a separator, variable name ends here
		return s[start:i], i
	}

	return s[start:i], i
}

// isValidVarStart checks if a character is valid for starting a variable name
func isValidVarStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == '@'
}

// isValidVarChar checks if a character is valid within a variable name
func isValidVarChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func Compile(template string) *Template {
	// find all variables and positions
	var positions []*varAndPosition
	varMap := make(map[string]bool)
	s := template
	i := 0
	index := 0

	for s != "" {
		// Look for both ${} and $ patterns
		braceOpenIdx := strings.Index(s, open)
		dollarIdx := findNextDollarVar(s)

		// Determine which pattern comes first
		var nextIdx int
		var isBracePattern bool

		if braceOpenIdx >= 0 && dollarIdx >= 0 {
			if braceOpenIdx < dollarIdx {
				nextIdx = braceOpenIdx
				isBracePattern = true
			} else {
				nextIdx = dollarIdx
				isBracePattern = false
			}
		} else if braceOpenIdx >= 0 {
			nextIdx = braceOpenIdx
			isBracePattern = true
		} else if dollarIdx >= 0 {
			nextIdx = dollarIdx
			isBracePattern = false
		} else {
			// no more variables
			break
		}

		// Check for escaping
		if nextIdx > 0 && s[nextIdx-1] == '\\' {
			i += nextIdx + 1
			s = s[nextIdx+1:]
			continue
		}

		var v *varAndPosition
		var endIdx int

		if isBracePattern {
			// Handle ${name} pattern
			openIdxEnd := nextIdx + len(open)
			closeIdx := strings.Index(s[openIdxEnd:], close)
			if closeIdx < 0 {
				i += openIdxEnd
				s = s[openIdxEnd:]
				continue
			}
			closeIdx += openIdxEnd
			varName := strings.TrimSpace(s[openIdxEnd:closeIdx])

			v = parseVarName(varName)
			if v.varName == "" {
				i += closeIdx + len(close)
				s = s[closeIdx+len(close):]
				continue
			}

			v.open = i + nextIdx
			v.close = i + closeIdx
			endIdx = closeIdx + len(close)
		} else {
			// Handle $name pattern
			varName, varEnd := extractDollarVarName(s[nextIdx:])
			if varName == "" {
				i += nextIdx + 1
				s = s[nextIdx+1:]
				continue
			}

			v = parseVarName(varName)
			if v.varName == "" {
				i += nextIdx + 1
				s = s[nextIdx+1:]
				continue
			}

			v.open = i + nextIdx
			v.close = i + nextIdx + varEnd - 1
			endIdx = nextIdx + varEnd
		}

		varMap[v.varName] = true
		index++
		v.index = index
		positions = append(positions, v)
		i += endIdx
		s = s[endIdx:]
	}

	// Post-process to handle escaped sequences and adjust positions
	processedTemplate, adjustedPositions := processEscapesAndAdjustPositions(template, positions)

	return &Template{
		template:     processedTemplate,
		varPositions: adjustedPositions,
		vars:         getVars(varMap),
	}
}

// processEscapesAndAdjustPositions removes backslashes from escaped variable patterns
// and adjusts variable positions accordingly
func processEscapesAndAdjustPositions(template string, positions []*varAndPosition) (string, []*varAndPosition) {
	result := template
	adjustedPositions := make([]*varAndPosition, len(positions))

	// Copy positions
	for i, pos := range positions {
		adjustedPositions[i] = pos.clone()
	}

	// Process escapes and adjust positions
	adjustment := 0
	for i := 0; i < len(result); i++ {
		if i > 0 && result[i-1] == '\\' && (result[i] == '$') {
			// Remove the backslash
			result = result[:i-1] + result[i:]
			i-- // Adjust index after removal
			adjustment++

			// Adjust all positions that come after this escape
			for _, pos := range adjustedPositions {
				if pos.open > i {
					pos.open--
				}
				if pos.close > i {
					pos.close--
				}
			}
		}
	}

	return result, adjustedPositions
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
	var isFile bool
	var isBash bool
	var isShellQuote bool

	// Handle macro prefix
	if strings.HasPrefix(varName, "@") {
		isMacro = true
		actualVarName = varName // Keep the @ prefix for macros
	} else {
		// Parse using the new approach
		var err error
		actualVarName, required, hasDefaultValue, defaultValue, isNumber, repMode, isFile, isBash, isShellQuote, err = parseVariableDefinition(varName)
		if err != nil {
			// Return an empty varAndPosition for invalid variables
			return &varAndPosition{
				raw:     varName,
				varName: "",
			}
		}
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
		// New directive fields
		isFile:       isFile,
		isBash:       isBash,
		isShellQuote: isShellQuote,
	}
}

// parseVariableDefinition parses a variable definition using the new approach
func parseVariableDefinition(varName string) (name string, required bool, hasDefault bool, defaultVal string, isNumber bool, repMode repeatMode, isFile bool, isBash bool, isShellQuote bool, err error) {
	repMode = repeatMode_Same

	// Special handling for bash directive - check if it ends with :bash
	if strings.HasSuffix(varName, ":bash") {
		// Check for multiple directives first
		beforeBash := varName[:len(varName)-5] // Remove ":bash"

		// For bash directive, the variable name is the command (everything before :bash)
		name = beforeBash
		isBash = true
		return
	}
	if strings.HasSuffix(varName, ":file") {
		// Check for multiple directives first
		beforeFile := varName[:len(varName)-5] // Remove ":file"
		name = beforeFile
		isFile = true
		return
	}

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

		// Check for multiple directives (should be an error)
		if strings.Contains(remainder, ":") {
			return "", false, false, "", false, repeatMode_Same, false, false, false, fmt.Errorf("multiple directives not allowed: %s", remainder)
		}

		// Check for directives
		if remainder == "%d" {
			isNumber = true
		} else if remainder == "+" {
			repMode = repeatMode_Uniq
		} else if remainder == "*" {
			repMode = repeatMode_Any
		} else if remainder == "shell_quote" {
			isShellQuote = true
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
				if next == "%d" || next == "+" || next == "*" || next == "file" || next == "bash" || next == "shell_quote" {
					// This is a directive marker
					return remainder[:i], remainder[i:]
				}
			}
		}
	}
	// No directive found, the entire remainder is the default value
	return remainder, ""
}
