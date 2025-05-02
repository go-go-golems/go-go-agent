package agent

import (
	"testing"
)

func TestXMLStructuredDataExtractor_Extract(t *testing.T) {
	tests := []struct {
		name         string
		tagName      string
		response     string
		expectedData string
	}{
		{
			name:         "Basic extraction",
			tagName:      "results",
			response:     "Some text <results>Data inside</results> More text",
			expectedData: "Data inside",
		},
		{
			name:         "Tag not found",
			tagName:      "missing",
			response:     "Some text <results>Data inside</results> More text",
			expectedData: "",
		},
		{
			name:         "Start tag found, end tag missing",
			tagName:      "results",
			response:     "Some text <results>Data inside...",
			expectedData: "",
		},
		{
			name:         "Empty content between tags",
			tagName:      "empty",
			response:     "<empty></empty>",
			expectedData: "",
		},
		{
			name:         "Content with leading/trailing whitespace",
			tagName:      "whitespace",
			response:     "<whitespace>  \n Data with space \t </whitespace>",
			expectedData: "Data with space",
		},
		{
			name:         "Multiple tags, extract first occurrence",
			tagName:      "data",
			response:     "<data>Content 1</data><data>Content 2</data>",
			expectedData: "Content 1",
		},
		{
			name:         "Self-closing style tag (should not extract)",
			tagName:      "self",
			response:     "<self />",
			expectedData: "",
		},
		{
			name:         "Nested tags (extract outer content)",
			tagName:      "outer",
			response:     "<outer>Outer start <inner>Inner data</inner> Outer end</outer>",
			expectedData: "Outer start <inner>Inner data</inner> Outer end",
		},
		{
			name:         "Deeply nested target tags",
			tagName:      "target",
			response:     "<target>Level 1 <target>Level 2</target> Level 1 end</target>",
			expectedData: "Level 1 <target>Level 2</target> Level 1 end",
		},
		{
			name:         "Nested different tags",
			tagName:      "target",
			response:     "<target>Content <other>Other stuff</other> More content</target>",
			expectedData: "Content <other>Other stuff</other> More content",
		},
		{
			name:         "Tag with attributes",
			tagName:      "tagWithAttr",
			response:     `Before <tagWithAttr id="123" flag>Attribute content</tagWithAttr> After`,
			expectedData: "Attribute content",
		},
		{
			name:         "Tag with attributes and whitespace",
			tagName:      "tagWithAttrSpaced",
			response:     `Before <tagWithAttrSpaced  id = "123" >Spaced attribute content</tagWithAttrSpaced> After`,
			expectedData: "Spaced attribute content",
		},
		{
			name:         "Similar tag names (prefix)",
			tagName:      "tag",
			response:     "<tag>Correct</tag><tagSuffix>Wrong</tagSuffix>",
			expectedData: "Correct",
		},
		{
			name:         "Similar tag names (suffix) - should be ignored",
			tagName:      "tagSuffix",
			response:     "<tag>Wrong</tag><tagSuffix>Correct</tagSuffix>",
			expectedData: "Correct",
		},
		{
			name:         "Malformed XML (missing closing tag)",
			tagName:      "malformed",
			response:     "<malformed>This data starts...",
			expectedData: "",
		},
		{
			name:         "Malformed XML (missing closing tag nested)",
			tagName:      "outer",
			response:     "<outer>Start <inner>incomplete</outer>",
			expectedData: "",
		},
		{
			name:         "Malformed XML (wrong closing tag)",
			tagName:      "outer",
			response:     "<outer>Content</inner>",
			expectedData: "",
		},
		{
			name:         "No content",
			tagName:      "any",
			response:     "",
			expectedData: "",
		},
		{
			name:         "Special characters in content",
			tagName:      "special",
			response:     "<special>Data with < > & \" ' characters</special>",
			expectedData: "Data with < > & \" ' characters",
		},
		{
			name:         "Tag name with namespace (should be treated as whole name)",
			tagName:      "ns:tag",
			response:     "<ns:tag>Namespace content</ns:tag> <tag>Regular</tag>",
			expectedData: "Namespace content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewXMLStructuredDataExtractor(tt.tagName)
			actualData := extractor.Extract(tt.response)
			if actualData != tt.expectedData {
				t.Errorf("Extract() got = %q, want %q", actualData, tt.expectedData)
			}
		})
	}
}

func TestExtractMultipleData(t *testing.T) {
	tests := []struct {
		name          string
		tagNames      []string
		response      string
		expectedDatas map[string]string
	}{
		{
			name:     "Extract multiple existing tags",
			tagNames: []string{"data1", "data2"},
			response: "<data1>Content 1</data1>Some text<data2>Content 2</data2>",
			expectedDatas: map[string]string{
				"data1": "Content 1",
				"data2": "Content 2",
			},
		},
		{
			name:     "Extract some existing, some missing tags",
			tagNames: []string{"data1", "missing", "data2"},
			response: "<data1>Content 1</data1><data2>Content 2</data2>",
			expectedDatas: map[string]string{
				"data1": "Content 1",
				"data2": "Content 2",
			},
		},
		{
			name:          "Extract only non-existing tags",
			tagNames:      []string{"missing1", "missing2"},
			response:      "<data1>Content 1</data1>",
			expectedDatas: map[string]string{},
		},
		{
			name:     "Extract tags with overlapping content (extracts independently based on first occurrence)",
			tagNames: []string{"outer", "inner"},
			response: "<outer><inner>Inner data</inner></outer>",
			expectedDatas: map[string]string{
				"outer": "<inner>Inner data</inner>",
				"inner": "Inner data",
			},
		},
		{
			name:     "Extract tags with attributes",
			tagNames: []string{"data1", "data2"},
			response: `<data1 id="1">Content 1</data1> <data2 name='foo'>Content 2</data2>`, // Added attributes
			expectedDatas: map[string]string{
				"data1": "Content 1",
				"data2": "Content 2",
			},
		},
		{
			name:          "Empty response string",
			tagNames:      []string{"data1", "data2"},
			response:      "",
			expectedDatas: map[string]string{},
		},
		{
			name:          "No tag names provided",
			tagNames:      []string{},
			response:      "<data1>Content 1</data1>",
			expectedDatas: map[string]string{},
		},
		{
			name:     "Multiple occurrences of the same tag (extracts first)",
			tagNames: []string{"repeat"},
			response: "<repeat>First</repeat><repeat>Second</repeat>",
			expectedDatas: map[string]string{
				"repeat": "First",
			},
		},
		{
			name:     "Multiple tags, one malformed",
			tagNames: []string{"good", "bad"},
			response: "<good>Good Content</good><bad>Bad content....",
			expectedDatas: map[string]string{
				"good": "Good Content", // Bad tag shouldn't prevent good tag extraction
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualDatas := ExtractMultipleData(tt.response, tt.tagNames)

			// Check if the lengths are the same
			if len(actualDatas) != len(tt.expectedDatas) {
				t.Errorf("ExtractMultipleData() map length got = %d, want %d\nActual: %v\nExpected: %v", len(actualDatas), len(tt.expectedDatas), actualDatas, tt.expectedDatas)
			}

			// Check if all expected keys and values are present
			for key, expectedValue := range tt.expectedDatas {
				actualValue, ok := actualDatas[key]
				if !ok {
					t.Errorf("ExtractMultipleData() missing expected key: %q", key)
				} else if actualValue != expectedValue {
					t.Errorf("ExtractMultipleData() for key %q got = %q, want %q", key, actualValue, expectedValue)
				}
			}

			// Optional: Check if there are any unexpected keys in the actual map
			for key := range actualDatas {
				if _, ok := tt.expectedDatas[key]; !ok {
					t.Errorf("ExtractMultipleData() got unexpected key: %q", key)
				}
			}
		})
	}
}
