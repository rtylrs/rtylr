package domain

import "errors"

// Points scales a project can estimate story points on (T-381). T-shirt sizes
// are stored as numbers so cycle totals and velocity can add them up.
const (
	PointsScaleFibonacci = "fibonacci"
	PointsScaleLinear    = "linear"
	PointsScaleTShirt    = "tshirt"
)

// ErrInvalidPointsScale rejects a scale outside the three above.
var ErrInvalidPointsScale = errors.New("points_scale must be fibonacci, linear or tshirt")

// ErrInvalidStoryPoints rejects a value that is not on the project's scale.
var ErrInvalidStoryPoints = errors.New("story_points must be a value on the project's points scale")

var pointsScaleValues = map[string][]int32{
	PointsScaleFibonacci: {1, 2, 3, 5, 8, 13},
	PointsScaleLinear:    {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	PointsScaleTShirt:    {1, 2, 3, 5, 8},
}

// IsValidPointsScale reports whether scale is one of the known scales.
func IsValidPointsScale(scale string) bool {
	_, ok := pointsScaleValues[scale]
	return ok
}

// IsValidStoryPoints reports whether points is a value on scale. An unknown or
// empty scale is read as Fibonacci, the default.
func IsValidStoryPoints(scale string, points int32) bool {
	values, ok := pointsScaleValues[scale]
	if !ok {
		values = pointsScaleValues[PointsScaleFibonacci]
	}
	for _, v := range values {
		if v == points {
			return true
		}
	}
	return false
}
