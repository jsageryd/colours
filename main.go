package main

import "fmt"

func main() {
	for _, t := range []bool{false, true} {
		for _, n := range []int{0, 7} {
			printColour(n, t)
		}
	}

	fmt.Println()

	for _, t := range []bool{false, true} {
		for _, n := range []int{8, 15} {
			printColour(n, t)
		}
	}

	fmt.Println()
	fmt.Println()

	for _, t := range []bool{false, true} {
		for n := 1; n <= 6; n++ {
			printColour(n, t)
		}
	}

	fmt.Println()

	for _, t := range []bool{false, true} {
		for n := 9; n <= 14; n++ {
			printColour(n, t)
		}
	}

	fmt.Println()
	fmt.Println()

	for n := 16; n <= 255; n++ {
		printColour(n, false)

		if (n-15)%6 == 0 {
			for m := range 6 {
				printColour(n-5+m, true)
			}
			fmt.Println()
		}

		if (n-15)%36 == 0 {
			fmt.Println()
		}
	}
}

func printColour(n int, fg bool) {
	if fg {
		fmt.Printf("\x1b[38;5;%[1]dm  |%03[1]d|\x1b[0m", n)
	} else {
		fmt.Printf("\x1b[48;5;%[1]dm  %03[1]d  \x1b[0m", n)
	}
}
