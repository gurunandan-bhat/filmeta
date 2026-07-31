package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

const dateLayout = "2006-01-02"

// addDateRangeFlags registers the --from-date/--to-date pair shared by the
// reporting commands. Both bounds are inclusive calendar days; the default
// window is the last year ending today.
func addDateRangeFlags(cmd *cobra.Command) {

	now := time.Now()

	cmd.Flags().TimeP("from-date", "f", now.AddDate(-1, 0, 0), []string{dateLayout}, "Start date (inclusive), as YYYY-MM-DD")
	cmd.Flags().TimeP("to-date", "t", now, []string{dateLayout}, "End date (inclusive), as YYYY-MM-DD")
}

// dateRange reads --from-date/--to-date and returns the half-open interval
// [start, end) covering both days in full.
//
// Two normalisations happen here. pflag parses a bare YYYY-MM-DD with
// time.Parse, which yields midnight UTC, whereas the timestamps Hugo writes
// into its JSON are local; both bounds are therefore re-anchored to local
// midnight by calendar date. The end bound is then advanced to the start of the
// following day, so that everything published on --to-date is included rather
// than only the instant of midnight.
func dateRange(cmd *cobra.Command) (start, end time.Time, err error) {

	from, err := cmd.Flags().GetTime("from-date")
	if err != nil {
		return start, end, fmt.Errorf("error reading from date: %w", err)
	}
	to, err := cmd.Flags().GetTime("to-date")
	if err != nil {
		return start, end, fmt.Errorf("error reading to date: %w", err)
	}

	start = startOfDay(from)
	end = startOfDay(to).AddDate(0, 0, 1)

	if !start.Before(end) {
		return start, end, fmt.Errorf("from-date %s is after to-date %s", start.Format(dateLayout), startOfDay(to).Format(dateLayout))
	}

	return start, end, nil
}

// startOfDay re-anchors t to local midnight on its own calendar date,
// discarding both its time of day and its original location.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

// inRange reports whether t falls in the half-open interval [start, end).
func inRange(t, start, end time.Time) bool {
	return !t.Before(start) && t.Before(end)
}

// describeRange renders the interval back as the inclusive pair of days the
// user asked for.
func describeRange(start, end time.Time) string {
	return fmt.Sprintf("%s to %s (inclusive)", start.Format(dateLayout), end.AddDate(0, 0, -1).Format(dateLayout))
}
