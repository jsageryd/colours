package main

import (
	"cmp"
	"flag"
	"fmt"
	"math"
	"os"
	"slices"
)

func main() {
	distance := flag.Int("dist", 0, "sort colours by distance from specified colour (e.g., -dist=21)")
	greyscale := flag.Bool("grey", false, "sort colours by greyscale value")
	hue := flag.Bool("hue", false, "sort colours by hue")
	luminance := flag.Bool("lum", false, "sort colours by brightness")
	saturation := flag.Bool("sat", false, "sort colours by vibrancy")
	similarity := flag.Bool("sim", false, "sort colours by visual similarity")
	temperature := flag.Bool("temp", false, "sort colours by warm/cool")

	flag.Parse()

	// Validate distance reference colour if specified
	if *distance != 0 && (*distance < 16 || *distance > 231) {
		fmt.Printf("Error: -dist requires a colour number between 16-231 (got %d)\n", *distance)
		os.Exit(1)
	}

	var (
		standard = makeRange(0, 7)
		high     = makeRange(8, 15)
		colour   = makeRange(16, 231)
		grey     = makeRange(232, 255)
	)

	switch {
	case *distance != 0:
		// Distance-based sorting (closest to specified colour first)
		// Only sort the main colour cube (16-231) by distance
		slices.SortFunc(colour, func(a, b int) int {
			aDist := colourDistance(a, *distance)
			bDist := colourDistance(b, *distance)

			// Secondary sort by brightness for visual appeal
			_, _, aV := hsv(a)
			_, _, bV := hsv(b)

			// Tertiary sort by saturation
			_, aS, _ := hsv(a)
			_, bS, _ := hsv(b)

			return cmp.Or(
				cmp.Compare(aDist, bDist), // Closer to specified colour first
				cmp.Compare(-aV, -bV),     // Brighter first
				cmp.Compare(-aS, -bS),     // More saturated first
			)
		})
	case *greyscale:
		// Greyscale-based sorting (lightest to darkest)
		slices.SortFunc(colour, func(a, b int) int {
			aGrey := colourGreyscale(a)
			bGrey := colourGreyscale(b)

			// Secondary sort by original hue for better visual flow
			aH, _, _ := hsv(a)
			bH, _, _ := hsv(b)

			// Tertiary sort by saturation
			_, aS, _ := hsv(a)
			_, bS, _ := hsv(b)

			return cmp.Or(
				cmp.Compare(-aGrey, -bGrey), // Lighter first
				cmp.Compare(aH, bH),
				cmp.Compare(-aS, -bS),
			)
		})
	case *hue:
		// Hue-based sorting (rainbow order)
		slices.SortFunc(colour, func(a, b int) int {
			aH, aS, aV := hsv(a)
			bH, bS, bV := hsv(b)

			// Separate greys from coloured pixels entirely
			aIsGrey := aS < 0.1
			bIsGrey := bS < 0.1

			// If one is grey and the other isn't, sort non-grey first
			if aIsGrey && !bIsGrey {
				return 1 // a goes after b
			}
			if !aIsGrey && bIsGrey {
				return -1 // a goes before b
			}

			// If both are grey or both are coloured, continue with normal sorting
			return cmp.Or(
				cmp.Compare(aH, bH),
				cmp.Compare(-aS, -bS), // More saturated colours first within each hue
				cmp.Compare(-aV, -bV), // Brighter colours first within saturation levels
			)
		})
	case *luminance:
		// Luminance-based sorting (darkest to lightest)
		slices.SortFunc(colour, func(a, b int) int {
			aLum := colourLuminance(a)
			bLum := colourLuminance(b)

			// Secondary sort by hue for better grouping within similar brightness
			aH, _, _ := hsv(a)
			bH, _, _ := hsv(b)

			// Tertiary sort by saturation for final ordering
			_, aS, _ := hsv(a)
			_, bS, _ := hsv(b)

			return cmp.Or(
				cmp.Compare(aLum, bLum),
				cmp.Compare(aH, bH),
				cmp.Compare(-aS, -bS),
			)
		})
	case *saturation:
		// Saturation-based sorting (muted to vivid)
		slices.SortFunc(colour, func(a, b int) int {
			_, aS, aV := hsv(a)
			_, bS, bV := hsv(b)

			// Secondary sort by brightness for better gradients within saturation levels
			// Tertiary sort by hue for consistent ordering
			aH, _, _ := hsv(a)
			bH, _, _ := hsv(b)

			return cmp.Or(
				cmp.Compare(aS, bS),
				cmp.Compare(-aV, -bV), // Brighter colours first within same saturation
				cmp.Compare(aH, bH),
			)
		})
	case *similarity:
		// Similarity-based sorting (group similar colours)
		slices.SortFunc(colour, func(a, b int) int {
			aGroup := colourSimilarityGroup(a)
			bGroup := colourSimilarityGroup(b)

			// Secondary sort by brightness within groups
			_, _, aV := hsv(a)
			_, _, bV := hsv(b)

			// Tertiary sort by saturation
			_, aS, _ := hsv(a)
			_, bS, _ := hsv(b)

			return cmp.Or(
				cmp.Compare(aGroup, bGroup), // Similarity group
				cmp.Compare(-aV, -bV),       // Brighter first within group
				cmp.Compare(-aS, -bS),       // More saturated first
			)
		})
	case *temperature:
		// Temperature-based sorting (warm to cool)
		slices.SortFunc(colour, func(a, b int) int {
			aTemp := colourTemperature(a)
			bTemp := colourTemperature(b)

			// Secondary sort by hue within temperature groups for natural progression
			aH, aS, aV := hsv(a)
			bH, bS, bV := hsv(b)

			// Tertiary sort by brightness, then saturation for smooth gradients
			return cmp.Or(
				cmp.Compare(aTemp, bTemp),
				cmp.Compare(aH, bH),
				cmp.Compare(-aV, -bV), // Brighter colours first
				cmp.Compare(-aS, -bS), // More saturated first
			)
		})
	default:
		// Original RGB-based sorting
		slices.SortFunc(colour, func(a, b int) int {
			aR, aG, aB := rgb(a)
			bR, bG, bB := rgb(b)

			return cmp.Or(
				cmp.Compare(aR, bR),
				cmp.Compare(aB, bB),
				cmp.Compare(aG, bG),
			)
		})
	}

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

func colourDistance(c1, c2 int) float64 {
	r1, g1, b1 := rgb(c1)
	r2, g2, b2 := rgb(c2)

	// Euclidean distance in RGB space
	dr := float64(r1 - r2)
	dg := float64(g1 - g2)
	db := float64(b1 - b2)

	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func colourGreyscale(c int) float64 {
	r, g, b := rgb(c)

	// Normalize RGB values to 0-1 range
	rNorm, gNorm, bNorm := float64(r)/5.0, float64(g)/5.0, float64(b)/5.0

	// Alternative greyscale conversion (simple average)
	return (rNorm + gNorm + bNorm) / 3.0
}

func hsv(c int) (h, s, v float64) {
	r, g, b := rgb(c)

	// Normalize RGB values to 0-1 range
	rNorm, gNorm, bNorm := float64(r)/5.0, float64(g)/5.0, float64(b)/5.0
	max := math.Max(math.Max(rNorm, gNorm), bNorm)
	min := math.Min(math.Min(rNorm, gNorm), bNorm)
	delta := max - min

	// Value
	v = max

	// Saturation
	if max == 0 {
		s = 0
	} else {
		s = delta / max
	}

	// Hue
	if delta == 0 {
		h = 0 // Undefined, but we'll use 0
	} else if max == rNorm {
		h = 60 * (math.Mod((gNorm-bNorm)/delta, 6))
	} else if max == gNorm {
		h = 60 * ((bNorm-rNorm)/delta + 2)
	} else {
		h = 60 * ((rNorm-gNorm)/delta + 4)
	}
	if h < 0 {
		h += 360
	}

	return h, s, v
}

func colourLuminance(c int) float64 {
	r, g, b := rgb(c)

	// Normalize RGB values to 0-1 range
	rNorm, gNorm, bNorm := float64(r)/5.0, float64(g)/5.0, float64(b)/5.0

	// Use standard luminance formula for perceptual brightness
	return 0.299*rNorm + 0.587*gNorm + 0.114*bNorm
}

func colourSimilarityGroup(c int) int {
	h, s, v := hsv(c)

	// Group colours by visual similarity using HSV clustering
	// Create 12 groups based on hue ranges and saturation/value

	if s < 0.2 {
		// Low saturation - group by value (brightness)
		if v < 0.3 {
			return 0 // Dark greys
		} else if v < 0.7 {
			return 1 // Medium greys
		} else {
			return 2 // Light greys
		}
	}

	// High saturation - group by hue ranges
	hueGroup := int(h / 30) // 12 groups of 30° each
	return 3 + hueGroup     // Groups 3-14
}

func colourTemperature(c int) float64 {
	h, _, _ := hsv(c)

	// Map hue to temperature: 0-120 = warm, 120-300 = cool, 300-360 = warm
	if h <= 60 || h >= 300 {
		return 0 // Warm (reds, oranges, magentas)
	} else if h <= 180 {
		return 2 // Cool (greens, cyans)
	} else {
		return 1 // Medium (blues, purples)
	}
}

func makeRange(from, to int) []int {
	s := make([]int, to-from+1)
	for i := range s {
		s[i] = from + i
	}
	return s
}

func rgb(c int) (r, g, b int) {
	return (c - 16) / 36, ((c - 16) % 36) / 6, (c - 16) % 6
}

func printColour(c int, fg bool) {
	if fg {
		fmt.Printf("\x1b[38;5;%[1]dm  |%03[1]d|\x1b[0m", c)
	} else {
		fmt.Printf("\x1b[48;5;%[1]dm  %03[1]d  \x1b[0m", c)
	}
}
