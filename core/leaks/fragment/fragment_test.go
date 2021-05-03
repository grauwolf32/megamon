package fragment

import (
	"fmt"
	"testing"
	"unicode/utf8"
)

func TestFragmentUnion1(t *testing.T) {
	f1 := Fragment{0, 10}
	f2 := Fragment{3, 15}

	f := UnionFragments(&f1, &f2)
	if f.Length != 18 {
		t.Errorf("Fragment length must be 18 got %d", f.Length)
	}

	if f.Offset != 0 {
		t.Errorf("Fragment offset must be 0 got %d", f.Offset)
	}
}

func TestFragmentUnion2(t *testing.T) {
	f1 := Fragment{0, 10}
	f2 := Fragment{0, 10}

	f := UnionFragments(&f1, &f2)
	if f.Length != 10 {
		t.Errorf("Fragment length must be 10 got %d", f.Length)
	}

	if f.Offset != 0 {
		t.Errorf("Fragment offset must be 0 got %d", f.Offset)
	}
}

func TestFragmentUnion3(t *testing.T) {
	f1 := Fragment{0, 5}
	f2 := Fragment{5, 5}

	f := UnionFragments(&f1, &f2)
	if f.Length != 10 {
		t.Errorf("Fragment length must be 11 got %d", f.Length)
	}

	if f.Offset != 0 {
		t.Errorf("Fragment offset must be 0 got %d", f.Offset)
	}
}

func TestAlignToRunes1(t *testing.T) {
	test := "Hello, мир, test"
	f := Fragment{7, 6}
	f.AlignToRunes(test)

	if res, _ := f.Apply(test); res != "мир" {
		t.Errorf("Expected: мир got: %s.", res)
	}
}

func TestConvertToRunes(t *testing.T) {
	test := "мир"
	if utf8.RuneCountInString(test) != 3 {
		t.Errorf("Expected lenght: 3 got %d", utf8.RuneCountInString(test))
	}
	if len(test) != 6 {
		t.Errorf("Expected lenght: 6 got %d", len(test))
	}

	f := Fragment{0, 6}
	err := f.ConvertToRunes(test)

	if err != nil {
		t.Errorf("%s", err.Error())
		return
	}

	if f.Offset != 0 {
		t.Errorf("Offset must be 0 got %d", f.Offset)
	}

	if f.Length != 3 {
		t.Errorf("Length must be 3 got %d", f.Length)
	}
	return
}

func TestGetKeywordFragments1(t *testing.T) {
	test := "test test xxx test"
	frags := GetKeywordFragments(test, "test")

	if len(frags) != 3 {
		t.Errorf("Expected 3 fragment, got: %d", len(frags))
	}

	kwLen := len("test")
	if frags[0].Offset != 0 {
		t.Errorf("Expected 1st offset 0 , got: %d", frags[0].Offset)
	}

	if frags[1].Offset != 5 {
		t.Errorf("Expected 2nd offset 5, got: %d", frags[1].Offset)
	}
	if frags[2].Offset != 14 {
		t.Errorf("Expected 3rd offset 14, got: %d", frags[2].Offset)
	}

	if frags[0].Length != kwLen || frags[1].Length != kwLen || frags[2].Length != kwLen {
		t.Errorf("Wrong fragment length: expected %d, got %d %d %d", kwLen, frags[0].Length, frags[1].Length, frags[2].Length)
	}
	return
}

func TestGetKeywordFragments2(t *testing.T) {
	test := "abababab"
	frags := GetKeywordFragments(test, "abab")

	if len(frags) != 3 {
		t.Errorf("Expected 3 fragment, got: %d", len(frags))
	}

	kwLen := len("test")
	if frags[0].Offset != 0 {
		t.Errorf("Expected 1st offset 0 , got: %d", frags[0].Offset)
	}

	if frags[1].Offset != 2 {
		t.Errorf("Expected 2nd offset 2, got: %d", frags[1].Offset)
	}
	if frags[2].Offset != 4 {
		t.Errorf("Expected 3rd offset 4, got: %d", frags[2].Offset)
	}

	if frags[0].Length != kwLen || frags[1].Length != kwLen || frags[2].Length != kwLen {
		t.Errorf("Wrong fragment length: expected %d, got %d %d %d", kwLen, frags[0].Length, frags[1].Length, frags[2].Length)
	}
	return
}

func TestJoin1(t *testing.T) {
	a := Fragment{0, 5}
	b := Fragment{5, 7}
	c := Fragment{12, 8}

	f := []Fragment{a, b, c}
	r := Join(&f, 15)

	if len(r) != 2 {
		t.Errorf("Wrong length of the join: expected 2, got: %d", len(r))
		return
	}

	if r[0].Offset != 0 || r[0].Length != 12 {
		t.Errorf("Wrong offset&length of the 1st element: expected 0 & 12, got: %d %d", r[0].Offset, r[0].Length)
	}

	if r[1].Offset != 12 || r[1].Length != 8 {
		t.Errorf("Wrong offset&length of the 1st element: expected 12 & 8, got: %d %d", r[1].Offset, r[1].Length)
	}

	return
}

func TestJoin2(t *testing.T) {
	a := Fragment{0, 5}
	b := Fragment{5, 7}
	c := Fragment{12, 8}

	f := []Fragment{a, b, c}
	r := Join(&f, 20)

	if len(r) != 1 {
		t.Errorf("Wrong length of the join: expected 1, got: %d", len(r))
		return
	}

	if r[0].Offset != 0 || r[0].Length != 20 {
		t.Errorf("Wrong offset&length of the 1st element: expected 0 & 12, got: %d %d", r[0].Offset, r[0].Length)
	}
	return
}

func TestMerge1(t *testing.T) {
	f1 := []Fragment{
		{0, 10},
		{5, 10},
	}

	f2 := []Fragment{
		{2, 10},
		{7, 10},
	}

	f := Merge(&f1, &f2)
	expected := "[{0 10} {2 10} {5 10} {7 10}]"
	if fmt.Sprintf("%v", f) != expected {
		t.Errorf("Wrong merge: expected: %s got %v", expected, f)
	}
	return
}

func TestMerge2(t *testing.T) {
	f1 := []Fragment{
		{0, 10},
		{5, 10},
		{12, 10},
	}

	f2 := []Fragment{
		{2, 10},
		{7, 10},
	}

	f := Merge(&f1, &f2)
	expected := "[{0 10} {2 10} {5 10} {7 10} {12 10}]"
	if fmt.Sprintf("%v", f) != expected {
		t.Errorf("Wrong merge: expected: %s got %v", expected, f)
	}
	return
}

func TestMergeFragments1(t *testing.T) {
	f1 := []Fragment{
		{0, 10},
		{5, 10},
	}

	f2 := []Fragment{
		{2, 10},
		{7, 10},
	}

	f3 := []Fragment{
		{12, 10},
		{15, 8},
	}

	r := MergeFragments(&[][]Fragment{f1, f2, f3}, 15)
	expected := "[{0 15} {7 15} {15 8}]"

	if fmt.Sprintf("%v", r) != expected {
		t.Errorf("Wrong merge: expected: %s got %v", expected, r)
	}
	return
}

func TestGetKeywordContext(t *testing.T) {
	text := "jsadjf; sdjfsdfjk adjsfk sdafjkds fjadsfkj afjdask test jfdskalfjds dsjfkljadsf ajkdflads"
	kwFrags := GetKeywordFragments(text, "test")
	contexts := make([]Fragment, 0, len(kwFrags))

	for _, frag := range kwFrags {
		context := GetKeywordContext(text, 10, frag)
		contexts = append(contexts, context)
	}

	if len(contexts) != 1 {
		t.Errorf("Expected length: 1 got %d", len(contexts))
	}

	r, _ := contexts[0].Apply(text)
	expected := "sk test jfd"

	if r != expected {
		t.Errorf("Wrong merge: expected: %s got %v", expected, r)
	}
	return
}
