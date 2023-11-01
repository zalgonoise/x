package cronlex

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/parse"
)

func TestParser(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input string
		wants Schedule
		err   error
	}{
		{
			name:  "Success/Simple/AllStar",
			input: "* * * * *",
			wants: Schedule{
				Min:      Everytime{},
				Hour:     Everytime{},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek:  Everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryMinuteZero",
			input: "0 * * * *",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour:     Everytime{},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek:  Everytime{},
			},
		},
		{
			name:  "Success/Simple/LargeMinute",
			input: "50 * * * *",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      50,
				},
				Hour:     Everytime{},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek:  Everytime{},
			},
		},
		{
			name:  "Success/Simple/Every3rdMinute",
			input: "*/3 * * * *",
			wants: Schedule{
				Min: StepSchedule{
					maximum: 59,
					steps:   []int{0, 3, 6, 9, 12, 15, 18, 21, 24, 27, 30, 33, 36, 39, 42, 45, 48, 51, 54, 57},
				},
				Hour:     Everytime{},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek:  Everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryMinuteFrom0Through3",
			input: "0-3 * * * *",
			wants: Schedule{
				Min: RangeSchedule{
					maximum: 59,
					from:    0,
					to:      3,
				},
				Hour:     Everytime{},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek:  Everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryMinuteFrom0Through3And5And7",
			input: "0-3,5,7 * * * *",
			wants: Schedule{
				Min: StepSchedule{
					maximum: 59,
					steps:   []int{0, 1, 2, 3, 5, 7},
				},
				Hour:     Everytime{},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek:  Everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryMinuteLiteral",
			input: "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59 * * * *",
			wants: Schedule{
				Min: StepSchedule{
					maximum: 59,
					steps: []int{
						0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59,
					},
				},
				Hour:     Everytime{},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek:  Everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryHourRange",
			input: "0 0-23 * * *",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour: RangeSchedule{
					maximum: 23,
					from:    0,
					to:      23,
				},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek:  Everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryDayRange",
			input: "0 0 1-31 * *",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour: FixedSchedule{
					maximum: 23,
					at:      0,
				},
				DayMonth: RangeSchedule{
					maximum: 31,
					from:    1,
					to:      31,
				},
				Month:   Everytime{},
				DayWeek: Everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryMonthNumericLiteral",
			input: "0 0 1 1,2,3,4,5,6,7,8,9,10,11,12 *",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour: FixedSchedule{
					maximum: 23,
					at:      0,
				},
				DayMonth: FixedSchedule{
					maximum: 31,
					at:      1,
				},
				Month: StepSchedule{
					maximum: 12,
					steps:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				},
				DayWeek: Everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryMonthStringLiteral",
			input: "0 0 1 jan,Feb,MAR,aPR,maY,JuN,JUl,AUG,sep,oct,nov,dec *",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour: FixedSchedule{
					maximum: 23,
					at:      0,
				},
				DayMonth: FixedSchedule{
					maximum: 31,
					at:      1,
				},
				Month: StepSchedule{
					maximum: 12,
					steps:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				},
				DayWeek: Everytime{},
			},
		},
		{
			name:  "Success/Simple/EveryWeekdayNumericLiteralSundayFirst",
			input: "0 0 * * 0,1,2,3,4,5,6",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour: FixedSchedule{
					maximum: 23,
					at:      0,
				},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek: StepSchedule{
					maximum: 7,
					steps:   []int{0, 1, 2, 3, 4, 5, 6},
				},
			},
		},
		{
			name:  "Success/Simple/EveryWeekdayNumericLiteralSundayLast",
			input: "0 0 * * 1,2,3,4,5,6,7",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour: FixedSchedule{
					maximum: 23,
					at:      0,
				},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek: StepSchedule{
					maximum: 7,
					steps:   []int{0, 1, 2, 3, 4, 5, 6},
				},
			},
		},
		{
			name:  "Success/Simple/EveryWeekdayStringLiteral",
			input: "0 0 * * sun,Mon,TUE,wED,thU,FrI,sAt",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour: FixedSchedule{
					maximum: 23,
					at:      0,
				},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek: StepSchedule{
					maximum: 7,
					steps:   []int{0, 1, 2, 3, 4, 5, 6},
				},
			},
		},
		{
			name:  "Success/Overrides/reboot",
			input: "@reboot",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour:     Everytime{},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek:  Everytime{},
			},
		},
		{
			name:  "Success/Overrides/hourly",
			input: "@hourly",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour:     Everytime{},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek:  Everytime{},
			},
		},
		{
			name:  "Success/Overrides/daily",
			input: "@daily",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour: FixedSchedule{
					maximum: 23,
					at:      0,
				},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek:  Everytime{},
			},
		},
		{
			name:  "Success/Overrides/weekly",
			input: "@weekly",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour: FixedSchedule{
					maximum: 23,
					at:      0,
				},
				DayMonth: Everytime{},
				Month:    Everytime{},
				DayWeek: FixedSchedule{
					maximum: 6,
					at:      0,
				},
			},
		},
		{
			name:  "Success/Overrides/monthly",
			input: "@monthly",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour: FixedSchedule{
					maximum: 23,
					at:      0,
				},
				DayMonth: FixedSchedule{
					maximum: 31,
					at:      1,
				},
				Month:   Everytime{},
				DayWeek: Everytime{},
			},
		},
		{
			name:  "Success/Overrides/annually",
			input: "@annually",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour: FixedSchedule{
					maximum: 23,
					at:      0,
				},
				DayMonth: FixedSchedule{
					maximum: 31,
					at:      1,
				},
				Month: FixedSchedule{
					maximum: 12,
					at:      1,
				},
				DayWeek: Everytime{},
			},
		},
		{
			name:  "Success/Overrides/yearly",
			input: "@yearly",
			wants: Schedule{
				Min: FixedSchedule{
					maximum: 59,
					at:      0,
				},
				Hour: FixedSchedule{
					maximum: 23,
					at:      0,
				},
				DayMonth: FixedSchedule{
					maximum: 31,
					at:      1,
				},
				Month: FixedSchedule{
					maximum: 12,
					at:      1,
				},
				DayWeek: Everytime{},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			cron, err := parse.Run([]byte(testcase.input), StateFunc, ParseFunc, ProcessFunc)

			require.ErrorIs(t, err, testcase.err)
			require.Equal(t, testcase.wants, cron)
		})
	}
}
