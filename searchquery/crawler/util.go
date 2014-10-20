package main

import (
	"strings"
)

func StringInMap(m map[string]bool, s string) bool {

	for k, v := range m {

		if strings.EqualFold(k, s) {

			return true
		}
	}

	return false
}

func MimeEnabled(m map[string]bool, s string) bool {

	for k, v := range m {

		if (strings.EqualFold(k, s)) && v {

			return true
		}
	}

	return false
}
