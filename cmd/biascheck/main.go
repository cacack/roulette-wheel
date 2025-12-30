package main

import (
	"fmt"
	"sort"

	"roulette-wheel/ball"
	"roulette-wheel/wheel"
)

func runBiasTest(numSpins int) {
	// Track results
	slotCounts := make(map[int]int)
	numberCounts := make(map[string]int)

	// Simulate wheel state
	wheelRotation := 0.0
	wheelSpeed := 0.02

	for i := 0; i < numSpins; i++ {
		// Create a fresh ball for each spin
		b := ball.New(400, 300, 200)
		b.StartSpin(wheelRotation)

		// Run physics until settled (no real-time delay)
		// Use ReferenceTPS for consistent simulation behavior
		maxIterations := 10000 // Safety limit
		for iter := 0; iter < maxIterations && !b.IsSettled(); iter++ {
			wheelRotation += wheelSpeed
			b.Update(wheelRotation, wheelSpeed, ball.ReferenceTPS)
		}

		if b.IsSettled() {
			slot := b.GetSettledSlot()
			slotCounts[slot]++
			num := b.GetWinningNumber(wheel.NumberSequence)
			numberCounts[num]++
		}
	}

	// Print results
	fmt.Printf("\n=== Bias Test Results (%d spins) ===\n\n", numSpins)

	// Expected count per slot
	expected := float64(numSpins) / 38.0
	fmt.Printf("Expected hits per number: %.1f\n\n", expected)

	// Sort numbers for display
	type result struct {
		num   string
		count int
	}
	var results []result
	for num, count := range numberCounts {
		results = append(results, result{num, count})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].count > results[j].count
	})

	// Show top 10 and bottom 10
	fmt.Println("TOP 10 (most frequent):")
	for i := 0; i < 10 && i < len(results); i++ {
		r := results[i]
		deviation := (float64(r.count) - expected) / expected * 100
		fmt.Printf("  %2s: %4d hits (%+.1f%%)\n", r.num, r.count, deviation)
	}

	fmt.Println("\nBOTTOM 10 (least frequent):")
	for i := len(results) - 10; i < len(results); i++ {
		if i >= 0 {
			r := results[i]
			deviation := (float64(r.count) - expected) / expected * 100
			fmt.Printf("  %2s: %4d hits (%+.1f%%)\n", r.num, r.count, deviation)
		}
	}

	// Chi-square test
	chiSquare := 0.0
	for _, count := range slotCounts {
		diff := float64(count) - expected
		chiSquare += (diff * diff) / expected
	}
	fmt.Printf("\nChi-square statistic: %.2f\n", chiSquare)
	fmt.Printf("(For 37 df, values > 52.2 indicate bias at p<0.05)\n")

	// Check specifically for 0 and 00
	fmt.Printf("\nGreen zeros:\n")
	fmt.Printf("  0:  %d hits (expected %.1f)\n", numberCounts["0"], expected)
	fmt.Printf("  00: %d hits (expected %.1f)\n", numberCounts["00"], expected)
}

func main() {
	runBiasTest(10000)
}
