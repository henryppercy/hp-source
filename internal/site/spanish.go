package site

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/site/templates"
)

// dsLevels are the Dreaming Spanish level thresholds in hours. The level a
// learner sits on is the highest threshold they have passed; the band figure
// shows progress toward the next.
var dsLevels = []int{0, 50, 150, 300, 600, 1000, 1500}

// The current goal: reach spanishGoalHours by the end of the given month and
// year. Change these three and the figure updates everywhere on the page.
const (
	spanishGoalHours = 800
	spanishGoalYear  = 2026
	spanishGoalMonth = time.December
)

const spanishStandfirst = "The stats and trends which map my journey to functional " +
	"Spanish fluency. Currently learning via Comprehensible Input."

const spanishIntro = "I am currently learning Spanish using Comprehensible Input, " +
	"a language learning style favouring comprehension over output. The core is formed " +
	"of listening to spoken language at a level slightly above my current proficiency, " +
	"before incorporating reading and speaking."

const spanishNote = "I'm currently aiming for roughly an hour of input per day. The majority of this comes from intermediate podcasts such as Español Al Vuelo, and YouTube channels like La Capital or more recently Ramilla de Aventura."

// spanishGoalDeadline is the last moment of the goal month.
func spanishGoalDeadline(loc *time.Location) time.Time {
	return time.Date(spanishGoalYear, spanishGoalMonth+1, 0, 23, 59, 59, 0, loc)
}

// spanishDay is a single day's total input, aggregated across sessions.
type spanishDay struct {
	date time.Time
	sec  int
}

// spanishView assembles the /spanish dashboard from the raw log and the Spanish
// writing feed items.
func spanishView(
	entries []repo.SpanishLogEntry,
	articles []templates.PostListItem,
	now time.Time,
) templates.SpanishView {
	days, secByDate := aggregateSpanish(entries)
	if len(days) == 0 {
		return templates.SpanishView{Standfirst: spanishStandfirst, Year: now.Year(), Articles: articles}
	}

	total := 0
	for _, d := range days {
		total += d.sec
	}
	start := days[0].date
	today := dateOnly(now)
	dayCount := spanishDayCount(days, now)

	v := templates.SpanishView{
		Total:      fmt.Sprintf("%d", total/3600),
		Standfirst: spanishStandfirst,
		Intro:      spanishIntro,
		StartDate:  start,
		DayCount:   dayCount,
		Year:       now.Year(),
		Goal:       spanishGoal(days, total, start, now),
		Calendar:   spanishCalendar(secByDate, start, today),
		Records:    spanishRecords(days),
		Averages:   spanishAverages(total, dayCount, len(days)),
		Note:       spanishNote,
		Articles:   articles,
	}
	v.Band = spanishBand(total)
	v.Stats = spanishStats(days, total, dayCount, now, v.Goal.Delta)
	v.Months, v.PeakMonth = spanishMonths(days, now.Year())
	v.PeakLabel = peakMonthLabel(v.Months, v.PeakMonth)
	return v
}

// aggregateSpanish sums sessions per day, returning the days oldest first and a
// lookup of seconds by ISO date.
func aggregateSpanish(entries []repo.SpanishLogEntry) ([]spanishDay, map[string]int) {
	byDate := map[string]int{}
	for _, e := range entries {
		byDate[e.Date] += e.Seconds
	}
	days := make([]spanishDay, 0, len(byDate))
	for iso, sec := range byDate {
		days = append(days, spanishDay{date: parseDate(iso), sec: sec})
	}
	sort.Slice(days, func(i, j int) bool { return days[i].date.Before(days[j].date) })
	return days, byDate
}

// spanishBand is progress through the current hours band toward the next mark.
func spanishBand(total int) templates.BandView {
	hours := float64(total) / 3600
	idx := 0
	for i, t := range dsLevels {
		if hours >= float64(t) {
			idx = i
		}
	}
	if idx+1 >= len(dsLevels) {
		return templates.BandView{
			AtMax:  true,
			ToNext: fmt.Sprintf("past %s hours", commaNum(dsLevels[idx])),
		}
	}
	prev, next := dsLevels[idx], dsLevels[idx+1]
	return templates.BandView{
		PrevLabel: fmt.Sprintf("%sh", commaNum(prev)),
		NextLabel: fmt.Sprintf("%sh", commaNum(next)),
		Pct:       int((hours - float64(prev)) / float64(next-prev) * 100),
		ToNext:    fmt.Sprintf("%s hours to %s", commaNum(next-int(hours)), commaNum(next)),
	}
}

// spanishStats are the six frontispiece figures: recent volume, streaks, the
// year's total and where it stands against the goal.
func spanishStats(days []spanishDay, total, dayCount int, now time.Time, goalDelta string) []templates.Stat {
	cur, longest := streaks(days, dateOnly(now))
	monthSec, yearSec := 0, 0
	for _, d := range days {
		if d.date.Year() == now.Year() {
			yearSec += d.sec
			if d.date.Month() == now.Month() {
				monthSec += d.sec
			}
		}
	}
	return []templates.Stat{
		{Label: "this month", Value: hoursShort(monthSec)},
		{Label: "daily average", Value: durShort(total / dayCount)},
		{Label: "current streak", Value: fmt.Sprintf("%dd", cur)},
		{Label: "longest streak", Value: fmt.Sprintf("%dd", longest)},
		{Label: "this year", Value: hoursShort(yearSec)},
		{Label: fmt.Sprintf("%dh goal", spanishGoalHours), Value: goalDelta},
	}
}

// spanishMonths tallies seconds per month for the year's density strip, plus the
// peak month's seconds so the bars scale against it in the same unit.
func spanishMonths(days []spanishDay, year int) ([12]int, int) {
	var months [12]int
	peak := 0
	for _, d := range days {
		if d.date.Year() != year {
			continue
		}
		m := int(d.date.Month()) - 1
		months[m] += d.sec
	}
	for _, sec := range months {
		if sec > peak {
			peak = sec
		}
	}
	return months, peak
}

func peakMonthLabel(months [12]int, peak int) string {
	if peak <= 0 {
		return ""
	}
	for i, sec := range months {
		if sec == peak {
			return fmt.Sprintf("peak %s", time.Month(i + 1).String()[:3])
		}
	}
	return ""
}

// spanishGoal builds the burn-up toward the goal's hour target.
func spanishGoal(days []spanishDay, total int, start, now time.Time) templates.GoalView {
	const w, h, pad = 720.0, 240.0, 6.0
	deadline := spanishGoalDeadline(now.Location())
	span := deadline.Sub(start).Hours()
	target := float64(spanishGoalHours)

	x := func(t time.Time) float64 { return w * t.Sub(start).Hours() / span }
	y := func(hours float64) float64 {
		if hours > target {
			hours = target
		}
		return (h - pad) - (h-2*pad)*(hours/target)
	}

	points := ""
	cum := 0
	for _, d := range days {
		cum += d.sec
		points += fmt.Sprintf("%.1f,%.1f ", x(d.date), y(float64(cum)/3600))
	}

	totalHours := float64(total) / 3600
	elapsed := now.Sub(start).Hours() / 24
	remaining := deadline.Sub(now).Hours() / 24
	expected := target * elapsed / (span / 24)
	ahead := totalHours - expected

	g := templates.GoalView{
		Head:         fmt.Sprintf("%d hours by %s %d", spanishGoalHours, spanishGoalMonth, spanishGoalYear),
		Reached:      totalHours >= target,
		ActualPoints: points,
		PacePoints:   fmt.Sprintf("%.1f,%.1f %.1f,%.1f", x(start), y(0), w, y(target)),
		NowX:         fmt.Sprintf("%.1f", x(now)),
		NowY:         fmt.Sprintf("%.1f", y(totalHours)),
	}
	if g.Reached {
		g.Verdict = "reached"
		g.Delta = fmt.Sprintf("on %dh", spanishGoalHours)
		g.Pace = fmt.Sprintf("%.0f hours in, target cleared", totalHours)
		return g
	}
	if ahead >= 0 {
		g.Verdict = fmt.Sprintf("%.0fh ahead", ahead)
		g.Delta = fmt.Sprintf("+%.0fh", ahead)
	} else {
		g.Verdict = fmt.Sprintf("%.0fh behind", -ahead)
		g.Delta = fmt.Sprintf("-%.0fh", -ahead)
	}
	if remaining > 0 {
		perDay := (target - totalHours) / remaining * 3600
		g.Pace = fmt.Sprintf("%s a day to finish on time, %s so far", durShort(int(perDay)), g.Verdict)
	}
	return g
}

// spanishCalendar builds the weekly heatmap columns from start to today, plus a
// year marker per run of columns sharing a year (dated by each column's Monday).
func spanishCalendar(secByDate map[string]int, start, today time.Time) templates.CalendarView {
	gridStart := start.AddDate(0, 0, -((int(start.Weekday()) + 6) % 7)) // back to Monday
	var cal templates.CalendarView
	for col := gridStart; !col.After(today); col = col.AddDate(0, 0, 7) {
		var week templates.CalWeek
		for i := 0; i < 7; i++ {
			d := col.AddDate(0, 0, i)
			if d.Before(start) || d.After(today) {
				continue
			}
			sec := secByDate[d.Format("2006-01-02")]
			week.Days[i] = templates.CalDay{
				InRange: true,
				Class:   calClass(sec),
				Title:   fmt.Sprintf("%s ; %s", fmtDay(d), durShort(sec)),
			}
		}
		cal.Weeks = append(cal.Weeks, week)

		label := fmt.Sprintf("%d", col.Year())
		if n := len(cal.Years); n > 0 && cal.Years[n-1].Label == label {
			cal.Years[n-1].Weeks++
		} else {
			cal.Years = append(cal.Years, templates.YearSpan{Label: label, Weeks: 1})
		}
	}
	return cal
}

// calClass shades a cell by the day's logged time, from an empty rule to full
// accent through two tints in between.
func calClass(sec int) string {
	switch {
	case sec <= 0:
		return "bg-rule"
	case sec < 900:
		return "bg-accent-tint"
	case sec < 1800:
		return "bg-[#f0a39b]"
	case sec < 3600:
		return "bg-[#d0574c]"
	default:
		return "bg-accent"
	}
}

// crossingDate is the first day the running total reached threshold hours.
func crossingDate(days []spanishDay, threshold int) time.Time {
	if threshold == 0 && len(days) > 0 {
		return days[0].date
	}
	cum := 0
	for _, d := range days {
		cum += d.sec
		if cum/3600 >= threshold {
			return d.date
		}
	}
	return time.Time{}
}

// spanishRecords are the rail's all-time bests: biggest day and month, the
// longest streak and how many days have been logged.
func spanishRecords(days []spanishDay) []templates.Stat {
	bestDay := 0
	for _, d := range days {
		if d.sec > bestDay {
			bestDay = d.sec
		}
	}
	months := map[int]int{} // seconds keyed by year*12+month
	bestMonth := 0
	for _, d := range days {
		key := d.date.Year()*12 + int(d.date.Month())
		months[key] += d.sec
		if months[key] > bestMonth {
			bestMonth = months[key]
		}
	}
	_, longest := streaks(days, days[len(days)-1].date)
	return []templates.Stat{
		{Label: "Best day", Value: durShort(bestDay)},
		{Label: "Best month", Value: hoursShort(bestMonth)},
		{Label: "Longest streak", Value: fmt.Sprintf("%d days", longest)},
		{Label: "Days logged", Value: fmt.Sprintf("%d", len(days))},
	}
}

// spanishAverages are the rail's mean input per calendar day, per active day and
// per week, plus the share of days with any input.
func spanishAverages(total, dayCount, activeDays int) []templates.Stat {
	stats := []templates.Stat{
		{Label: "Per day", Value: durShort(total / dayCount)},
	}
	if activeDays > 0 {
		stats = append(stats, templates.Stat{Label: "Per active day", Value: durShort(total / activeDays)})
	}
	stats = append(stats,
		templates.Stat{Label: "Per week", Value: durShort(total / dayCount * 7)},
		templates.Stat{Label: "Consistency", Value: fmt.Sprintf("%d%%", activeDays*100/dayCount)},
	)
	return stats
}

// streaks returns the current and longest runs of consecutive logged days. The
// current run counts only if it reaches yesterday or today.
func streaks(days []spanishDay, today time.Time) (current, longest int) {
	logged := map[string]bool{}
	for _, d := range days {
		logged[d.date.Format("2006-01-02")] = true
	}
	run := 0
	for _, d := range days {
		prev := d.date.AddDate(0, 0, -1).Format("2006-01-02")
		if logged[prev] {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
	}
	last := days[len(days)-1].date
	if daysBetween(last, today) <= 2 { // still current if it reaches today or yesterday
		current = run
	}
	return current, longest
}

// commaNum groups a whole number with thousands commas, e.g. 1500 -> "1,500".
func commaNum(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	head := len(s) % 3
	out := s[:head]
	for i := head; i < len(s); i += 3 {
		if out != "" {
			out += ","
		}
		out += s[i : i+3]
	}
	return out
}

// hoursShort renders seconds as whole hours, e.g. "18h".
func hoursShort(sec int) string {
	return fmt.Sprintf("%dh", int(math.Round(float64(sec)/3600)))
}

// durShort renders seconds as hours and minutes, dropping a zero part, e.g.
// "1h 40m", "42m".
func durShort(sec int) string {
	if sec < 0 {
		sec = 0
	}
	h := sec / 3600
	m := int(math.Round(float64(sec%3600) / 60))
	if m == 60 {
		h++
		m = 0
	}
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	default:
		return fmt.Sprintf("%dm", m)
	}
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// spanishDayCount is the day number of the run: calendar days from the first
// logged day to now, counting both ends. Zero when nothing is logged.
func spanishDayCount(days []spanishDay, now time.Time) int {
	if len(days) == 0 {
		return 0
	}
	return int(dateOnly(now).Sub(days[0].date).Hours()/24) + 1
}

func fmtDay(t time.Time) string {
	return t.Format("2 Jan 2006")
}
