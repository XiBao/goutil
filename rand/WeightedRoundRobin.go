package rand

import (
	"fmt"
	"math/rand"
)

// WeightedRoundRobin Weight Round Robin Alghoritm
type WeightedRoundRobin struct {
	pdf      []int
	index100 int
}

// NewWeightedRoundRobin instantiate weight round robin
func NewWeightedRoundRobin(pdf []int) (wrr *WeightedRoundRobin, err error) {
	r := 0
	max100 := -1
	for i, v := range pdf {
		if v == 100 {
			max100 = i
		}
		r += v
		if v < 0 || v > 100 {
			return wrr, fmt.Errorf("value %v out of range [0;100]", v)
		}
	}
	if r != 100 {
		return wrr, fmt.Errorf("sum of pdf elements must be equal to 100 perent")
	}
	// rand.Seed(time.Now().UnixNano())
	wrr = &WeightedRoundRobin{
		pdf:      pdf,
		index100: max100,
	}
	return wrr, nil
}

// PickVector returns slice shuffled by pdf distribution.
// The item with the highest probability will occur more often
// at the position that has the highest probability in the PDF
// see README.md
func (w *WeightedRoundRobin) PickVector() (indexes []int) {
	if w.index100 != -1 {
		return w.handle100()
	}

	pdf := make([]int, len(w.pdf))
	copy(pdf, w.pdf)
	balance := 100
	for i := 0; i < len(pdf); i++ {
		cdf := w.getCDF(pdf)
		index := w.pick(cdf, balance)
		indexes = append(indexes, index)

		balance -= pdf[index]
		pdf[index] = 0
	}
	return indexes
}

// Pick returns one index with probability given by pdf
// see README.md
func (w *WeightedRoundRobin) Pick() int {
	cdf := w.getCDF(w.pdf)
	return w.pick(cdf, 100)
}

// pick one index
func (w *WeightedRoundRobin) pick(cdf []int, n int) int {
	r := rand.Intn(n)
	index := 0
	for r >= cdf[index] {
		index++
	}
	return index
}

func (w *WeightedRoundRobin) getCDF(pdf []int) (cdf []int) {
	// prepare cdf
	cdf = make([]int, len(pdf))
	cdf[0] = pdf[0]
	for i := 1; i < len(pdf); i++ {
		cdf[i] = cdf[i-1] + pdf[i]
	}
	return cdf
}

// there is no reason to calculate CDF and recompute PDF's if some field has 100%
func (w *WeightedRoundRobin) handle100() (indexes []int) {
	for i := 0; i < len(w.pdf); i++ {
		indexes = append(indexes, i)
	}
	indexes[0], indexes[w.index100] = indexes[w.index100], indexes[0]
	return indexes
}
