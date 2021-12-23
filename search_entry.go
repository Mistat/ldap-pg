package ldap_pg

type SearchEntry struct {
	schemaMap  *SchemaMap
	dnOrig     string
	attributes map[string][]string
}

func NewSearchEntry(schemaMap *SchemaMap, dnOrig string, valuesOrig map[string][]string) *SearchEntry {
	readEntry := &SearchEntry{
		schemaMap:  schemaMap,
		dnOrig:     dnOrig,
		attributes: valuesOrig,
	}
	return readEntry
}

func (j *SearchEntry) DNOrig() string {
	return j.dnOrig
}

func (j *SearchEntry) GetAttrsOrig() map[string][]string {
	return j.attributes
}

func (j *SearchEntry) GetAttrOrig(attrName string) (string, []string, bool) {
	name, lang, err := ParseLanguageTag(attrName)
	if err != nil {
		return "", nil, false
	}
	s, ok := j.schemaMap.AttributeType(name)
	if !ok {
		return "", nil, false
	}

	name = s.Name
	if lang != "" {
		name = name + ";" + lang
	}

	v, ok := j.attributes[name]
	if !ok {
		return "", nil, false
	}
	return name, v, true
}

func (j *SearchEntry) GetAttrsOrigWithoutOperationalAttrs() map[string][]string {
	m := map[string][]string{}
	for k, v := range j.attributes {
		if s, ok := j.schemaMap.AttributeType(k); ok {
			if !s.IsOperationalAttribute() {
				m[k] = v
			}
		}
	}
	return m
}

func (j *SearchEntry) GetOperationalAttrsOrig() map[string][]string {
	m := map[string][]string{}
	for k, v := range j.attributes {
		if s, ok := j.schemaMap.AttributeType(k); ok {
			if s.IsOperationalAttribute() {
				m[k] = v
			}
		}
	}
	return m
}
