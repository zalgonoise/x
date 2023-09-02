package schedule

import (
	"errors"
	"testing"

	"github.com/zalgonoise/parse"
)

func TestParser(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input string
		err   error
	}{
		{
			name:  "Success/Simple/AllStar",
			input: "* * * * *",
		},
		{
			name:  "Success/Simple/EveryMinuteZero",
			input: "0 * * * *",
		},
		{
			name:  "Success/Simple/LargeMinute",
			input: "50 * * * *",
		},
		{
			name:  "Success/Simple/Every3rdMinute",
			input: "*/3 * * * *",
		},
		{
			name:  "Success/Simple/EveryMinuteFrom0Through3",
			input: "0-3 * * * *",
		},
		{
			name:  "Success/Simple/EveryMinuteFrom0Through3And5And7",
			input: "0-3,5,7 * * * *",
		},
		{
			name:  "Success/Simple/EveryMinuteLiteral",
			input: "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59 * * * *",
		},
		{
			name:  "Success/Simple/EveryHourRange",
			input: "0 0-23 * * *",
		},
		{
			name:  "Success/Simple/EveryDayRange",
			input: "0 0 1-31 * *",
		},
		{
			name:  "Success/Simple/EveryMonthNumericLiteral",
			input: "0 0 1 1,2,3,4,5,6,7,8,9,10,11,12 *",
		},
		{
			name:  "Success/Simple/EveryMonthStringLiteral",
			input: "0 0 1 jan,Feb,MAR,aPR,maY,JuN,JUl,AUG,sep,oct,nov,dec *",
		},
		{
			name:  "Success/Simple/EveryWeekdayNumericLiteralSundayFirst",
			input: "0 0 * * 0,1,2,3,4,5,6",
		},
		{
			name:  "Success/Simple/EveryWeekdayNumericLiteralSundayLast",
			input: "0 0 * * 1,2,3,4,5,6,7",
		},
		{
			name:  "Success/Simple/EveryWeekdayStringLiteral",
			input: "0 0 * * sun,Mon,TUE,wED,thU,FrI,sAt",
		},
		{
			name:  "Success/Overrides/reboot",
			input: "@reboot",
		},
		{
			name:  "Success/Overrides/hourly",
			input: "@hourly",
		},
		{
			name:  "Success/Overrides/daily",
			input: "@daily",
		},
		{
			name:  "Success/Overrides/weekly",
			input: "@weekly",
		},
		{
			name:  "Success/Overrides/monthly",
			input: "@monthly",
		},
		{
			name:  "Success/Overrides/annually",
			input: "@annually",
		},
		{
			name:  "Success/Overrides/yearly",
			input: "@yearly",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// TODO: replace stringer implementation with a schedule data structure; currently serves for validation testing
			str, err := parse.Run([]byte(testcase.input), initState, initParse, process)
			if !errors.Is(err, testcase.err) {
				t.Errorf("output mismatch: wants: %v ; got %v", testcase.err, err)
			}

			t.Log(str)
		})
	}
}
