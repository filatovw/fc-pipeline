package main

import (
	"fmt"
	"regexp"
)

var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// validationErr input value validation error
type validationErr struct {
	msg string
}

func (e validationErr) Error() string {
	return e.msg
}

func newValidationErr(msg string, args ...interface{}) validationErr {
	return validationErr{msg: fmt.Sprintf(msg, args...)}
}

func validateEmail(email string) error {
	if !emailRegexp.Match([]byte(email)) {
		return newValidationErr("invalid email %s", email)
	}
	return nil
}

func validateRecord(value []string) error {
	if len(value) != 2 {
		return newValidationErr("unexpected length of a value %v", value)
	}
	if value[0] == "" {
		return newValidationErr("name can not be empty")
	}
	if err := validateEmail(value[1]); err != nil {
		return err
	}
	return nil
}
