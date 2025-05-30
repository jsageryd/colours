package main

import (
	"fmt"
	"slices"
)

func main() {
	var (
		standard = makeRange(0, 7)
		high     = makeRange(8, 15)
		colour   = makeRange(16, 231)
		grey     = makeRange(232, 255)
	)

	for row := range slices.Chunk([]int{
		standard[0], standard[len(standard)-1],
		high[0], high[len(high)-1],
	}, 2) {
		for _, t := range []bool{false, true} {
			for _, c := range row {
				printColour(c, t)
			}
		}
		fmt.Println()
	}

	for row := range slices.Chunk(slices.Concat(
		standard[1:len(standard)-1], high[1:len(high)-1],
	), 6) {
		fmt.Println()
		for _, t := range []bool{false, true} {
			for _, c := range row {
				printColour(c, t)
			}
		}
	}

	fmt.Println()

	for block := range slices.Chunk(append(colour, grey...), 36) {
		fmt.Println()
		for row := range slices.Chunk(block, 6) {
			for _, t := range []bool{false, true} {
				for _, c := range row {
					printColour(c, t)
				}
			}
			fmt.Println()
		}
	}
}

func makeRange(from, to int) []int {
	s := make([]int, to-from+1)
	for i := range s {
		s[i] = from + i
	}
	return s
}

func printColour(c int, fg bool) {
	if fg {
		fmt.Printf("\x1b[38;5;%[1]dm  |%03[1]d|\x1b[0m", c)
	} else {
		fmt.Printf("\x1b[48;5;%[1]dm  %03[1]d  \x1b[0m", c)
	}
}
