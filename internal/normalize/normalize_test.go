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

func TestContextFromFile(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/home/user/.kube/configs/kubeconfig_john@prod-payments.yaml", "john@prod-payments"},
		{"/home/user/.kube/configs/kubeconfig_JOHN@PROD.yaml", "john@prod"},
		{"kubeconfig_john@prod.yaml", "john@prod"},
		// Pas de @ → retourne juste le nom normalisé
		{"kubeconfig_mycluster.yaml", "mycluster"},
		// Préfixe absent
		{"monfichier.yaml", "monfichier"},
	}
	for _, tt := range tests {
		got := normalize.ContextFromFile(tt.path)
		if got != tt.want {
			t.Errorf("ContextFromFile(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestIsValidSourceFilename(t *testing.T) {
	tests := []struct {
		path  string
		valid bool
	}{
		{"kubeconfig_john@prod.yaml", true},
		{"/abs/path/kubeconfig_john@prod-payments.yaml", true},
		// @ manquant
		{"kubeconfig_mycluster.yaml", false},
		// user vide (@cluster)
		{"kubeconfig_@prod.yaml", false},
		// cluster vide (user@)
		{"kubeconfig_john@.yaml", false},
		// préfixe absent
		{"john@prod.yaml", false},
		{"config.yaml", false},
		// double @ — valide : user=a, cluster=b@c
		{"kubeconfig_a@b@c.yaml", true},
	}
	for _, tt := range tests {
		got := normalize.IsValidSourceFilename(tt.path)
		if got != tt.valid {
			t.Errorf("IsValidSourceFilename(%q) = %v, want %v", tt.path, got, tt.valid)
		}
	}
}
