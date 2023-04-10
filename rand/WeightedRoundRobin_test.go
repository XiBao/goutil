package rand

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name  string
		pdf   []int
		valid bool
	}{
		{"happy distribution", []int{30, 40, 20, 10}, true},
		{"90%", []int{50, 40}, false},
		{"No elements ", []int{}, false},
		{"out of range value 1", []int{101, 2}, false},
		{"out of range value 2", []int{30, 40, 20, 10, -10, 10}, false},
		{"zero", []int{0}, false},
		{"hundred", []int{0, 0, 100}, true},
		{"hundred", []int{100, 0}, true},
		{"hundred", []int{100}, true},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s: %v", test.name, test.pdf), func(t *testing.T) {
			_, err := NewWeightedRoundRobin(test.pdf)
			if !test.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPick(t *testing.T) {
	const n = 1000
	tests := []struct {
		name              string
		pdf               []int
		allowedMaxDiffPct int
	}{
		{"happy distribution", []int{30, 40, 20, 10}, 5},
		{"50/50 distribution ", []int{50, 50}, 5},
		{"one element ", []int{100}, 0},
		{"twenty elements", []int{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, 3},
		{"strongly unbalanced", []int{90, 2, 2, 2, 2, 2}, 3},
		{"strongly unbalanced2", []int{1, 99}, 1},
		{"one zero", []int{100, 0}, 0},
		{"multiple zeros", []int{100, 0, 0}, 0},
		{"multiple zeros", []int{0, 100, 0}, 0},
		{"multiple zeros", []int{0, 0, 100}, 0},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s: %v", test.name, test.pdf), func(t *testing.T) {
			wrr, err := NewWeightedRoundRobin(test.pdf)
			require.NoError(t, err)
			result := make([]int, len(test.pdf))
			for i := 0; i < n; i++ {
				idx := wrr.Pick()
				assert.True(t, idx >= 0 && idx < len(test.pdf), "Pick returned index out of range")
				result[idx]++
			}
			sum := sum(result)
			assert.Equal(t, sum, n)
			b := checkAllowedDiff(test.pdf, result, sum, test.allowedMaxDiffPct)
			assert.True(t, b, "crossed maximum allowed diff", result)
		})
	}
}

func TestPickVector(t *testing.T) {
	const n = 1000
	tests := []struct {
		name              string
		pdf               []int
		allowedMaxDiffPct int
	}{
		{"happy distribution", []int{30, 40, 20, 10}, 5},
		{"50/50 distribution ", []int{50, 50}, 5},
		{"one element ", []int{100}, 0},
		{"twenty elements", []int{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, 3},
		{"strongly unbalanced", []int{90, 2, 2, 2, 2, 2}, 2},
		{"one zero", []int{100, 0}, 0},
		{"multiple zeros", []int{100, 0, 0}, 0},
		{"multiple zeros", []int{0, 100, 0}, 0},
		{"multiple zeros", []int{0, 0, 100}, 0},
		{"multiple zeros", []int{0, 0, 0, 100, 0, 0}, 0},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s: %v", test.name, test.pdf), func(t *testing.T) {
			wrr, err := NewWeightedRoundRobin(test.pdf)
			require.NoError(t, err)

			result := map[int][]int{}
			for i := 0; i < len(test.pdf); i++ {
				result[i] = make([]int, len(test.pdf))
			}

			for i := 0; i < n; i++ {

				indexes := wrr.PickVector()
				for _, v := range indexes {
					assert.True(t, v >= 0 && v < len(test.pdf), "Pick returned index out of range")
				}

				for i := 0; i < len(test.pdf); i++ {
					result[i][indexes[i]]++
				}
			}

			for i := 0; i < len(test.pdf); i++ {
				verticalSum := 0
				horizontalSum := sum(result[i])
				for _, v := range result {
					verticalSum += v[i]
				}
				assert.Equal(t, horizontalSum, n)
				assert.Equal(t, verticalSum, n)
			}
		})
	}
}

func sum(result []int) (sum int) {
	for _, v := range result {
		sum += v
	}
	return sum
}

// checks if result is in allowed diff ±diffPercent%
func checkAllowedDiff(pdf, result []int, sum int, diffPercent int) bool {
	for i, p := range pdf {
		var pct = float64(p) / 100
		var diff = (float64(diffPercent) / 100) * float64(sum)
		var base = pct * float64(sum)

		if float64(result[i]) > base+diff || float64(result[i]) < base-diff {
			return false
		}
	}
	return true
}
