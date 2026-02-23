package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type auditLogFilterDefinitionValidator struct{}

var _ validator.String = auditLogFilterDefinitionValidator{}

func (v auditLogFilterDefinitionValidator) Description(context.Context) string {
	return "must be valid JSON and follow MySQL audit log filter logical condition structure"
}

func (v auditLogFilterDefinitionValidator) MarkdownDescription(context.Context) string {
	return "must be valid JSON and follow MySQL audit log filter logical condition structure"
}

func (v auditLogFilterDefinitionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if err := validateAuditLogFilterDefinition(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid JSON Definition",
			err.Error(),
		)
	}
}

func validateAuditLogFilterDefinition(definition string) error {
	var parsed any

	if err := json.Unmarshal([]byte(definition), &parsed); err != nil {
		return fmt.Errorf("the filter definition must be valid JSON: %w", err)
	}

	rootObject, ok := parsed.(map[string]any)
	if !ok {
		return fmt.Errorf("the filter definition must be a JSON object")
	}

	filterValue, exists := rootObject["filter"]
	if !exists {
		return fmt.Errorf("the filter definition must include a top-level \"filter\" object")
	}

	if _, ok := filterValue.(map[string]any); !ok {
		return fmt.Errorf("the \"filter\" value must be a JSON object")
	}

	if err := validateConditionTree(parsed, "$"); err != nil {
		return fmt.Errorf("the filter definition must follow MySQL logical condition structure: %w", err)
	}

	return nil
}

func validateConditionTree(value any, path string) error {
	switch typed := value.(type) {
	case map[string]any:
		return validateConditionObject(typed, path)
	case []any:
		for i, item := range typed {
			if err := validateConditionTree(item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateConditionObject(object map[string]any, path string) error {
	operatorKeys := []string{"and", "or", "not", "field"}
	var present []string

	for _, key := range operatorKeys {
		if _, ok := object[key]; ok {
			present = append(present, key)
		}
	}

	if len(present) > 0 {
		if len(present) != 1 {
			return fmt.Errorf("%s contains multiple logical operators (%s); each condition object must contain exactly one of and, or, not, field",
				path, strings.Join(present, ", "))
		}

		operator := present[0]
		operatorValue := object[operator]

		switch operator {
		case "and", "or":
			expressions, ok := operatorValue.([]any)
			if !ok {
				return fmt.Errorf("%s.%s must be an array of condition objects", path, operator)
			}
			if len(expressions) == 0 {
				return fmt.Errorf("%s.%s must contain at least one condition object", path, operator)
			}
			for i, expression := range expressions {
				if _, ok := expression.(map[string]any); !ok {
					return fmt.Errorf("%s.%s[%d] must be a condition object", path, operator, i)
				}
				if err := validateConditionTree(expression, fmt.Sprintf("%s.%s[%d]", path, operator, i)); err != nil {
					return err
				}
			}
			return nil
		case "not":
			nested, ok := operatorValue.(map[string]any)
			if !ok {
				return fmt.Errorf("%s.not must be a condition object", path)
			}
			return validateConditionTree(nested, fmt.Sprintf("%s.not", path))
		case "field":
			if _, ok := operatorValue.(map[string]any); !ok {
				return fmt.Errorf("%s.field must be an object", path)
			}
			return nil
		}
	}

	for key, child := range object {
		if err := validateConditionTree(child, fmt.Sprintf("%s.%s", path, key)); err != nil {
			return err
		}
	}

	return nil
}
