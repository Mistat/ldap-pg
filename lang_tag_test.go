package ldap_pg

import "testing"

func TestParseLanguageTag(t *testing.T) {

	name, lang, err := ParseLanguageTag("cn")
	if err != nil {
		t.Fatal(err)
	}
	if name != "cn" {
		t.Fatal("Cn error")
	}
	if lang != "" {
		t.Fatal("lang error")
	}

	name, lang, err = ParseLanguageTag("cn;lang-ja")
	if err != nil {
		t.Fatal(err)
	}
	if name != "cn" {
		t.Fatal("Cn error")
	}
	if lang != "lang-ja" {
		t.Fatal("lang error")
	}

	name, lang, err = ParseLanguageTag("cn;xxxx-ja")
	if err == nil {
		t.Fatal("need error")
	}

}
