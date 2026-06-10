package tags

import (
	"reflect"
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"  Code ":  "code",
		"#release": "release",
		"":         "",
		"WIP":      "wip",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q want %q", in, got, want)
		}
	}
}

func TestParse(t *testing.T) {
	got := Parse("Code, #qa  release,code")
	want := []string{"code", "qa", "release"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse = %v want %v", got, want)
	}
	if Parse("") != nil {
		t.Error("Parse(empty) want nil")
	}
}

func TestUnion(t *testing.T) {
	got := Union([]string{"a", "b"}, []string{"b", "c"})
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Union = %v want %v", got, want)
	}
}

func TestIntersect(t *testing.T) {
	got := Intersect([]string{"a", "b", "c"}, []string{"b", "c", "d"})
	want := []string{"b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Intersect = %v want %v", got, want)
	}
}
