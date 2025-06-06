package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trento-project/workbench/pkg/utils"
	"github.com/trento-project/workbench/test/helpers"
)

type SystemReplicationState struct {
	Online              bool   `ini:"online"`
	Mode                string `ini:"mode"`
	OperationMode       string `ini:"operation mode"`
	SiteID              string `ini:"site id"`
	SiteName            string `ini:"site name"`
	IsSource            bool   `ini:"isSource"`
	IsConsumer          bool   `ini:"isConsumer"`
	HasConsumers        bool   `ini:"hasConsumers"`
	IsTakeoverActive    bool   `ini:"isTakeoverActive"`
	IsPrimarySuspended  bool   `ini:"isPrimarySuspended"`
	IsTimetravelEnabled bool   `ini:"isTimetravelEnabled"`
	ReplayMode          string `ini:"replayMode"`
	ActivePrimarySite   string `ini:"active primary site"`
	PrimaryMasters      string `ini:"primary masters"`
}

func TestFindMatches(t *testing.T) {
	tests := []struct {
		name     string
		text     []byte
		expected map[string]any
	}{
		{
			name:     "no matches",
			text:     []byte("no match"),
			expected: map[string]any{},
		},
		{
			name: "single match",
			text: []byte("key=value"),
			expected: map[string]any{
				"key": "value",
			},
		},
		{
			name: "multiple matches",
			text: []byte("key1=value1\nkey2=value2\nkey3=2"),
			expected: map[string]any{
				"key1": "value1",
				"key2": "value2",
				"key3": "2",
			},
		},
		{
			name: "repeated keys",
			text: []byte("key1=value1\nkey2=value2\nkey1=value3"),
			expected: map[string]any{
				"key1": []any{"value1", "value3"},
				"key2": "value2",
			},
		},
		{
			name: "from `hdbnsutil -sr_state -sapcontrol=1` output",
			text: helpers.ReadFixture("hana_secondary_unregistration/stopped_secondary.output"),
			expected: map[string]any{
				"online":              "false",
				"mode":                "sync",
				"operation_mode":      "unknown",
				"site_id":             "2",
				"site_name":           "Site2",
				"isSource":            "unknown",
				"isConsumer":          "true",
				"hasConsumers":        "unknown",
				"isTakeoverActive":    "false",
				"isPrimarySuspended":  "false",
				"isTimetravelEnabled": "false",
				"replayMode":          "auto",
				"active_primary_site": "1",
				"primary_masters":     "node-vmhana01",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FindMatches("(.+)=(.*)", tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}
