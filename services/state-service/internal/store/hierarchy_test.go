package store

import "testing"

func TestInstitutionDepthFromChain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		chain   []string
		want    int
		wantErr bool
	}{
		{name: "root only", chain: []string{}, want: 1},
		{name: "subsidiary", chain: []string{"parent"}, want: 2},
		{name: "sub subsidiary", chain: []string{"parent", "grandparent"}, want: 3},
		{name: "too deep", chain: []string{"a", "b", "c"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InstitutionDepthFromChain(tt.chain)
			if tt.wantErr {
				if err != ErrHierarchyDepth {
					t.Fatalf("expected ErrHierarchyDepth, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("depth = %d, want %d", got, tt.want)
			}
		})
	}
}
