package cpu

import "testing"

func BenchmarkNOP(b *testing.B) {
	var (
		state State
		cycle Cycle
	)

	for i := 0; i < b.N; i++ {
		cycle = FetchCycle(state)
		cycle = StartCycle(&state, cycle)
		FinishCycle(&state, cycle, 0)
	}
}
