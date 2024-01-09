package gohome

import "testing"

func TestExpand(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name       string
		args       args
		want       string
		mustExpand bool
	}{
		{
			name: "Expand returns same path when no ~ exists",
			args: args{
				path: "/users/me/something",
			},
			want:       "/users/me/something",
			mustExpand: false,
		},
		{
			name: "Expand returns empty path when empty path provided",
			args: args{
				path: "",
			},
			want:       "",
			mustExpand: false,
		},
		{
			name: "Expand returns same path when ~ exists in the middle of the given path",
			args: args{
				path: "/users/~/something",
			},
			want:       "/users/~/something",
			mustExpand: false,
		},
		{
			name: "Expand returns expanded path when ~ exists in prefix",
			args: args{
				path: "~/something",
			},
			want:       "",
			mustExpand: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Expand(tt.args.path)
			if !tt.mustExpand {
				if got != tt.want {
					t.Errorf("Expand() = %v, want %v", got, tt.want)
				}
			} else {
				if got == tt.want {
					t.Errorf("Expand() = %v, don't want %v", got, tt.want)
				}
			}
		})
	}
}
