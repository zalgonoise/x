package schedule

import "time"

type resolver interface {
	Resolve(value int) int
}

type cronSchedule struct {
	Loc *time.Location

	min      resolver
	hour     resolver
	dayMonth resolver
	month    resolver
	dayWeek  resolver
}

func (s cronSchedule) Next(t time.Time) time.Time {
	t = t.Truncate(time.Minute)
	year, month, day := t.Date()
	hour := t.Hour()
	minute := t.Minute()

	nextMinute := s.min.Resolve(minute) + 1
	nextHour := s.hour.Resolve(hour)
	nextDay := s.dayMonth.Resolve(day)
	nextMonth := s.month.Resolve(int(month))

	if hour+nextHour > 24 {
		nextDay--
	}

	dayOfMonthTime := time.Date(
		year,
		month+time.Month(nextMonth),
		day+nextDay,
		hour+nextHour,
		minute+nextMinute,
		0, 0, s.Loc,
	)

	// short circuit if unset or star '*'
	if _, ok := (s.dayWeek).(everytime); s.dayWeek == nil || ok {
		return dayOfMonthTime
	}

	futureWeekday := dayOfMonthTime.Truncate(time.Hour * 24 * time.Duration(day))
	nextWeekday := s.dayWeek.Resolve(int(futureWeekday.Weekday()))

	if hour+nextHour > 24 {
		nextWeekday--
	}

	weekdayTime := time.Date(
		year,
		month+time.Month(nextMonth),
		futureWeekday.Day()+nextWeekday,
		hour+nextHour,
		minute+nextMinute,
		0, 0, s.Loc,
	)

	if dayOfMonthTime.Before(weekdayTime) {
		return dayOfMonthTime
	}

	return weekdayTime
}
