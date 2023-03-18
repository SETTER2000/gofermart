package scripts

import (
	"github.com/SETTER2000/gofermart/config"
	"regexp"
	"testing"
)

func TestCheckEnvironFlag(t *testing.T) {
	type args struct {
		environName string
		flagName    string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "positive test #1",
			args: args{
				environName: "environName",
				flagName:    "flagName",
			},
			want: true,
		}, {
			name: "positive test #2 flag variable set",
			args: args{
				environName: "",
				flagName:    "flagName",
			},
			want: true,
		}, {
			name: "positive test #3 same naming of environment variable and flag",
			args: args{
				environName: "environName",
				flagName:    "DATABASE_DSN",
			},
			want: true,
		}, {
			name: "negative test #1",
			args: args{
				environName: "",
				flagName:    "",
			},
			want: false,
		}, {
			name: "negative test #2",
			args: args{
				environName: "environName",
				flagName:    "",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckEnvironFlag(tt.args.environName, tt.args.flagName); got != tt.want {
				t.Errorf("CheckEnvironFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHost(t *testing.T) {
	type args struct {
		cfg       config.HTTP
		gofermart string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				cfg:       config.HTTP{BaseURL: "http://localhost:8080"},
				gofermart: "wEuWothteri_t23",
			},
			want: "http://localhost:8080/wEuWothteri_t23",
		}, {
			name: "negative test #1 extra closing slash",
			args: args{
				cfg:       config.HTTP{BaseURL: "http://localhost:8080/"},
				gofermart: "wEuWothteri_t23",
			},
			want: "http://localhost:8080//wEuWothteri_t23",
		}, {
			name: "negative test #2 missing protocol",
			args: args{
				cfg:       config.HTTP{BaseURL: "localhost:8080"},
				gofermart: "wEuWothteri_t23",
			},
			want: "localhost:8080/wEuWothteri_t23",
		}, {
			name: "negative test #3 missing protocol and port",
			args: args{
				cfg:       config.HTTP{BaseURL: "localhost"},
				gofermart: "wEuWothteri_t23",
			},
			want: "localhost/wEuWothteri_t23",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHost(tt.args.cfg, tt.args.gofermart); got != tt.want {
				t.Errorf("GetHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateString(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name    string
		args    args
		pattern string
		want    bool
	}{
		{
			name:    "positive test #1",
			args:    args{n: 10},
			pattern: `[_0-9a-zA-Z]`,
			want:    true,
		}, {
			name:    "positive test #2",
			args:    args{n: 1},
			pattern: `[_0-9a-zA-Z]{3}`,
			want:    true,
		}, {
			name:    "positive test #3",
			args:    args{n: 5},
			pattern: `[_0-9a-zA-Z]{5}`,
			want:    true,
		}, {
			name:    "positive test #4",
			args:    args{n: 0},
			pattern: `[_0-9a-zA-Z]{3}`,
			want:    true,
		}, {
			name:    "negative test #1",
			args:    args{n: -1},
			pattern: `[_0-9a-zA-Z]`,
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateString(tt.args.n)
			if match, _ := regexp.MatchString(tt.pattern, got); !match {
				t.Errorf("GenerateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrim(t *testing.T) {
	type args struct {
		s string
		t string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "positive test #1",
			args:    args{"\nstring home\n", ""},
			want:    "string home",
			wantErr: false,
		},
		{
			name:    "positive test #2",
			args:    args{"\nstring home:", ":"},
			want:    "string home",
			wantErr: false,
		},
		{
			name:    "positive test #3",
			args:    args{"\nstring home\n", ":"},
			want:    "string home",
			wantErr: false,
		},
		{
			name:    "positive test #4",
			args:    args{"string home", ":"},
			want:    "string home",
			wantErr: false,
		},
		{
			name:    "positive test #5",
			args:    args{"string home", "s"},
			want:    "tring home",
			wantErr: false,
		},
		{
			name:    "positive test #5",
			args:    args{"string home", ""},
			want:    "string home",
			wantErr: false,
		},
		{
			name:    "negative test #1",
			args:    args{"", ""},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Trim(tt.args.s, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("Trim() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Trim() got = %v, want %v", got, tt.want)
			}
		})
	}
}
