package linkedin

import (
	"testing"
)

func TestAddSingleNoSub(t *testing.T) {
	fields := Fields{}

	fields.Add("a")

	if len(fields.Values) == 1 {
		if value, ok := fields.Values["a"]; ok {
			if len(value) != 0 {
				t.Fatal("not expecting subfields")
			}
		} else {
			t.Fatal("field 'a' not found")
		}
	} else {
		t.Fatalf("incorrect number of values: %v", len(fields.Values))
	}
}

func TestAddSingleWithSub(t *testing.T) {
	fields := Fields{}

	fields.Add("a", "b")

	if len(fields.Values) == 1 {
		if value, ok := fields.Values["a"]; ok {
			if !(len(value) == 1 && value[0] == "b") {
				t.Fatalf("incorrect number of subfields: %v", len(value))
			}
		} else {
			t.Fatal("field 'a' not found")
		}
	} else {
		t.Fatalf("incorrect number of fields: %v", len(fields.Values))
	}
}

func TestAddSingleWithMultiSub(t *testing.T) {
	fields := Fields{}

	fields.Add("a", "b", "c")

	if len(fields.Values) == 1 {
		if value, ok := fields.Values["a"]; ok {
			if !(len(value) == 2 && value[0] == "b" && value[1] == "c") {
				t.Fatalf("incorrect number of subfields: %v", len(value))
			}
		} else {
			t.Fatal("field 'a' not found")
		}
	} else {
		t.Fatalf("incorrect number of fields: %v", len(fields.Values))
	}
}

func TestMultiAdd(t *testing.T) {
	fields := Fields{}

	fields.Add("a", "b")
	fields.Add("c")

	if len(fields.Values) != 2 {
		t.Fatalf("incorrect number of fields: %v", len(fields.Values))
	}
}

func TestEncodeSingleNoSub(t *testing.T) {
	fields := Fields{}

	fields.Add("a")

	if fields.Encode() != ":(a)" {
		t.Fatalf("expecting ':(a)', got '%v'", fields.Encode())
	}
}

func TestEncodeSingleWithSub(t *testing.T) {
	fields := Fields{}

	fields.Add("a", "b")

	if fields.Encode() != ":(a:(b))" {
		t.Fatalf("expecting ':(a:(b))', got '%v'", fields.Encode())
	}
}

func TestEncodeSingleMultiSub(t *testing.T) {
	fields := Fields{}

	fields.Add("a", "b", "c")

	if fields.Encode() != ":(a:(b,c))" {
		t.Fatalf("expecting ':(a:(b,c))', got '%v'", fields.Encode())
	}
}

func TestEncodeMulti(t *testing.T) {
	fields := Fields{}

	fields.Add("a")
	fields.Add("b", "c")

	if fields.Encode() != ":(a,b:(c))" {
		t.Fatalf("expecting ':(a,b:(c))', got '%v'", fields.Encode())
	}
}
