package thinkgo

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Validator provides ThinkPHP-style validation.
// Supports both struct tag validation and explicit rule maps.
//
// Struct tag usage:
//
//	type LoginRequest struct {
//	    Username string `validate:"required|minLen:3|maxLen:20"`
//	    Password string `validate:"required|minLen:6"`
//	    Email    string `validate:"required|email"`
//	}
//
//	req := LoginRequest{Username: "a"}
//	v := thinkgo.NewValidator()
//	if !v.ValidateStruct(req) {
//	    fmt.Println(v.Errors())  // map[Password:... Username:... Email:...]
//	}
type Validator struct {
	errors map[string]string
}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{
		errors: make(map[string]string),
	}
}

// Rule is a single validation rule.
type Rule struct {
	Field   string
	Name    string // rule name: required, minLen, maxLen, email, numeric, etc.
	Param   string // parameter for the rule (e.g., "6" for minLen:6)
	Message string // custom error message
}

// ValidateRules validates a map of fields against rules.
// This is the explicit rule-map style (like ThinkPHP).
//
//	v := thinkgo.NewValidator()
//	rules := map[string]string{
//	    "username": "required|minLen:3|maxLen:20",
//	    "email":    "required|email",
//	}
//	if !v.ValidateRules(data, rules) {
//	    fmt.Println(v.Errors())
//	}
func (v *Validator) ValidateRules(data map[string]string, rules map[string]string) bool {
	v.errors = make(map[string]string)
	valid := true

	for field, ruleStr := range rules {
		ruleList := parseRules(ruleStr)
		value := data[field]

		for _, rule := range ruleList {
			if !v.applyRule(field, value, rule) {
				valid = false
			}
		}
	}

	return valid
}

// ValidateStruct validates a struct using `validate` struct tags.
// Supports nested structs (recursive validation).
//
// Validates all fields with `validate` tag automatically.
func (v *Validator) ValidateStruct(obj any) bool {
	v.errors = make(map[string]string)
	return v.validateValue(reflect.ValueOf(obj), "")
}

// validateValue recursively validates a reflected value.
func (v *Validator) validateValue(val reflect.Value, prefix string) bool {
	// Unwrap pointer
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return true
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return true
	}

	valid := true
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		if prefix != "" {
			fieldName = prefix + "." + field.Name
		}

		// Recurse into nested structs
		if fieldVal.Kind() == reflect.Struct || fieldVal.Kind() == reflect.Ptr {
			if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
				// nil pointer — still validate if there's a tag
			} else {
				if !v.validateValue(fieldVal, fieldName) {
					valid = false
				}
			}
		}

		// Check validate tag
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		ruleList := parseRules(tag)
		strVal := fmt.Sprintf("%v", fieldVal.Interface())

		for _, rule := range ruleList {
			if !v.applyRule(fieldName, strVal, rule) {
				valid = false
			}
		}
	}

	return valid
}

// Errors returns all validation errors.
func (v *Validator) Errors() map[string]string {
	return v.errors
}

// Error returns the first error message.
func (v *Validator) Error() string {
	for _, msg := range v.errors {
		return msg
	}
	return ""
}

// applyRule applies a single rule to a value.
func (v *Validator) applyRule(field, value string, rule Rule) bool {
	switch rule.Name {
	case "required":
		if strings.TrimSpace(value) == "" {
			v.errors[field] = v.msg(rule, field+" is required")
			return false
		}
	case "minLen":
		if len(value) < toInt(rule.Param) {
			v.errors[field] = v.msg(rule, field+" must be at least "+rule.Param+" characters")
			return false
		}
	case "maxLen":
		if len(value) > toInt(rule.Param) {
			v.errors[field] = v.msg(rule, field+" must not exceed "+rule.Param+" characters")
			return false
		}
	case "len":
		if len(value) != toInt(rule.Param) {
			v.errors[field] = v.msg(rule, field+" must be exactly "+rule.Param+" characters")
			return false
		}
	case "min":
		if toFloat(value) < toFloat(rule.Param) {
			v.errors[field] = v.msg(rule, field+" must be at least "+rule.Param)
			return false
		}
	case "max":
		if toFloat(value) > toFloat(rule.Param) {
			v.errors[field] = v.msg(rule, field+" must not exceed "+rule.Param)
			return false
		}
	case "email":
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, value)
		if !matched {
			v.errors[field] = v.msg(rule, field+" must be a valid email")
			return false
		}
	case "numeric":
		matched, _ := regexp.MatchString(`^\d+(\.\d+)?$`, value)
		if !matched {
			v.errors[field] = v.msg(rule, field+" must be numeric")
			return false
		}
	case "integer":
		matched, _ := regexp.MatchString(`^\d+$`, value)
		if !matched {
			v.errors[field] = v.msg(rule, field+" must be an integer")
			return false
		}
	case "alpha":
		matched, _ := regexp.MatchString(`^[a-zA-Z]+$`, value)
		if !matched {
			v.errors[field] = v.msg(rule, field+" must contain only letters")
			return false
		}
	case "alphaNum":
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, value)
		if !matched {
			v.errors[field] = v.msg(rule, field+" must contain only letters and numbers")
			return false
		}
	case "phone":
		matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, value)
		if !matched {
			v.errors[field] = v.msg(rule, field+" must be a valid phone number")
			return false
		}
	case "url":
		matched, _ := regexp.MatchString(`^https?://`, value)
		if !matched {
			v.errors[field] = v.msg(rule, field+" must be a valid URL")
			return false
		}
	case "in":
		allowed := strings.Split(rule.Param, ",")
		inList := false
		for _, a := range allowed {
			if strings.TrimSpace(a) == value {
				inList = true
				break
			}
		}
		if !inList {
			v.errors[field] = v.msg(rule, field+" must be one of: "+rule.Param)
			return false
		}
	case "regex":
		matched, err := regexp.MatchString(rule.Param, value)
		if err != nil || !matched {
			v.errors[field] = v.msg(rule, field+" format is invalid")
			return false
		}
	}
	return true
}

// msg returns the custom error message or a default.
func (v *Validator) msg(rule Rule, defaultMsg string) string {
	if rule.Message != "" {
		return rule.Message
	}
	return defaultMsg
}

// parseRules parses "required|minLen:3|maxLen:20" into Rule objects.
func parseRules(s string) []Rule {
	var rules []Rule
	parts := strings.Split(s, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		rule := Rule{Name: part}
		if idx := strings.Index(part, ":"); idx >= 0 {
			rule.Name = part[:idx]
			rule.Param = part[idx+1:]
		}
		rules = append(rules, rule)
	}
	return rules
}

// toInt converts a string to int, returning 0 on error.
func toInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// toFloat converts a string to float64, returning 0 on error.
func toFloat(s string) float64 {
	var n float64
	fmt.Sscanf(s, "%f", &n)
	return n
}
