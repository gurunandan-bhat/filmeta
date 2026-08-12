// Package tagname converts an article title into a canonical Hugo taxonomy tag.
//
// A canonical tag contains only ASCII alphanumerics separated by single spaces,
// with no leading or trailing space, and preserves the original letter case:
//
//	"Q&A: What's Next?"  ->  "Q A What s Next"
//	"Naïve Café"         ->  "Naive Cafe"
//	"Node.js  v1.2.3"    ->  "Node js v1 2 3"
//
// Such a tag is stable under Hugo's term-to-path transformation: the published
// URL segment is exactly strings.ToLower(strings.ReplaceAll(tag, " ", "-")),
// which is also what Hugo's `urlize` template function returns for it. Because
// the tag contains no character that Hugo's paths.Sanitize deletes, the two
// internal normalization passes cannot disagree, and because case and spacing
// survive, the tag still reads as a human title in .Page.Title.
//
// Pipeline:
//
//  1. Strip Unicode combining marks (category Mn). This handles NFD input,
//     which macOS filesystems produce, where "é" arrives as "e" + U+0301.
//  2. Fold precomposed Latin letters to ASCII via an explicit table.
//  3. Replace every run of non-[A-Za-z0-9] characters with a single space.
//  4. Trim leading and trailing spaces.
package tagname

import (
	"regexp"
	"strings"
	"unicode"
)

// foldGroups maps a set of Latin letters onto the single ASCII letter they fold
// to. Covers Latin-1 Supplement and Latin Extended-A.
var foldGroups = []struct {
	ascii string
	runes string
}{
	{"A", "ÀÁÂÃÄÅĀĂĄ"},
	{"a", "àáâãäåāăą"},
	{"C", "ÇĆĈĊČ"},
	{"c", "çćĉċč"},
	{"D", "ÐĎĐ"},
	{"d", "ðďđ"},
	{"E", "ÈÉÊËĒĔĖĘĚ"},
	{"e", "èéêëēĕėęě"},
	{"G", "ĜĞĠĢ"},
	{"g", "ĝğġģ"},
	{"H", "ĤĦ"},
	{"h", "ĥħ"},
	{"I", "ÌÍÎÏĨĪĬĮİ"},
	{"i", "ìíîïĩīĭįı"},
	{"J", "Ĵ"},
	{"j", "ĵ"},
	{"K", "Ķ"},
	{"k", "ķĸ"},
	{"L", "ĹĻĽĿŁ"},
	{"l", "ĺļľŀł"},
	{"N", "ÑŃŅŇŊ"},
	{"n", "ñńņňŉŋ"},
	{"O", "ÒÓÔÕÖØŌŎŐ"},
	{"o", "òóôõöøōŏő"},
	{"R", "ŔŖŘ"},
	{"r", "ŕŗř"},
	{"S", "ŚŜŞŠ"},
	{"s", "śŝşš"},
	{"T", "ŢŤŦ"},
	{"t", "ţťŧ"},
	{"U", "ÙÚÛÜŨŪŬŮŰŲ"},
	{"u", "ùúûüũūŭůűų"},
	{"W", "Ŵ"},
	{"w", "ŵ"},
	{"Y", "ÝŸŶ"},
	{"y", "ýÿŷ"},
	{"Z", "ŹŻŽ"},
	{"z", "źżž"},
}

// foldMulti holds the letters with no single-letter ASCII equivalent. These
// have no Unicode decomposition, so a normalize-and-strip-marks approach misses
// them entirely.
var foldMulti = map[rune]string{
	'Æ': "AE", 'æ': "ae",
	'Œ': "OE", 'œ': "oe",
	'ß': "ss",
	'Þ': "TH", 'þ': "th",
	'Ĳ': "IJ", 'ĳ': "ij",
}

var fold map[rune]string

func init() {
	fold = make(map[rune]string, 256)
	for _, g := range foldGroups {
		for _, r := range g.runes {
			fold[r] = g.ascii
		}
	}
	for r, s := range foldMulti {
		fold[r] = s
	}
}

var nonAlnum = regexp.MustCompile(`[^A-Za-z0-9]+`)

// FromTitle converts a title into a canonical tag.
//
// It returns the empty string when the title contains no character that folds
// to an ASCII alphanumeric — for example a title written entirely in a
// non-Latin script, which has no ASCII equivalent to fold to. Callers should
// treat that as "needs a hand-written tag" rather than silently accepting it.
func FromTitle(title string) string {
	var b strings.Builder
	b.Grow(len(title))

	for _, r := range title {
		if unicode.Is(unicode.Mn, r) {
			continue // combining mark from NFD input
		}
		if s, ok := fold[r]; ok {
			b.WriteString(s)
			continue
		}
		b.WriteRune(r)
	}

	return strings.TrimSpace(nonAlnum.ReplaceAllString(b.String(), " "))
}

// IsCanonical reports whether tag is already in canonical form, i.e. whether
// FromTitle would leave it unchanged. Useful for auditing a corpus before
// rewriting it, and for making a migration safely re-runnable.
func IsCanonical(tag string) bool {
	return tag != "" && tag == FromTitle(tag)
}
