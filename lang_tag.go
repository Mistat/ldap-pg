package ldap_pg

import (
	"errors"
	"fmt"
	"strings"
)

var langs = []string {
	"ja",
}

func langContains(e string) bool {
	for _, a := range langs {
		if a == e {
			return true
		}
	}
	return false
}

func ParseLanguageTag(name string) (string, string, error) {
	as := strings.Split(name, ";")
	if len(as) == 1 {
		return name, "", nil
	}
	if len(as) == 2 {
		if strings.Index(as[1], "lang-") != 0 {
			return name, "", errors.New(fmt.Sprintf("error: LangError : %v", as[1]))
		}
		if !langContains(as[1][5:]) {
			return name, "", errors.New(fmt.Sprintf("error: LangError2 : %v", as[1]))
		}
		return as[0], as[1], nil
	}
	return name, "", errors.New("LangError")
}
