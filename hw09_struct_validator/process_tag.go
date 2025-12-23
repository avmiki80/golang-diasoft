package hw09structvalidator

import (
	"reflect"
	"strings"
)

type FieldTags struct {
	JSONTags     map[string]string
	ValidateTags map[string]ValidateRule
	OtherTags    map[string]string
}

type ValidateRule struct {
	Name  string
	Value string
	Args  []string
}

func NewValidateTag(field reflect.StructField) map[string]ValidateRule {
	fieldTags := FieldTags{
		JSONTags:     make(map[string]string),
		ValidateTags: make(map[string]ValidateRule),
		OtherTags:    make(map[string]string),
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		fieldTags.JSONTags["name"] = parts[0]
		for _, opt := range parts[1:] {
			fieldTags.JSONTags[opt] = "true"
		}
	}

	validateTag := field.Tag.Get("validate")
	if validateTag != "" {
		rules := parseValidateTag(validateTag)
		fieldTags.ValidateTags = rules
	}

	for _, tagName := range []string{"xml", "yaml", "bson", "db"} {
		if tagValue := field.Tag.Get(tagName); tagValue != "" {
			fieldTags.OtherTags[tagName] = tagValue
		}
	}
	return fieldTags.ValidateTags
}

func parseValidateTag(tag string) map[string]ValidateRule {
	rules := make(map[string]ValidateRule)

	ruleStrings := strings.Split(tag, "|")

	for _, ruleStr := range ruleStrings {
		parts := strings.SplitN(ruleStr, ":", 2)
		rule := ValidateRule{
			Name: parts[0],
		}

		if len(parts) > 1 {
			rule.Value = parts[1]
			rule.Args = strings.Split(parts[1], ",")
		}

		rules[rule.Name] = rule
	}

	return rules
}
