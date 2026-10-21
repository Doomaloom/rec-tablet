package common

import (
	"time"
)

type Benchmark struct {
	SkillID     string
	Limit       time.Duration
	Description string
}

var benchmarks = map[string]Benchmark{
	"Splash7:10": {SkillID: "Splash7:10", Limit: 3 * time.Minute, Description: "100 m · 3:00 limit (100 yd · 2:40)"},
	"Splash8:9":  {SkillID: "Splash8:9", Limit: 6 * time.Minute, Description: "200 m · 6:00 limit (200 yd · 5:20)"},
	"Splash9:11": {SkillID: "Splash9:11", Limit: 9 * time.Minute, Description: "300 m · 9:00 limit (300 yd · 8:00)"},
}

func BenchmarkFor(skillID string) (Benchmark, bool) {
	benchmark, ok := benchmarks[skillID]
	return benchmark, ok
}
