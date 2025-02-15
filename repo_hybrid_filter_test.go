//go:build test

package ldap_pg

import (
	"reflect"
	"strings"
	"testing"

	"github.com/openstandia/goldap/message"
)

type HybridFilterTestData struct {
	label     string
	schemaMap SchemaMap
	filter    message.Filter
	err       string
	out       *HybridDBFilterTranslatorResult
}

func TestHybridFilter(t *testing.T) {
	server := NewServer(&ServerConfig{
		Suffix:          "dc=example,dc=com",
		QueryTranslator: "default",
	})
	server.LoadSchema()

	translator := HybridDBFilterTranslator{}

	for i, test := range createHybridFilterTestData() {
		var sb strings.Builder
		q := &HybridDBFilterTranslatorResult{
			where:  &sb,
			params: map[string]interface{}{},
		}

		err := translator.translate(server.schemaMap, test.filter, q, false)
		if err == nil {
			if test.out == nil {
				t.Errorf("#%d: %s\nEXPECTED ERROR MESSAGE:\n%s\nGOT A STRUCT INSTEAD:\n%#+v", i, test.label, test.err, q)
			} else if q.where.String() != test.out.where.String() || !reflect.DeepEqual(q.params, test.out.params) {
				t.Errorf(`#%d: %s
GOT:
	where: %s
	params: %v
EXPECTED:
	where: %s
	params: %v`, i, test.label, q.where.String(), q.params, test.out.where.String(), test.out.params)
			}
		} else if !strings.Contains(err.Error(), test.err) {
			t.Errorf("#%d: %s\nGOT:\n%s\nEXPECTED:\n%s", i, test.label, err.Error(), test.err)
		}
	}
}

func createHybridFilterTestData() (ret []HybridFilterTestData) {
	sb := func(s string) *strings.Builder {
		var b strings.Builder
		b.WriteString(s)
		return &b
	}
	return []HybridFilterTestData{
		{
			label: "cn=foo",
			schemaMap: SchemaMap{
				AttributeTypes: map[string]*AttributeType{
					"cn": {
						Name:        "cn",
						Equality:    "",
						SingleValue: true,
					},
				},
			},
			filter: message.NewFilterEqualityMatch("cn", "foo"),
			out: &HybridDBFilterTranslatorResult{
				where: sb("e.attrs_norm @@ :0"),
				params: map[string]interface{}{
					"0": `$."cn" == "foo"`,
				},
			},
		},

		{
			label: "(&(cn=foo)(uid=bar))",
			schemaMap: SchemaMap{
				AttributeTypes: map[string]*AttributeType{
					"cn": {
						Name:        "cn",
						Equality:    "",
						SingleValue: true,
					},
					"uid": {
						Name:        "uid",
						Equality:    "",
						SingleValue: true,
					},
				},
			},
			filter: message.FilterAnd{
				message.NewFilterEqualityMatch("cn", "foo"),
				message.NewFilterEqualityMatch("uid", "bar"),
			},
			out: &HybridDBFilterTranslatorResult{
				where: sb("(e.attrs_norm @@ :0 AND e.attrs_norm @@ :1)"),
				params: map[string]interface{}{
					"0": `$."cn" == "foo"`,
					"1": `$."uid" == "bar"`,
				},
			},
		},

		{
			label: "(|(cn=foo)(cn=bar))",
			schemaMap: SchemaMap{
				AttributeTypes: map[string]*AttributeType{
					"cn": {
						Name:        "cn",
						Equality:    "",
						SingleValue: true,
					},
				},
			},
			filter: message.FilterOr{
				message.NewFilterEqualityMatch("cn", "foo"),
				message.NewFilterEqualityMatch("cn", "bar"),
			},
			out: &HybridDBFilterTranslatorResult{
				where: sb("(e.attrs_norm @@ :0 OR e.attrs_norm @@ :1)"),
				params: map[string]interface{}{
					"0": `$."cn" == "foo"`,
					"1": `$."cn" == "bar"`,
				},
			},
		},

		{
			label: "(|(cn=foo)(cn=bar)(cn=hoge))",
			schemaMap: SchemaMap{
				AttributeTypes: map[string]*AttributeType{
					"cn": {
						Name:        "cn",
						Equality:    "",
						SingleValue: true,
					},
				},
			},
			filter: message.FilterOr{
				message.NewFilterEqualityMatch("cn", "foo"),
				message.NewFilterEqualityMatch("cn", "bar"),
				message.NewFilterEqualityMatch("cn", "hoge"),
			},
			out: &HybridDBFilterTranslatorResult{
				where: sb("(e.attrs_norm @@ :0 OR e.attrs_norm @@ :1 OR e.attrs_norm @@ :2)"),
				params: map[string]interface{}{
					"0": `$."cn" == "foo"`,
					"1": `$."cn" == "bar"`,
					"2": `$."cn" == "hoge"`,
				},
			},
		},

		{
			label: "(|(cn=foo)(&(uid=bar)(sn=hoge)))",
			schemaMap: SchemaMap{
				AttributeTypes: map[string]*AttributeType{
					"cn": {
						Name:        "cn",
						Equality:    "",
						SingleValue: true,
					},
					"uid": {
						Name:        "uid",
						Equality:    "",
						SingleValue: true,
					},
					"sn": {
						Name:        "sn",
						Equality:    "",
						SingleValue: true,
					},
				},
			},
			filter: message.FilterOr{
				message.NewFilterEqualityMatch("cn", "foo"),
				message.FilterAnd{
					message.NewFilterEqualityMatch("uid", "bar"),
					message.NewFilterEqualityMatch("sn", "hoge"),
				},
			},
			out: &HybridDBFilterTranslatorResult{
				where: sb("(e.attrs_norm @@ :0 OR (e.attrs_norm @@ :1 AND e.attrs_norm @@ :2))"),
				params: map[string]interface{}{
					"0": `$."cn" == "foo"`,
					"1": `$."uid" == "bar"`,
					"2": `$."sn" == "hoge"`,
				},
			},
		},

		{
			label: "(!(cn=foo))",
			schemaMap: SchemaMap{
				AttributeTypes: map[string]*AttributeType{
					"cn": {
						Name:        "cn",
						Equality:    "",
						SingleValue: true,
					},
				},
			},
			filter: message.FilterNot{
				Filter: message.NewFilterEqualityMatch("cn", "foo"),
			},
			out: &HybridDBFilterTranslatorResult{
				where: sb("e.attrs_norm @@ :0"),
				params: map[string]interface{}{
					"0": `!($."cn" == "foo")`,
				},
			},
		},

		{
			label: "(!(&(cn=foo)(uid=bar)))",
			schemaMap: SchemaMap{
				AttributeTypes: map[string]*AttributeType{
					"cn": {
						Name:        "cn",
						Equality:    "",
						SingleValue: true,
					},
					"uid": {
						Name:        "uid",
						Equality:    "",
						SingleValue: true,
					},
				},
			},
			filter: message.FilterNot{
				Filter: message.FilterAnd{
					message.NewFilterEqualityMatch("cn", "foo"),
					message.NewFilterEqualityMatch("uid", "bar"),
				},
			},
			out: &HybridDBFilterTranslatorResult{
				where: sb("(e.attrs_norm @@ :0 OR e.attrs_norm @@ :1)"),
				params: map[string]interface{}{
					"0": `!($."cn" == "foo")`,
					"1": `!($."uid" == "bar")`,
				},
			},
		},

		{
			label: "(!(|(cn=foo)(uid=bar)))",
			schemaMap: SchemaMap{
				AttributeTypes: map[string]*AttributeType{
					"cn": {
						Name:        "cn",
						Equality:    "",
						SingleValue: true,
					},
					"uid": {
						Name:        "uid",
						Equality:    "",
						SingleValue: true,
					},
				},
			},
			filter: message.FilterNot{
				Filter: message.FilterOr{
					message.NewFilterEqualityMatch("cn", "foo"),
					message.NewFilterEqualityMatch("uid", "bar"),
				},
			},
			out: &HybridDBFilterTranslatorResult{
				where: sb("(e.attrs_norm @@ :0 AND e.attrs_norm @@ :1)"),
				params: map[string]interface{}{
					"0": `!($."cn" == "foo")`,
					"1": `!($."uid" == "bar")`,
				},
			},
		},

		{
			label: "(!(|(&(cn=foo)(uid=bar))(sn=hoge)))",
			schemaMap: SchemaMap{
				AttributeTypes: map[string]*AttributeType{
					"cn": {
						Name:        "cn",
						Equality:    "",
						SingleValue: true,
					},
					"uid": {
						Name:        "uid",
						Equality:    "",
						SingleValue: true,
					},
					"sn": {
						Name:        "sn",
						Equality:    "",
						SingleValue: true,
					},
				},
			},
			filter: message.FilterNot{
				Filter: message.FilterOr{
					message.FilterAnd{
						message.NewFilterEqualityMatch("cn", "foo"),
						message.NewFilterEqualityMatch("uid", "bar"),
					},
					message.NewFilterEqualityMatch("sn", "hoge"),
				},
			},
			out: &HybridDBFilterTranslatorResult{
				where: sb("((e.attrs_norm @@ :0 OR e.attrs_norm @@ :1) AND e.attrs_norm @@ :2)"),
				params: map[string]interface{}{
					"0": `!($."cn" == "foo")`,
					"1": `!($."uid" == "bar")`,
					"2": `!($."sn" == "hoge")`,
				},
			},
		},

		{
			label: "(!(!(cn=foo)))",
			schemaMap: SchemaMap{
				AttributeTypes: map[string]*AttributeType{
					"cn": {
						Name:        "cn",
						Equality:    "",
						SingleValue: true,
					},
				},
			},
			filter: message.FilterNot{
				Filter: message.FilterNot{
					Filter: message.NewFilterEqualityMatch("cn", "foo"),
				},
			},
			out: &HybridDBFilterTranslatorResult{
				where: sb("e.attrs_norm @@ :0"),
				params: map[string]interface{}{
					"0": `$."cn" == "foo"`,
				},
			},
		},
	}
}
