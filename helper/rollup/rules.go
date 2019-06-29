package rollup

import (
	"fmt"
	"regexp"
	"time"

	"github.com/lomik/graphite-clickhouse/helper/point"
)

type Retention struct {
	Age       uint32 `json:"age"`
	Precision uint32 `json:"precision"`
}

type Pattern struct {
	Regexp    string         `json:"regexp"`
	Function  string         `json:"function"`
	Retention []Retention    `json:"retention"`
	aggr      *Aggr          `json:"-"`
	re        *regexp.Regexp `json:"-"`
}

type Rules struct {
	Pattern []Pattern `json:"pattern"`
	Updated int64     `json:"updated"`
}

// should never be used in real conditions
var superDefaultRetention = []Retention{
	Retention{Age: 0, Precision: 60},
}

const superDefaultFunction = "avg"

func (p *Pattern) compile() error {
	var err error
	if p.Regexp != "" {
		p.re, err = regexp.Compile(p.Regexp)
		if err != nil {
			return err
		}
	}

	if p.Function != "" {
		var exists bool
		p.aggr, exists = AggrMap[p.Function]

		if !exists {
			return fmt.Errorf("unknown function %#v", p.Function)
		}
	}

	return nil
}

func (r *Rules) compile() error {
	if r.Pattern == nil {
		r.Pattern = make([]Pattern, 0)
	}

	for _, p := range r.Pattern {
		if err := p.compile(); err != nil {
			return err
		}
	}

	return nil
}

// func (r *Rules) addDefaultPrecision(p uint32) {
// 	for _, pt := range append(r.Pattern, r.Default) {
// 		hasZeroAge := false
// 		for _, rt := range pt.Retention {
// 			if rt.Age == 0 {
// 				hasZeroAge = true
// 			}
// 		}

// 		if !hasZeroAge {
// 			pt.Retention = append([]*Retention{&Retention{0, p}}, pt.Retention...)
// 		}
// 	}
// }

// Match returns rollup rules for metric
func (r *Rules) match(metric string) (*Aggr, []Retention) {
	var ag *Aggr
	var rt []Retention

	for _, p := range r.Pattern {
		if p.re == nil || p.re.MatchString(metric) {
			if ag == nil && p.aggr != nil {
				ag = p.aggr
			}
			if len(rt) == 0 && len(p.Retention) > 0 {
				rt = p.Retention
			}

			if ag != nil && len(rt) > 0 {
				return ag, rt
			}
		}
	}

	if ag == nil {
		ag = AggrMap["avg"]
	}
	if len(rt) == 0 {
		rt = superDefaultRetention
	}

	return ag, rt
}

func (r *Rules) Step(metric string, from uint32) (uint32, error) {
	_, rt := r.match(metric)
	now := uint32(time.Now().Unix())

	if len(rt) == 0 {
		return 0, fmt.Errorf("rollup retention not found for metric %#v", metric)
	}

	for i := range rt {
		if i == len(rt)-1 || from+rt[i+1].Age > now {
			return rt[i].Precision, nil
		}
	}
	return rt[len(rt)-1].Precision, nil
}

func doMetricPrecision(points []point.Point, precision uint32, aggr *Aggr) []point.Point {
	l := len(points)
	var i, n int
	// i - current position of iterator
	// n - position of the first record with time rounded to precision

	if l == 0 {
		return points
	}

	// set first point time
	t := points[0].Time
	t = t - (t % precision)
	points[0].Time = t

	for i = 1; i < l; i++ {
		t = points[i].Time
		t = t - (t % precision)
		points[i].Time = t

		if points[n].Time == t {
			points[i].MetricID = 0
		} else {
			if i > n+1 {
				points[n].Value = aggr.Do(points[n:i])
			}
			n = i
		}
	}
	if i > n+1 {
		points[n].Value = aggr.Do(points[n:i])
	}

	return point.CleanUp(points)
}

// RollupMetric rolling up list of points of ONE metric sorted by key "time"
// returns (new points slice, precision)
func (r *Rules) RollupMetric(metricName string, fromTimestamp uint32, points []point.Point) ([]point.Point, uint32, error) {
	// pp.Println(points)

	l := len(points)
	if l == 0 {
		return points, 1, nil
	}

	now := uint32(time.Now().Unix())
	ag, rt := r.match(metricName)
	precision := uint32(1)

	if len(rt) == 0 {
		return points, 0, fmt.Errorf("rollup retention not found for metric %#v", metricName)
	}
	if ag == nil {
		return points, 0, fmt.Errorf("rollup function not found for metric %#v", metricName)
	}

	for _, retention := range rt {
		if fromTimestamp+retention.Age > now && retention.Age != 0 {
			break
		}

		points = doMetricPrecision(points, retention.Precision, ag)
		precision = retention.Precision
	}

	// pp.Println(points)
	return points, precision, nil
}
