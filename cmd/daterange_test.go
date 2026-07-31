package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// The bounds are anchored to time.Local, and the fixtures below are written in
// the +05:30 offset the site's JSON uses. Pin the zone so the results do not
// depend on the host the tests run on.
func TestMain(m *testing.M) {

	time.Local = time.FixedZone("IST", 5*60*60+30*60)
	os.Exit(m.Run())
}

// mustParse builds an instant in a fixed +05:30 zone, matching the offset Hugo
// writes into its JSON on the machine this tool is run from.
func mustParse(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("bad test timestamp %s: %v", value, err)
	}

	return parsed
}

func rangeFor(t *testing.T, from, to string) (time.Time, time.Time) {
	t.Helper()

	cmd := &cobra.Command{}
	addDateRangeFlags(cmd)
	if err := cmd.Flags().Parse([]string{"--from-date", from, "--to-date", to}); err != nil {
		t.Fatalf("parsing flags: %v", err)
	}

	start, end, err := dateRange(cmd)
	if err != nil {
		t.Fatalf("dateRange(%s, %s): %v", from, to, err)
	}

	return start, end
}

// A single-day window must cover that whole day. Before the fix both bounds
// landed on midnight UTC, so this window matched nothing at all.
func TestDateRangeCoversWholeDay(t *testing.T) {

	start, end := rangeFor(t, "2025-07-07", "2025-07-07")

	cases := []struct {
		instant string
		want    bool
	}{
		{"2025-07-06T23:59:59+05:30", false},
		{"2025-07-07T00:00:00+05:30", true},  // first instant of the day
		{"2025-07-07T00:17:57+05:30", true},  // before 05:30, previously lost to the UTC offset
		{"2025-07-07T09:15:41+05:30", true},  // a real Lastmod from the site
		{"2025-07-07T23:59:59+05:30", true},  // last instant, previously excluded
		{"2025-07-08T00:00:00+05:30", false}, // start of the next day
	}

	for _, c := range cases {
		if got := inRange(mustParse(t, c.instant), start, end); got != c.want {
			t.Errorf("inRange(%s) = %v, want %v", c.instant, got, c.want)
		}
	}
}

// The end bound is exclusive internally but must be reported to the user as the
// inclusive day they asked for.
func TestDateRangeBounds(t *testing.T) {

	start, end := rangeFor(t, "2025-01-01", "2025-12-31")

	if got := start.Format(dateLayout); got != "2025-01-01" {
		t.Errorf("start = %s, want 2025-01-01", got)
	}
	if got := end.Format(dateLayout); got != "2026-01-01" {
		t.Errorf("end = %s, want 2026-01-01 (exclusive)", got)
	}
	if got := describeRange(start, end); got != "2025-01-01 to 2025-12-31 (inclusive)" {
		t.Errorf("describeRange = %q", got)
	}
	if start.Location() != time.Local || end.Location() != time.Local {
		t.Errorf("bounds must be anchored to local time, got %s and %s", start.Location(), end.Location())
	}
}

func TestDateRangeRejectsInvertedWindow(t *testing.T) {

	cmd := &cobra.Command{}
	addDateRangeFlags(cmd)
	if err := cmd.Flags().Parse([]string{"--from-date", "2026-01-01", "--to-date", "2025-01-01"}); err != nil {
		t.Fatalf("parsing flags: %v", err)
	}

	if _, _, err := dateRange(cmd); err == nil {
		t.Fatal("expected an error when from-date is after to-date")
	}
}

// The defaults are wall-clock times with a time of day, unlike the flags, which
// parse to midnight UTC. Both must normalise the same way.
func TestDateRangeDefaultsCoverToday(t *testing.T) {

	cmd := &cobra.Command{}
	addDateRangeFlags(cmd)

	start, end, err := dateRange(cmd)
	if err != nil {
		t.Fatalf("dateRange with defaults: %v", err)
	}

	now := time.Now()
	if !inRange(now, start, end) {
		t.Errorf("default window %s does not include now (%s)", describeRange(start, end), now)
	}

	yearAgo := startOfDay(now.AddDate(-1, 0, 0))
	if !start.Equal(yearAgo) {
		t.Errorf("default start = %s, want %s", start, yearAgo)
	}
}

func TestStartOfDayDiscardsOriginalZone(t *testing.T) {

	// What pflag hands back for "--from-date 2025-07-07": midnight UTC.
	utcMidnight := time.Date(2025, 7, 7, 0, 0, 0, 0, time.UTC)

	got := startOfDay(utcMidnight)
	want := time.Date(2025, 7, 7, 0, 0, 0, 0, time.Local)

	if !got.Equal(want) {
		t.Errorf("startOfDay(%s) = %s, want %s", utcMidnight, got, want)
	}
}
