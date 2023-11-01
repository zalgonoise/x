package schedule

import (
	"context"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/parse"
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
		cron = schedulerWithMetrics(cron, config.metrics)
	}

	if config.handler != nil {
		cron = schedulerWithLogs(cron, config.handler)
	}

	if config.tracer != nil {
		cron = schedulerWithTrace(cron, config.tracer)
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
	year, month, day := t.Date()
	hour := t.Hour()
	minute := t.Minute()

	nextMinute := s.min.Resolve(minute) + 1
	nextHour := s.hour.Resolve(hour)
	nextDay := s.dayMonth.Resolve(day)
	nextMonth := s.month.Resolve(int(month))

	// time.Date automatically normalizes overflowing values in the context of dates
	// (e.g. a result containing 27 hours is 3 AM on the next day)
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

	curWeekday := dayOfMonthTime.Weekday()
	nextWeekday := s.dayWeek.Resolve(int(curWeekday))

	weekdayTime := time.Date(
		dayOfMonthTime.Year(),
		dayOfMonthTime.Month(),
		dayOfMonthTime.Day()+nextWeekday,
		dayOfMonthTime.Hour(),
		dayOfMonthTime.Minute(),
		0, 0, s.Loc,
	)

	return weekdayTime
}

func NoOp() Scheduler {
	return noOpScheduler{}
}

type noOpScheduler struct{}

func (s noOpScheduler) Next(_ context.Context, _ time.Time) time.Time {
	return time.Time{}
}
