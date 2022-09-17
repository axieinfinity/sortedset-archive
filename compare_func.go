package sortedset

import "math"

type compareScoreFunc func(float64, float64) bool

var greaterThan compareScoreFunc = func(f, f1 float64) bool {
	return f > f1
}
var greaterThanOrEqual compareScoreFunc = func(f, f1 float64) bool {
	return f > f1 || math.Abs(f-f1) < eps
}
var lesserThan compareScoreFunc = func(f, f1 float64) bool {
	return f < f1
}
var lesserThanOrEqual compareScoreFunc = func(f, f1 float64) bool {
	return f < f1 || math.Abs(f-f1) < eps
}
