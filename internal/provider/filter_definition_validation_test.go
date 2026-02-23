package provider

import (
	"strings"
	"testing"
)

func TestValidateAuditLogFilterDefinition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		definition  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid simple filter",
			definition: `{"filter":{"class":{"name":"connection"}}}`,
			wantErr:    false,
		},
		{
			name: "valid and/or structure with separate condition objects",
			definition: `{
				"filter": {
					"class": [
						{
							"name": "table_access",
							"event": {
								"name": ["insert", "update", "delete"],
								"log": {
									"and": [
										{
											"field": {
												"name": "table_database.str",
												"value": "prod"
											}
										},
										{
											"or": [
												{
													"field": {
														"name": "table_name.str",
														"value": "employee"
													}
												},
												{
													"field": {
														"name": "table_name.str",
														"value": "projects"
													}
												}
											]
										}
									]
								}
							}
						}
					]
				}
			}`,
			wantErr: false,
		},
		{
			name: "invalid condition object with field and or together",
			definition: `{
				"filter": {
					"class": [
						{
							"name": "table_access",
							"event": {
								"name": ["insert", "update", "delete"],
								"log": {
									"and": [
										{
											"field": {
												"name": "table_database.str",
												"value": "prod"
											},
											"or": [
												{
													"field": {
														"name": "table_name.str",
														"value": "employee"
													}
												},
												{
													"field": {
														"name": "table_name.str",
														"value": "projects"
													}
												}
											]
										}
									]
								}
							}
						}
					]
				}
			}`,
			wantErr:     true,
			errContains: "contains multiple logical operators",
		},
		{
			name:        "invalid json syntax",
			definition:  `{"filter":{"class":{"name":"connection"}}`,
			wantErr:     true,
			errContains: "must be valid JSON",
		},
		{
			name:        "missing top-level filter key",
			definition:  `{"class":{"name":"connection"}}`,
			wantErr:     true,
			errContains: "top-level \"filter\" object",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateAuditLogFilterDefinition(tc.definition)
			if tc.wantErr && err == nil {
				t.Fatalf("expected an error but got none")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if tc.wantErr && tc.errContains != "" && err != nil && !strings.Contains(err.Error(), tc.errContains) {
				t.Fatalf("expected error to contain %q, got: %q", tc.errContains, err.Error())
			}
		})
	}
}
