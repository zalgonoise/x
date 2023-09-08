package schedule

import (
	"context"
	"time"

	"github.com/zalgonoise/parse"
	"github.com/zalgonoise/x/cfg"
)

type resolver interface {
	Resolve(value int) int
}

type Scheduler interface {
	Next(ctx context.Context, now time.Time) time.Time
}

type CronSchedule struct {
	Loc *time.Location

	min      resolver
	hour     resolver
	dayMonth resolver
	month    resolver
	dayWeek  resolver
}

func New(options ...cfg.Option[SchedulerConfig]) (Scheduler, error) {
	config := cfg.New(options...)

	cron, err := newScheduler(config)
	if err != nil {
		return noOpScheduler{}, err
	}

	if config.metrics != nil {
		cron = withMetrics(cron, config.metrics)
	}

	if config.logger != nil {
		cron = withLogs(cron, config.logger)
	}

	if config.tracer != nil {
		cron = withTrace(cron, config.tracer)
	}

	return cron, nil
}

func newScheduler(config SchedulerConfig) (Scheduler, error) {
	// parse cron string
	cron, err := parse.Run([]byte(config.cronString), initState, initParse, process)
	if err != nil {
		return noOpScheduler{}, err
	}

	// set location if provided
	if config.loc != nil {
		cron.Loc = config.loc
	}

	return cron, nil
}

func (s CronSchedule) Next(_ context.Context, t time.Time) time.Time {
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

type noOpScheduler struct{}

func (s noOpScheduler) Next(_ context.Context, _ time.Time) time.Time {
	return time.Time{}
}
