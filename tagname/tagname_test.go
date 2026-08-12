package tagname

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Emits the parallel strings the Hugo partial needs, derived from the same
// tables, so the two implementations cannot drift.
func TestGeneratePartialTables(t *testing.T) {
	var from, to []string
	for _, g := range foldGroups {
		for _, r := range g.runes {
			from = append(from, string(r))
			to = append(to, g.ascii)
		}
	}
	fmt.Printf("SINGLE from (%d runes):\n%s\n", len(from), strings.Join(from, ""))
	fmt.Printf("SINGLE to   (%d runes):\n%s\n", len(to), strings.Join(to, ""))
	if len(from) != len(to) {
		t.Fatal("length mismatch")
	}

	keys := make([]string, 0, len(foldMulti))
	for r := range foldMulti {
		keys = append(keys, string(r))
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%q %q", k, foldMulti[[]rune(k)[0]]))
	}
	fmt.Printf("MULTI dict:\n(dict %s)\n", strings.Join(parts, " "))
}

// partialSim reproduces, operation for operation, what
// layouts/partials/taxonomy/title-to-tag.html does: findRE guard, \p{Mn} strip,
// multi-rune dict replaces, then 180 single-rune replaces off two parallel
// strings split into runes, then the collapse regex and trim.
const (
	simFrom = "ÀÁÂÃÄÅĀĂĄàáâãäåāăąÇĆĈĊČçćĉċčÐĎĐðďđÈÉÊËĒĔĖĘĚèéêëēĕėęěĜĞĠĢĝğġģĤĦĥħÌÍÎÏĨĪĬĮİìíîïĩīĭįıĴĵĶķĸĹĻĽĿŁĺļľŀłÑŃŅŇŊñńņňŉŋÒÓÔÕÖØŌŎŐòóôõöøōŏőŔŖŘŕŗřŚŜŞŠśŝşšŢŤŦţťŧÙÚÛÜŨŪŬŮŰŲùúûüũūŭůűųŴŵÝŸŶýÿŷŹŻŽźżž"
	simTo   = "AAAAAAAAAaaaaaaaaaCCCCCcccccDDDdddEEEEEEEEEeeeeeeeeeGGGGggggHHhhIIIIIIIIIiiiiiiiiiJjKkkLLLLLlllllNNNNNnnnnnnOOOOOOOOOoooooooooRRRrrrSSSSssssTTTtttUUUUUUUUUUuuuuuuuuuuWwYYYyyyZZZzzz"
)

var simMulti = map[string]string{
	"Æ": "AE", "Þ": "TH", "ß": "ss", "æ": "ae", "þ": "th",
	"Ĳ": "IJ", "ĳ": "ij", "Œ": "OE", "œ": "oe",
}

var (
	simNonASCII = regexp.MustCompile(`[^ -~]`)
	simMn       = regexp.MustCompile(`\p{Mn}`)
	simCollapse = regexp.MustCompile(`[^A-Za-z0-9]+`)
)

func partialSim(in string) string {
	s := in
	if simNonASCII.FindString(s) != "" {
		s = simMn.ReplaceAllString(s, "")
		for k, v := range simMulti {
			s = strings.ReplaceAll(s, k, v)
		}
		from := strings.Split(simFrom, "")
		to := strings.Split(simTo, "")
		if len(from) != len(to) {
			panic("parallel table length mismatch")
		}
		for i, c := range from {
			s = strings.ReplaceAll(s, c, to[i])
		}
	}
	s = simCollapse.ReplaceAllString(s, " ")
	return strings.Trim(s, " ")
}

func TestPartialMatchesPackage(t *testing.T) {
	fixed := []string{
		"Q&A: What's Next?", "Naïve Café", "Node.js  v1.2.3", "AT&T",
		"Rock & Roll", "  padded  ", "The Quick—Brown Fox", "Don't Panic!!!",
		"Ærøskøbing", "Straße", "Þórr", "Œuvre", "ĲsselmeerIJ", "İstanbul",
		"Łódź", "Šibenik", "Dvořák", "Ångström", "façade", "50% off!",
		"café" /* NFC */, "cafe\u0301" /* NFD */, "---", "日本語", "C#", "iPhone 15 Pro",
	}
	for _, in := range fixed {
		if a, b := FromTitle(in), partialSim(in); a != b {
			t.Errorf("MISMATCH %q: pkg=%q partial=%q", in, a, b)
		}
	}

	pool := []rune("abXY09 -_.&,:;!?'\"()%#/éïøÆßŁ\u0301\u0308日 ")
	r := rand.New(rand.NewSource(11))
	for i := 0; i < 300000; i++ {
		n := 1 + r.Intn(16)
		b := make([]rune, n)
		for j := range b {
			b[j] = pool[r.Intn(len(pool))]
		}
		in := string(b)
		if x, y := FromTitle(in), partialSim(in); x != y {
			t.Fatalf("FUZZ MISMATCH %q: pkg=%q partial=%q", in, x, y)
		}
	}
}

// Every tag produced must be canonical, and re-running must be a no-op.
func TestOutputIsCanonicalAndIdempotent(t *testing.T) {
	valid := regexp.MustCompile(`^[A-Za-z0-9]+( [A-Za-z0-9]+)*$`)
	pool := []rune("abXY09 -_.&,:;!?'\"()%#/éïøÆßŁ\u0301日 ")
	r := rand.New(rand.NewSource(23))
	for i := 0; i < 300000; i++ {
		n := 1 + r.Intn(16)
		b := make([]rune, n)
		for j := range b {
			b[j] = pool[r.Intn(len(pool))]
		}
		tag := FromTitle(string(b))
		if tag == "" {
			continue
		}
		if !valid.MatchString(tag) {
			t.Fatalf("not canonical: %q -> %q", string(b), tag)
		}
		if again := FromTitle(tag); again != tag {
			t.Fatalf("not idempotent: %q -> %q -> %q", string(b), tag, again)
		}
		if !IsCanonical(tag) {
			t.Fatalf("IsCanonical false for %q", tag)
		}
	}
}
