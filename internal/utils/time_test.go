package utils

import (
	"fmt"
	"testing"
	"time"
)

var (
	testSanitizeCases    = []string{"3h24m40s20ms", "3h24m20.323232323s", "59m", "59m59s", "2h28m", "2h28m59s59ms"}
	resultsSanitizeCases = []string{"3 hours 25 minutes", "3 hours 24 minutes", "59 minutes", "1 hour", "2 hours 28 minutes", "2 hours 29 minutes"}
)

func TestSanitizingTimeFormat(t *testing.T) {
	for i := 0; i < len(testSanitizeCases); i++ {
		t.Run(fmt.Sprintf("Test Case %d", i), func(t *testing.T) {
			dur, err := time.ParseDuration(testSanitizeCases[i])
			if err != nil {
				t.Fatalf("Unable to parse your test duration")
			}

			s := SanitizeDuration(dur)

			fmt.Println(s)

			if s != resultsSanitizeCases[i] {
				t.Fail()
			}
		})
	}
}
