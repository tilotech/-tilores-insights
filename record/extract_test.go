package record_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tilotech/tilores-insights/helpers"
	"github.com/tilotech/tilores-insights/record"
	api "github.com/tilotech/tilores-plugin-api"
)

func TestExtract(t *testing.T) {
	dataJSON := `
	{
		"value": "string",
		"nested": {
			"value": "nested string value",
			"super": {
				"value": "Super Nested String Value"
			}
		},
		"int": 123,
		"list": [
			"abc",
			"DEF",
			"geh"
		],
		"nullValue": null,
		"emptyString": ""
	}
	`
	data := map[string]any{}
	err := json.Unmarshal([]byte(dataJSON), &data)
	require.NoError(t, err)

	r := &api.Record{
		ID:   "some-id",
		Data: data,
	}

	cases := map[string]struct {
		useCustomData bool
		customData    *api.Record
		expected      any
	}{
		"value": {
			expected: "string",
		},
		"nested.value": {
			expected: "nested string value",
		},
		"nested.super.value": {
			expected: "Super Nested String Value",
		},
		"nested": {
			expected: map[string]any{
				"value": "nested string value",
				"super": map[string]any{
					"value": "Super Nested String Value",
				},
			},
		},
		"int": {
			expected: 123.0, // json parses numbers as float64!
		},
		"list": {
			expected: []any{
				"abc",
				"DEF",
				"geh",
			},
		},
		"list.0": {
			expected: "abc",
		},
		"list.1": {
			expected: "DEF",
		},
		"list.2": {
			expected: "geh",
		},
		"nil as input": {
			useCustomData: true,
			customData:    &api.Record{},
			expected:      nil,
		},
		"nil record as input": {
			useCustomData: true,
			customData:    nil,
			expected:      nil,
		},
		"nonexistent": {
			expected: nil,
		},
		"emptyString": {
			expected: "",
		},
		"nullValue": {
			expected: nil,
		},
		"non.existent": {
			expected: nil,
		},
		"nested.nonexistent": {
			expected: nil,
		},
		"nested.value.nonexistent": {
			expected: nil,
		},
		"int.nonexistent": {
			expected: nil,
		},
		"list.a": {
			expected: nil,
		},
		"list.4": {
			expected: nil,
		},
		"list.-1": {
			expected: nil,
		},
	}

	for path, c := range cases {
		t.Run(path, func(t *testing.T) {
			input := r
			if c.useCustomData {
				input = c.customData
			}
			actual := record.Extract(input, path)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestExtractNumber(t *testing.T) {
	dataJSON := `
	{
		"nonnumeric": "string",
		"int": 123,
		"float": 123.4,
		"nullValue": null,
		"numericText": "123",
		"exponent": "1e3",
		"nested": {
			"value": "123"
		}
	}
	`
	data := map[string]any{}
	err := json.Unmarshal([]byte(dataJSON), &data)
	require.NoError(t, err)

	r := &api.Record{
		ID:   "some-id",
		Data: data,
	}

	cases := map[string]struct {
		expected    any
		expectError bool
	}{
		"nonnumeric": {
			expectError: true,
		},
		"int": {
			expected: helpers.NullifyFloat(123.0),
		},
		"float": {
			expected: helpers.NullifyFloat(123.4),
		},
		"nullValue": {
			expected: nil,
		},
		"numericText": {
			expected: helpers.NullifyFloat(123.0),
		},
		"exponent": {
			expected: helpers.NullifyFloat(1000.0),
		},
		"nested": {
			expectError: true,
		},
	}

	for path, c := range cases {
		t.Run(path, func(t *testing.T) {
			actual, err := record.ExtractNumber(r, path)
			if c.expectError {
				assert.Error(t, err)
			} else if c.expected == nil {
				assert.Nil(t, actual)
			} else {
				assert.Equal(t, c.expected, actual)
			}
		})
	}
}

func TestExtractString(t *testing.T) {
	dataJSON := `
	{
		"nested": {
			"value": "nested string value",
			"super": {
				"value": "Super Nested String Value"
			}
		},
		"list": [
			"abc",
			"DEF",
			"geh"
		],
		"keepUpper": "Has Upper Case",
		"caseInsensitive": "Has Upper Case",
    "bool": true,
		"emptyString": "",
		"int": 123,
		"float": 123.4,
		"nullValue": null,
		"numericText": "123",
		"exponent": "1e3",
		"nested": {
			"propB": "valB",
			"propA": "valA"
		}
	}
	`
	data := map[string]any{}
	err := json.Unmarshal([]byte(dataJSON), &data)
	require.NoError(t, err)

	r := &api.Record{
		ID:   "some-id",
		Data: data,
	}

	cases := map[string]struct {
		expected      *string
		caseSensitive bool
	}{
		"bool": {
			expected: helpers.NullifyString("true"),
		},
		"int": {
			expected: helpers.NullifyString("123"),
		},
		"float": {
			expected: helpers.NullifyString("123.4"),
		},
		"nullValue": {
			expected: nil,
		},
		"numericText": {
			expected: helpers.NullifyString("123"),
		},
		"exponent": {
			expected: helpers.NullifyString("1e3"),
		},
		"nested": {
			caseSensitive: true,
			expected:      helpers.NullifyString(`{"propA":"valA","propB":"valB"}`),
		},
		"keepUpper": {
			caseSensitive: true,
			expected:      helpers.NullifyString("Has Upper Case"),
		},
		"caseInsensitive": {
			expected: helpers.NullifyString("has upper case"),
		},
		"list.0": {
			expected: helpers.NullifyString("abc"),
		},
		"list.1": {
			expected: helpers.NullifyString("def"),
		},
		"list.2": {
			expected: helpers.NullifyString("geh"),
		},
		"list": {
			expected: helpers.NullifyString(`["abc","def","geh"]`),
		},
		"emptyString": {
			expected: helpers.NullifyString(""),
		},
	}

	for path, c := range cases {
		t.Run(path, func(t *testing.T) {
			actual, err := record.ExtractString(r, path, c.caseSensitive)
			assert.NoError(t, err)
			if c.expected == nil {
				assert.Nil(t, actual)
			} else {
				require.NotNil(t, actual)
				assert.Equal(t, *c.expected, *actual)
			}
		})
	}
}
