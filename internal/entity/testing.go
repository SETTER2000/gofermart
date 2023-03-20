package entity

import "testing"

func TestAuthentication(t *testing.T) *Authentication {
	return &Authentication{
		Login:    "clare",
		Password: "123",
	}
}
