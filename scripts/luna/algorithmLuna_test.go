package luna

import (
	"testing"
)

const card = 4261022612411900

func TestCalculateLuhn(t *testing.T) {
	type args struct {
		number int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "positive #1 luna",
			args: args{
				card,
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateLuhn(tt.args.number); got != tt.want {
				t.Errorf("CalculateLuhn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checksum(t *testing.T) {
	type args struct {
		number int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "positive #1 luna",
			args: args{
				card,
			},
			want: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checksum(tt.args.number); got != tt.want {
				t.Errorf("checksum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValid(t *testing.T) {
	type args struct {
		number int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "positive #1 luna",
			args: args{
				card,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Valid(tt.args.number); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
