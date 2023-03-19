package repo

import (
	"context"
	"fmt"
	"github.com/SETTER2000/gofermart/config"
	"github.com/SETTER2000/gofermart/internal/entity"
	"sync"
	"testing"
)

func TestInMemory_Put(t *testing.T) {
	type fields struct {
		lock *sync.Mutex
		m    map[string]entity.Gofermarts
		cfg  *config.Config
	}
	type args struct {
		ctx context.Context
		sh  *entity.Gofermart
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		countUID   int
		countShort int
		UID        string
		url        string
		del        bool
	}{
		{
			name: "positive test #1",
			args: args{
				ctx: context.Background(),
				sh:  &entity.Gofermart{Slug: "qwerty1", UserID: "uid-1", URL: "https://example1.com", Del: false},
			},
			countUID:   1,
			countShort: 1,
			UID:        "uid-1",
			url:        "https://example1.com",
			del:        false,
		}, {
			name: "positive test #2",
			args: args{
				ctx: context.Background(),
				sh:  &entity.Gofermart{Slug: "qwerty2", UserID: "uid-2", URL: "https://example2.com", Del: false},
			},
			countUID:   2,
			countShort: 1,
			UID:        "uid-2",
			url:        "https://example2.com",
			del:        false,
		}, {
			name: "positive test #3",
			args: args{
				ctx: context.Background(),
				sh:  &entity.Gofermart{Slug: "qwerty3", UserID: "uid-2", URL: "https://example3.com", Del: false},
			},
			countUID:   2,
			countShort: 2,
			UID:        "uid-2",
			url:        "https://example3.com",
			del:        false,
		}, {
			name: "positive test #4",
			args: args{
				ctx: context.Background(),
				sh:  &entity.Gofermart{Slug: "qwerty4", UserID: "uid-2", URL: "https://example4.com", Del: false},
			},
			countUID:   2,
			countShort: 3,
			UID:        "uid-2",
			url:        "https://example4.com",
			del:        false,
		}, {
			name: "update Del and URL #5",
			args: args{
				ctx: context.Background(),
				sh:  &entity.Gofermart{Slug: "qwerty4", UserID: "uid-2", URL: "https://example5.com", Del: true},
			},
			countUID:   2,
			countShort: 3,
			UID:        "uid-2",
			url:        "https://example5.com",
			del:        true,
		},
	}
	s := &InMemory{
		m: make(map[string]entity.Gofermarts),
	}
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := s.Put(tt.args.ctx, tt.args.sh); err != nil {
				t.Errorf("Put() error = %v, countUID %v", err, nil)
			}
			cUID := len(s.m)
			if cUID != tt.countUID {
				t.Errorf("Len memory = %v, countUID %v\n", cUID, tt.countUID)
			}

			cSh := len(s.m[tt.UID])
			if cSh != tt.countShort {
				t.Errorf("Len shorts; expected: %v, actual: %v\n", tt.countShort, cSh)
				fmt.Printf("List Gofermarts: %v\n", s.m[tt.UID])
			}
			shorts := s.m[tt.UID]
			if idx == 4 && tt.url != shorts[2].URL {
				t.Errorf("Short URL; expected: %v, actual: %v\n", tt.url, shorts[2].URL)
			}
			if idx == 4 && tt.del != shorts[2].Del {
				t.Errorf("Short Del; expected: %v, actual: %v\n", tt.del, shorts[2].Del)
			}
		})
	}
}
