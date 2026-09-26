package main

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"roulette-wheel/ball"
	"roulette-wheel/wheel"
)

// runBiasTest simulates numSpins and prints a report; it returns an error if
// any spin failed to settle or the distribution shows bias at p < 0.001.
func runBiasTest(numSpins int) error {
	slotCounts := ball.SimulateSpins(numSpins)

	settled := 0
	for _, count := range slotCounts {
		settled += count
	}
	if settled == 0 {
		return errors.New("no spins settled")
	}

	// Print results
	fmt.Printf("\n=== Bias Test Results (%d of %d spins settled) ===\n\n", settled, numSpins)

	// Expected count per slot
	expected := float64(settled) / ball.NumSlots
	fmt.Printf("Expected hits per number: %.1f\n\n", expected)

	// Sort numbers for display
	type result struct {
		num   string
		count int
	}
	var results []result
	numberCounts := make(map[string]int)
	for slot, count := range slotCounts {
		num := wheel.NumberSequence[slot]
		results = append(results, result{num, count})
		numberCounts[num] = count
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
	chiSquare := ball.ChiSquare(slotCounts)
	fmt.Printf("\nChi-square statistic: %.2f\n", chiSquare)
	fmt.Printf("(For 37 df, values > %.2f indicate bias at p<0.001)\n", ball.ChiSquareCritical001)

	// Check specifically for 0 and 00
	fmt.Printf("\nGreen zeros:\n")
	fmt.Printf("  0:  %d hits (expected %.1f)\n", numberCounts["0"], expected)
	fmt.Printf("  00: %d hits (expected %.1f)\n", numberCounts["00"], expected)

	if settled != numSpins {
		return fmt.Errorf("%d of %d spins failed to settle", numSpins-settled, numSpins)
	}
	if !ball.IsFair(slotCounts) {
		return errors.New("distribution is biased at p<0.001")
	}
	return nil
}

// main exits non-zero when the bias test fails, so it can gate scripts or CI.
func main() {
	if err := runBiasTest(50000); err != nil {
		fmt.Printf("\nFAIL: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\nPASS: no bias detected at p<0.001")
}
