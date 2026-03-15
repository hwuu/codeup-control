package client

import (
	"encoding/json"
	"testing"
)

func TestRepositoryDetailAccessLevelUnmarshal(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload string
		want    AccessLevel
	}{
		{
			name:    "numeric access level",
			payload: `{"id":42,"pathWithNamespace":"demo-group/demo-repo","accessLevel":30}`,
			want:    30,
		},
		{
			name:    "string access level",
			payload: `{"id":42,"pathWithNamespace":"demo-group/demo-repo","accessLevel":"30"}`,
			want:    30,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var repo RepositoryDetail
			if err := json.Unmarshal([]byte(tc.payload), &repo); err != nil {
				t.Fatalf("json.Unmarshal returned error: %v", err)
			}
			if repo.AccessLevel != tc.want {
				t.Fatalf("AccessLevel = %d, want %d", repo.AccessLevel, tc.want)
			}
		})
	}
}

func TestRepositoryListAccessLevelUnmarshal(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload string
		want    AccessLevel
	}{
		{
			name:    "numeric access level",
			payload: `[{"id":42,"pathWithNamespace":"demo-group/demo-repo","accessLevel":40}]`,
			want:    40,
		},
		{
			name:    "string access level",
			payload: `[{"id":42,"pathWithNamespace":"demo-group/demo-repo","accessLevel":"40"}]`,
			want:    40,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var repos []Repository
			if err := json.Unmarshal([]byte(tc.payload), &repos); err != nil {
				t.Fatalf("json.Unmarshal returned error: %v", err)
			}
			if len(repos) != 1 {
				t.Fatalf("len(repos) = %d, want 1", len(repos))
			}
			if repos[0].AccessLevel != tc.want {
				t.Fatalf("AccessLevel = %d, want %d", repos[0].AccessLevel, tc.want)
			}
		})
	}
}
