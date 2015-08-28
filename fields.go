package linkedin

type Fields struct {
	Values map[string][]string
}

func (f *Fields) Add(field string, values ...string) {
	if f.Values == nil {
		f.Values = make(map[string][]string)
	}
	f.Values[field] = values
}

func (f *Fields) Encode() (fields string) {
	if len(f.Values) > 0 {
		str := ":("
		comma := ""
		for field, subfields := range f.Values {
			str += comma + field
			if len(subfields) > 0 {
				str += ":("
				subcomma := ""
				for _, subfield := range subfields {
					str += subcomma + subfield
					subcomma = ","
				}
				str += ")"
			}
			comma = ","
		}
		return str + ")"
	}
	return ""
}
