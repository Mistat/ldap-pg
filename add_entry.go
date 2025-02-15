package ldap_pg

type AddEntry struct {
	schemaMap  *SchemaMap
	dn         *DN
	attributes map[string]*SchemaValue
}

func NewAddEntry(schemaMap *SchemaMap, dn *DN) *AddEntry {
	entry := &AddEntry{
		schemaMap:  schemaMap,
		attributes: map[string]*SchemaValue{},
	}
	entry.SetDN(dn)

	return entry
}

func (j *AddEntry) HasAttr(attrName string) bool {
	s, ok := j.schemaMap.AttributeType(attrName)
	if !ok {
		return false
	}

	_, ok = j.attributes[s.Name]
	return ok
}

func (j *AddEntry) SetDN(dn *DN) {
	j.dn = dn

	rdn := dn.RDN()
	for k, v := range rdn {
		// rdn is validated already
		j.attributes[k], _ = NewSchemaValue(j.schemaMap, k, []string{v.Orig})
	}
}

func (j *AddEntry) IsRoot() bool {
	return j.dn.IsRoot()
}

func (j *AddEntry) DN() *DN {
	return j.dn
}

func (j *AddEntry) ParentDN() *DN {
	return j.dn.ParentDN()
}

func (j *AddEntry) IsDC() bool {
	return j.dn.IsDC()
}

func (j *AddEntry) Validate() error {
	// objectClass is required
	if !j.HasAttr("objectClass") {
		return NewObjectClassViolation()
	}

	// Validate objectClass
	sv := j.attributes["objectClass"]
	if err := j.schemaMap.ValidateObjectClass(sv.Orig(), j.attributes); err != nil {
		return err
	}

	return nil
}

// Append to current value(s).
func (j *AddEntry) Add(attrName string, attrValue []string) error {
	if len(attrValue) == 0 {
		return nil
	}
	// Duplicate error is detected here
	sv, err := NewSchemaValue(j.schemaMap, attrName, attrValue)
	if err != nil {
		return err
	}
	if sv.IsNoUserModificationWithMigrationDisabled() {
		return NewNoUserModificationAllowedConstraintViolation(sv.Name())
	}
	return j.addsv(sv)
}

func (j *AddEntry) addsv(value *SchemaValue) error {
	name := value.Name()
	add := func(n string, v *SchemaValue) {
		current, ok := j.attributes[n]
		if !ok {
			j.attributes[n] = v
			//return nil
		} else {
			// When adding the attribute with same value as both DN and attribute,
			// we need to ignore the duplicate error.
			current.Add(v)
		}
	}
	add(name, value)
	langTag := value.LanguageTag()
	if langTag != "" {
		name = name + ";" + langTag
		add(name, value)
	}
	return nil
}

func (j *AddEntry) Attrs() (map[string][]interface{}, map[string][]string) {
	norm := make(map[string][]interface{}, len(j.attributes))
	orig := make(map[string][]string, len(j.attributes))
	for k, v := range j.attributes {
		norm[k] = v.Norm()
		orig[k] = v.Orig()
	}
	return norm, orig
}
