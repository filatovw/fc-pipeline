package main

import (
	"fmt"
	"testing"
)

// TODO: write more tests OR
// TODO: use external email validator

func TestValidateEmail(t *testing.T) {
	testData := []struct {
		input string
		err   bool
	}{
		{
			input: "lala@gmail.com",
			err:   false,
		},
		{
			input: "lala@com",
			err:   false,
		},
		{
			input: "123456789@493403.ww",
			err:   false,
		},
		{
			input: "EMAIL+TO+ME@mail.run",
			err:   false,
		},
		{
			input: "email+to+me@mail.run",
			err:   false,
		},
		{
			input: "email+to.m-.-e@mail.run",
			err:   false,
		},
		{
			input: "email+to.m-.-e@mail-me.run",
			err:   false,
		},
		{
			input: "",
			err:   true,
		},
		{
			input: "emz",
			err:   true,
		},
		{
			input: "@dd",
			err:   true,
		},
		{
			input: "@dd.com",
			err:   true,
		},
		{
			input: "em@",
			err:   true,
		},
		{
			input: "email@aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.com",
			err:   true,
		},
	}

	for i, td := range testData {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := validateEmail(td.input)
			if err != nil {
				if !td.err {
					t.Fatalf("unexpected error %s", err)
				}
			}
			if td.err && err == nil {
				t.Fatal("expected error")
			}
		})
	}

}
