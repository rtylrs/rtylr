package model

import "math"

func linear(y, w, b, x []float32, rows, cols int) {
	for r := 0; r < rows; r++ {
		var acc float32
		base := r * cols
		for c := 0; c < cols; c++ {
			acc += w[base+c] * x[c]
		}
		if b != nil {
			acc += b[r]
		}
		y[r] = acc
	}
}

func dot(a, b []float32) float32 {
	var acc float32
	for i := range a {
		acc += a[i] * b[i]
	}
	return acc
}

func addInto(dst, src []float32) {
	for i := range dst {
		dst[i] += src[i]
	}
}

func layernorm(y, x, w, b []float32, eps float32) {
	n := float32(len(x))
	var mean float32
	for _, v := range x {
		mean += v
	}
	mean /= n
	var variance float32
	for _, v := range x {
		d := v - mean
		variance += d * d
	}
	variance /= n
	inv := 1 / float32(math.Sqrt(float64(variance+eps)))
	for i, v := range x {
		y[i] = (v-mean)*inv*w[i] + b[i]
	}
}

func softmaxInPlace(a []float32) {
	if len(a) == 0 {
		return
	}
	m := a[0]
	for _, v := range a[1:] {
		if v > m {
			m = v
		}
	}
	var sum float32
	for i, v := range a {
		d := v - m
		if math.IsNaN(float64(d)) {
			d = 0
		}
		e := float32(math.Exp(float64(d)))
		a[i] = e
		sum += e
	}
	if sum == 0 {
		uniform := 1 / float32(len(a))
		for i := range a {
			a[i] = uniform
		}
		return
	}
	inv := 1 / sum
	for i := range a {
		a[i] *= inv
	}
}

func reluInPlace(a []float32) {
	for i, v := range a {
		if v < 0 {
			a[i] = 0
		}
	}
}
