package normalize_test

import (
	"testing"

	"github.com/UnPoilTefal/kmgr/internal/normalize"
)

func TestName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"john", "john"},
		{"John", "john"},
		{"PROD-PAYMENTS", "prod-payments"},
		{"prod_payments", "prod_payments"},
		{"prod payments", "prod-payments"},
		{"prod@cluster", "prod@cluster"},
		{"prod/cluster", "prod/cluster"},
		{"héllo", "h-llo"},
		{"", ""},
	}
	for _, tt := range tests {
		got := normalize.Name(tt.input)
		if got != tt.want {
			t.Errorf("Name(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		user, cluster string
		want          string
	}{
		{"john", "prod-payments", "john@prod-payments"},
		{"JOHN", "PROD", "john@prod"},
		{"john doe", "prod cluster", "john-doe@prod-cluster"},
	}
	for _, tt := range tests {
		got := normalize.New(tt.user, tt.cluster).String()
		if got != tt.want {
			t.Errorf("New(%q, %q) = %q, want %q", tt.user, tt.cluster, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{"john@prod-payments", "john@prod-payments", true},
		{"JOHN@PROD", "JOHN@PROD", true}, // Parse never sanitizes.
		{"mycluster", "", false},
		{"@prod", "", false},
		{"john@", "", false},
		{"a@b@c", "a@b@c", true}, // double "@" is still parseable.
	}
	for _, tt := range tests {
		got, ok := normalize.Parse(tt.input)
		if ok != tt.ok || got.String() != tt.want {
			t.Errorf("Parse(%q) = (%q, %v), want (%q, %v)", tt.input, got, ok, tt.want, tt.ok)
		}
	}
}

func TestFromFilename(t *testing.T) {
	tests := []struct {
		path string
		want string
		ok   bool
	}{
		{"/home/user/.kube/configs/kubeconfig_john@prod-payments.yaml", "john@prod-payments", true},
		{"/home/user/.kube/configs/kubeconfig_JOHN@PROD.yaml", "john@prod", true},
		{"kubeconfig_john@prod.yaml", "john@prod", true},
		{"kubeconfig_a@b@c.yaml", "a@b@c", true}, // double "@": user=a, cluster=b@c
		{"kubeconfig_mycluster.yaml", "", false}, // no "@"
		{"kubeconfig_@prod.yaml", "", false},     // empty user
		{"kubeconfig_john@.yaml", "", false},     // empty cluster
		{"monfichier.yaml", "", false},           // missing prefix
	}
	for _, tt := range tests {
		got, ok := normalize.FromFilename(tt.path)
		if ok != tt.ok || got.String() != tt.want {
			t.Errorf("FromFilename(%q) = (%q, %v), want (%q, %v)", tt.path, got, ok, tt.want, tt.ok)
		}
	}
}

func TestContextNameAccessors(t *testing.T) {
	c := normalize.New("john", "prod-payments")
	if got := c.String(); got != "john@prod-payments" {
		t.Errorf("String() = %q", got)
	}
	if got := c.Cluster(); got != "prod-payments" {
		t.Errorf("Cluster() = %q", got)
	}
	if got := c.AuthInfo(); got != "john@prod-payments" {
		t.Errorf("AuthInfo() = %q", got)
	}
	if got := c.SourceFilename(); got != "kubeconfig_john@prod-payments.yaml" {
		t.Errorf("SourceFilename() = %q", got)
	}
}
