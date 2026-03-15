package cmd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hwuu/codeup-control/internal/client"
	"github.com/hwuu/codeup-control/internal/config"
)

func TestResolveRepoRef(t *testing.T) {
	cfg := &config.Config{DefaultRepo: "default-org/default-repo"}

	repoRef, err := resolveRepoRef(cfg, "input-org/input-repo")
	if err != nil {
		t.Fatalf("resolveRepoRef with arg returned error: %v", err)
	}
	if repoRef != "input-org/input-repo" {
		t.Fatalf("resolveRepoRef with arg = %q, want input-org/input-repo", repoRef)
	}

	repoRef, err = resolveRepoRef(cfg, "")
	if err != nil {
		t.Fatalf("resolveRepoRef with default returned error: %v", err)
	}
	if repoRef != "default-org/default-repo" {
		t.Fatalf("resolveRepoRef with default = %q, want default-org/default-repo", repoRef)
	}

	_, err = resolveRepoRef(&config.Config{}, "")
	if err == nil {
		t.Fatal("resolveRepoRef without arg or default should fail")
	}
}

func TestResolveRepoProjectID(t *testing.T) {
	const (
		orgID   = "org-123"
		repoRef = "demo-group/demo-repo"
	)

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := fmt.Sprintf("/oapi/v1/codeup/organizations/%s/repositories/%s", orgID, repoRef)
		if r.URL.Path != wantPath {
			t.Fatalf("request path = %q, want %q", r.URL.Path, wantPath)
		}
		if token := r.Header.Get("x-yunxiao-token"); token != "test-token" {
			t.Fatalf("token header = %q, want test-token", token)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"name":"demo","path":"demo-repo","pathWithNamespace":"demo-group/demo-repo"}`))
	}))
	defer server.Close()

	domain := strings.TrimPrefix(server.URL, "https://")
	c := client.New(domain, "test-token", false)
	c.HTTP = server.Client()
	if transport, ok := c.HTTP.Transport.(*http.Transport); ok {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	cfg := &config.Config{OrganizationID: orgID}
	gotRepoRef, gotProjectID, err := resolveRepoProjectID(c, cfg, repoRef)
	if err != nil {
		t.Fatalf("resolveRepoProjectID returned error: %v", err)
	}
	if gotRepoRef != repoRef {
		t.Fatalf("repoRef = %q, want %q", gotRepoRef, repoRef)
	}
	if gotProjectID != "42" {
		t.Fatalf("projectID = %q, want 42", gotProjectID)
	}
}
