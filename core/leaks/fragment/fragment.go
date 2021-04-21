package fragment

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

//Fragment :
// Offset: byte offset from the beginning of the text
// Length: the length of the fragment
type Fragment struct {
	Offset int
	Length int
}

//Apply : get text fragment
func (frag *Fragment) Apply(text string) (textFrag string, err error) {
	fragEnd := frag.Offset + frag.Length

	if len(text) < fragEnd {
		err = fmt.Errorf("Fragment scope is beyond the supplied text")
		return
	}

	textFrag = text[frag.Offset:fragEnd]
	return
}

//AlignToRunes : correct offset and length to include full rune in bytes
func (frag *Fragment) AlignToRunes(text string) (err error) {
	textFrag, err := frag.Apply(text)
	if err != nil {
		return
	}

	check, _ := utf8.DecodeRuneInString(textFrag)
	if check == utf8.RuneError {
		if frag.Offset > 0 {
			frag.Offset--
			frag.Length++
		} else if frag.Length > 0 {
			frag.Offset++
			frag.Length--
		}
	}

	check, _ = utf8.DecodeLastRuneInString(textFrag)
	if check == utf8.RuneError {
		if frag.Offset+frag.Length < len(text)-1 {
			frag.Length++
		} else if frag.Length > 0 {
			frag.Length--
		}
	}

	return
}

//ConvertToRunes : change length of the fragment & its aligment to fit runes
func (frag *Fragment) ConvertToRunes(text string) (err error) {
	textFrag, err := frag.Apply(text)

	if err != nil {
		return
	}

	if check, _ := utf8.DecodeRuneInString(textFrag); check == utf8.RuneError {
		err = fmt.Errorf("First rune decode error")
		return
	}

	if check, _ := utf8.DecodeLastRuneInString(textFrag); check == utf8.RuneError {
		err = fmt.Errorf("Last rune decode error")
		return
	}

	frag.Length = utf8.RuneCountInString(textFrag)
	return
}

//UnionFragments : unite two fragments
func UnionFragments(f1, f2 *Fragment) (frag Fragment) {
	minOffset := f1.Offset
	if minOffset > f2.Offset {
		minOffset = f2.Offset
	}

	maxInd := f1.Length + f1.Offset
	if maxInd < f2.Length+f2.Offset {
		maxInd = f2.Length + f2.Offset
	}

	frag.Offset = minOffset
	frag.Length = maxInd - minOffset

	return
}

//GetKeywordFragments : searches keyword in the text and return all its indices
func GetKeywordFragments(text, keyword string) (fragments []Fragment) {
	n := len(keyword)
	m := len(text)
	fragments = make([]Fragment, 0, 16)

	offset := 0
	textSlice := text

	for ; offset+n < m; offset++ {
		ind := strings.Index(textSlice, keyword)
		if ind == -1 {
			break
		}

		offset += ind
		fragments = append(fragments, Fragment{offset, n})
		textSlice = textSlice[ind+1:]
	}
	return
}

func le(f1, f2 *Fragment) bool {
	return f1.Offset <= f2.Offset
}

func merge(f1, f2 []Fragment) []Fragment {
	f1Len := len(f1)
	f2Len := len(f2)

	length := f1Len + f2Len
	f := make([]Fragment, 0, length)

	var p1, p2 int
	for p1 < f1Len && p2 < f2Len {
		if le(&f1[p1], &f2[p2]) {
			f = append(f, f1[p1])
			p1++
		} else {
			f = append(f, f2[p2])
			p2++
		}
	}

	if p1 < f1Len {
		f = append(f, f1[p1:]...)
	} else if p2 < f2Len {
		f = append(f, f2[p2:]...)
	}
	return f
}

func join(frags []Fragment, maxFragLen int) (f []Fragment) {
	if len(frags) < 2 {
		return frags
	}

	f = make([]Fragment, 0, len(frags))
	currFrag := frags[0]

	for i := 1; i < len(frags); i++ {
		unionFrag := UnionFragments(&currFrag, &frags[i])
		if unionFrag.Length <= maxFragLen {
			currFrag = unionFrag
		} else {
			f = append(f, currFrag)
			currFrag = frags[i]
		}
	}
	f = append(f, currFrag)
	return
}

//MergeFragments : merge fragments optimal way
func MergeFragments(frags [][]Fragment, maxFragLen int) []Fragment {
	if len(frags) == 0 {
		return []Fragment{}
	}
	if len(frags) == 1 {
		return join(frags[0], maxFragLen)
	}

	merged := merge(frags[0], frags[1])
	for i := 2; i < len(frags); i++ {
		merged = merge(merged, frags[i])
	}

	return join(merged, maxFragLen)
}

//GetKeywordContext : get desired lenght word context if possible
func GetKeywordContext(text string, desiredLen int, frags []Fragment) []Fragment {
	if len(frags) == 0 {
		return []Fragment{}
	}

	var n, c, offset int
	n = frags[0].Length
	c = (n + 1) / 2
	offset = (desiredLen + 1) / 2

	for i, frag := range frags {
		pivot := frag.Offset + c
		leftBorder := pivot - offset
		rightBorder := pivot + offset

		if leftBorder < 0 {
			extra := -leftBorder
			leftBorder = 0
			if rightBorder+extra < len(text) {
				rightBorder = rightBorder + extra
			} else {
				rightBorder = len(text) - 1
			}
		}

		if rightBorder >= len(text) {
			extra := rightBorder - len(text) + 1
			if leftBorder-extra > 0 {
				leftBorder = leftBorder - extra
			} else {
				leftBorder = 0
			}
		}

		frags[i].Offset = leftBorder
		frags[i].Length = rightBorder - leftBorder + 1
		frags[i].AlignToRunes(text)
	}
	return frags
}
