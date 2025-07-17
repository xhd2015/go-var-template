package var_template

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Template struct {
	template     string
	varPositions []*varAndPosition
	vars         []string
}

func (c *Template) HasVariables() bool {
	return len(c.vars) > 0
}
func (c *Template) HasNonMacroVariables() bool {
	for _, vr := range c.varPositions {
		if !vr.isMacro {
			return true
		}
	}
	return false
}

func (c *Template) GetGetRaw(varName string) string {
	for _, position := range c.varPositions {
		if position.varName == varName {
			return position.raw
		}
	}
	return ""
}

func (c *Template) UpdateVars(newVars []string) {
	c.vars = newVars
}

func (c *Template) NumVars() int {
	return len(c.varPositions)
}

func (c *Template) Var(i int) Var {
	return c.varPositions[i]
}

func (c *Template) Variables() []string {
	return c.vars
}

// get current template
func (c *Template) Template() string {
	return c.template
}
func (c *Template) String() string {
	return c.template
}

func (c *Template) PartialApply(vars map[string]string) *Template {
	if len(vars) == 0 {
		return c
	}
	t, err := c.apply(vars, false, false, false)
	if err != nil {
		// un expected
		panic(err)
	}
	return t
}

type ApplyOptions struct {
	ApplyDefault     bool
	ApplyMacro       bool
	ValidateRequired bool
}

func (c *Template) Apply(vars map[string]string, opts *ApplyOptions) *Template {
	if len(vars) == 0 && !opts.ApplyDefault && !opts.ApplyMacro {
		return c
	}
	t, err := c.apply(vars, opts.ValidateRequired, opts.ApplyDefault, opts.ApplyMacro)
	if err != nil {
		// un expected
		panic(err)
	}
	return t
}

func (c *Template) apply(vars map[string]string, validateRequired bool, applyDefault bool, applyMacro bool) (*Template, error) {
	if len(c.vars) == 0 && !applyDefault && !applyMacro {
		return c, nil
	}
	s := c.template
	var b strings.Builder
	b.Grow(len(s))
	oldIdx := 0

	var missingVarPositions []*varAndPosition
	missingVarMap := make(map[string]bool)
	// each varPosition represents its prefix upto its close
	// the last varPosition may have trailing suffix
	for j, vr := range c.varPositions {
		var val string
		var ok bool

		if vr.isFile {
			// also use varname as file directly
			if data, err := os.ReadFile(vr.varName); err == nil {
				val = string(data)
				ok = true
			} else {
				return nil, fmt.Errorf("failed to read file %s: %v", vr.varName, err)
			}
		} else if vr.isBash {
			// Execute bash command using variable name
			cmd := exec.Command("bash", "-c", vr.varName)
			if output, err := cmd.Output(); err == nil {
				val = strings.TrimRight(string(output), "\n\r")
				ok = true
			} else {
				return nil, fmt.Errorf("failed to execute bash command %s: %v", vr.varName, err)
			}
		} else if vr.isMacro {
			if applyMacro {
				macro := vr.varName
				if strings.HasPrefix(macro, "@") {
					macro = macro[1:] // Remove @ prefix
				}
				if macro == "timestamp" {
					val = strconv.FormatInt(time.Now().Unix(), 10)
					ok = true
				} else if macro == "timestamp_ms" {
					val = strconv.FormatInt(unixMilli(time.Now()), 10)
					ok = true
				} else if macro == "timestamp_us" {
					val = strconv.FormatInt(unixMicro(time.Now()), 10)
					ok = true
				} else if macro == "timestamp_ns" {
					val = strconv.FormatInt(time.Now().UnixNano(), 10)
					ok = true
				}
			}
		} else {
			val, ok = vars[vr.varName]
		}

		// Calculate the end position of the variable
		var varEndPos int
		if isDollarSyntax(s, vr.open) {
			// $name syntax - end position is already calculated correctly
			varEndPos = vr.close + 1
		} else {
			// ${name} syntax - add closing brace length
			varEndPos = vr.close + len(close)
		}

		if !ok {
			if applyDefault && !vr.isMacro && vr.hasDefaultValue {
				val = vr.defaultValue
				ok = true // Mark as ok so directives can be applied
			} else {
				if validateRequired && vr.required {
					return nil, fmt.Errorf("required variable %s is missing", vr.raw)
				}
				cpVar := vr.clone()
				cpVar.open = b.Len() + (vr.open - oldIdx)
				cpVar.close = b.Len() + (vr.close - oldIdx)
				missingVarPositions = append(missingVarPositions, cpVar)
				missingVarMap[vr.varName] = true
				b.WriteString(s[oldIdx:varEndPos])
				oldIdx = varEndPos
				continue
			}
		}

		// Process other directives if value is found (from variables or default)
		if ok && val != "" && !vr.isBash && !vr.isFile {
			if vr.isShellQuote {
				// Shell quote the value
				val = quoteShellStr(val)
			}
		}

		if vr.isNumber &&
			isChar(s, vr.open-1, '"') &&
			isChar(s, varEndPos, '"') &&
			(j == 0 || !c.varPositions[j-1].isNumber || vr.open-1 > getVarEndPos(s, c.varPositions[j-1])) /*does not cross with previous var's ending*/ {
			// trim quotes
			b.WriteString(s[oldIdx : vr.open-1])
			b.WriteString(val)
			oldIdx = varEndPos + 1 /*len of "*/
		} else {
			b.WriteString(s[oldIdx:vr.open])
			b.WriteString(val)
			oldIdx = varEndPos
		}
	}
	// last
	b.WriteString(s[oldIdx:])

	return &Template{
		template:     b.String(),
		varPositions: missingVarPositions,
		vars:         getVars(missingVarMap),
	}, nil
}

func quoteShellStr(s string) string {
	if s == "" {
		return "''"
	}
	if strings.ContainsAny(s, "\t \n;<>\\${}()&!*") { // special args
		s = strings.ReplaceAll(s, "'", "'\\''")
		return "'" + s + "'"
	}
	return s
}

// isDollarSyntax checks if a variable at the given position uses $name syntax
func isDollarSyntax(s string, pos int) bool {
	return pos < len(s) && s[pos] == '$' && (pos+1 >= len(s) || s[pos+1] != '{')
}

// getVarEndPos calculates the end position of a variable
func getVarEndPos(s string, vr *varAndPosition) int {
	if isDollarSyntax(s, vr.open) {
		return vr.close + 1
	} else {
		return vr.close + len(close)
	}
}

func isChar(s string, idx int, ch byte) bool {
	return idx >= 0 && idx < len(s) && s[idx] == ch
}

// Execute will format the value, apply defaults and validate required variables
func (c *Template) Execute(vars map[string]string) (string, error) {
	t, err := c.apply(vars, true, true, true)
	if err != nil {
		return "", err
	}
	return t.template, nil
}

// stable sorted
func getVars(varMap map[string]bool) []string {
	vars := make([]string, 0, len(varMap))
	for v := range varMap {
		vars = append(vars, v)
	}
	sort.Strings(vars)
	return vars
}
