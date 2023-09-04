package schedule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input string
		wants CronSchedule
		err   error
	}{
		{
			name:  "Success/Simple/AllStar",
			input: "* * * * *",
			wants: CronSchedule{
				Loc:      time.UTC,
				min:      everytime{},
				hour:     everytime{},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek:  everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryMinuteZero",
			input: "0 * * * *",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour:     everytime{},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek:  everytime{},
			},
		},
		{
			name:  "Success/Simple/LargeMinute",
			input: "50 * * * *",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      50,
				},
				hour:     everytime{},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek:  everytime{},
			},
		},
		{
			name:  "Success/Simple/Every3rdMinute",
			input: "*/3 * * * *",
			wants: CronSchedule{
				Loc: time.UTC,
				min: stepSchedule{
					maximum: 59,
					steps:   []int{0, 3, 6, 9, 12, 15, 18, 21, 24, 27, 30, 33, 36, 39, 42, 45, 48, 51, 54, 57},
				},
				hour:     everytime{},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek:  everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryMinuteFrom0Through3",
			input: "0-3 * * * *",
			wants: CronSchedule{
				Loc: time.UTC,
				min: rangeSchedule{
					maximum: 59,
					from:    0,
					to:      3,
				},
				hour:     everytime{},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek:  everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryMinuteFrom0Through3And5And7",
			input: "0-3,5,7 * * * *",
			wants: CronSchedule{
				Loc: time.UTC,
				min: stepSchedule{
					maximum: 59,
					steps:   []int{0, 1, 2, 3, 5, 7},
				},
				hour:     everytime{},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek:  everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryMinuteLiteral",
			input: "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59 * * * *",
			wants: CronSchedule{
				Loc: time.UTC,
				min: stepSchedule{
					maximum: 59,
					steps: []int{
						0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59,
					},
				},
				hour:     everytime{},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek:  everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryHourRange",
			input: "0 0-23 * * *",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: rangeSchedule{
					maximum: 23,
					from:    0,
					to:      23,
				},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek:  everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryDayRange",
			input: "0 0 1-31 * *",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: rangeSchedule{
					maximum: 31,
					from:    1,
					to:      31,
				},
				month:   everytime{},
				dayWeek: everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryMonthNumericLiteral",
			input: "0 0 1 1,2,3,4,5,6,7,8,9,10,11,12 *",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: fixedSchedule{
					maximum: 31,
					at:      1,
				},
				month: stepSchedule{
					maximum: 12,
					steps:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				},
				dayWeek: everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryMonthStringLiteral",
			input: "0 0 1 jan,Feb,MAR,aPR,maY,JuN,JUl,AUG,sep,oct,nov,dec *",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: fixedSchedule{
					maximum: 31,
					at:      1,
				},
				month: stepSchedule{
					maximum: 12,
					steps:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				},
				dayWeek: everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryWeekdayNumericLiteralSundayFirst",
			input: "0 0 * * 0,1,2,3,4,5,6",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek: stepSchedule{
					maximum: 7,
					steps:   []int{0, 1, 2, 3, 4, 5, 6},
				},
			},
		},
		{
			name:  "Success/Simple/EveryWeekdayNumericLiteralSundayLast",
			input: "0 0 * * 1,2,3,4,5,6,7",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek: stepSchedule{
					maximum: 7,
					steps:   []int{0, 1, 2, 3, 4, 5, 6},
				},
			},
		},
		{
			name:  "Success/Simple/EveryWeekdayStringLiteral",
			input: "0 0 * * sun,Mon,TUE,wED,thU,FrI,sAt",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek: stepSchedule{
					maximum: 7,
					steps:   []int{0, 1, 2, 3, 4, 5, 6},
				},
			},
		},
		{
			name:  "Success/Overrides/reboot",
			input: "@reboot",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour:     everytime{},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek:  everytime{},
			},
		},
		{
			name:  "Success/Overrides/hourly",
			input: "@hourly",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour:     everytime{},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek:  everytime{},
			},
		},
		{
			name:  "Success/Overrides/daily",
			input: "@daily",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek:  everytime{},
			},
		},
		{
			name:  "Success/Overrides/weekly",
			input: "@weekly",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek: fixedSchedule{
					maximum: 6,
					at:      0,
				},
			},
		},
		{
			name:  "Success/Overrides/monthly",
			input: "@monthly",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: fixedSchedule{
					maximum: 31,
					at:      1,
				},
				month:   everytime{},
				dayWeek: everytime{},
			},
		},
		{
			name:  "Success/Overrides/annually",
			input: "@annually",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: fixedSchedule{
					maximum: 31,
					at:      1,
				},
				month: fixedSchedule{
					maximum: 12,
					at:      1,
				},
				dayWeek: everytime{},
			},
		},
		{
			name:  "Success/Overrides/yearly",
			input: "@yearly",
			wants: CronSchedule{
				Loc: time.UTC,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: fixedSchedule{
					maximum: 31,
					at:      1,
				},
				month: fixedSchedule{
					maximum: 12,
					at:      1,
				},
				dayWeek: everytime{},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			cron, err := New(
				WithSchedule(testcase.input),
				WithLocation(time.UTC),
			)

			require.ErrorIs(t, err, testcase.err)
			require.Equal(t, testcase.wants, cron)
		})
	}
}
