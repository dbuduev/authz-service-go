package sphinx

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"sort"
	"testing"
)

func TestBranchGroupContent_Reverse(t *testing.T) {
	trans := cmp.Transformer("Sort", func(in []uuid.UUID) []uuid.UUID {
		out := append([]uuid.UUID(nil), in...) // Copy input to avoid mutating it
		sort.Slice(out, func(i, j int) bool {
			for k := 0; k < len(out[i]); k++ {
				if out[i][k] != out[j][k] {
					return out[i][k] < out[j][k]
				}
			}
			return false
		})
		return out
	})
	var g = [...]uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	var b = [...]uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}

	tests := []struct {
		name string
		m    BranchGroupContent
		want BranchGroupsOfBranch
	}{
		{
			name: "Simple case",
			m: map[uuid.UUID][]uuid.UUID{
				g[0]: {b[0], b[1], b[2]},
				g[1]: {b[0], b[3]},
			},
			want: map[uuid.UUID][]uuid.UUID{
				b[0]: {g[0], g[1]},
				b[1]: {g[0]},
				b[2]: {g[0]},
				b[3]: {g[1]},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.Reverse()
			if diff := cmp.Diff(tt.want, got, trans); diff != "" {
				t.Errorf("Reverse1() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
