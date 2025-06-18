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
	harmonious := flag.Int("harm", 0, "show harmonious colours for specified colour (e.g., -harm=21)")
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

	// Validate harmonious reference colour if specified
	if *harmonious != 0 && (*harmonious < 16 || *harmonious > 231) {
		fmt.Printf("Error: -harm requires a colour number from 16 to 231 (got %d)\n", *harmonious)
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
	case *harmonious != 0:
		// Show colour harmonies for specified reference colour
		printColourHarmonies(*harmonious, colour)
		return
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

// findClosestColour finds the colour closest to the target hue.
func findClosestColour(targetHue float64, colours []int, referenceColour int) int {
	var bestColour int = -1
	var bestDiff float64 = 360

	for _, colour := range colours {
		if colour == referenceColour {
			continue
		}

		h, s, v := hsv(colour)

		// Filter out colours that are too dark or too desaturated
		if v < 0.3 || s < 0.3 {
			continue
		}

		// Calculate circular distance between hues
		diff := math.Min(math.Abs(h-targetHue), 360-math.Abs(h-targetHue))
		if diff < bestDiff {
			bestDiff = diff
			bestColour = colour
		}
	}

	return bestColour
}

// generateHarmonyScheme generates evenly spaced colours for a harmony scheme.
func generateHarmonyScheme(referenceColour int, colours []int, numColours int) []int {
	if numColours < 2 {
		return []int{referenceColour}
	}

	refH, _, _ := hsv(referenceColour)
	result := []int{referenceColour}
	angleStep := 360.0 / float64(numColours)

	// Find the closest colour for each position
	for i := 1; i < numColours; i++ {
		targetHue := math.Mod(refH+angleStep*float64(i), 360)
		if closest := findClosestColour(targetHue, colours, referenceColour); closest != -1 {
			result = append(result, closest)
		}
	}

	return result
}

// generateMonochromeSequential generates colours with the same hue but
// different saturation/brightness.
func generateMonochromeSequential(referenceColour int, colours []int, numColours int) []int {
	if referenceColour < 16 || referenceColour > 231 {
		// Outside the 216-colour RGB cube, fallback to reference only
		return []int{referenceColour}
	}

	// Get the hue of the reference colour
	refH, _, _ := hsv(referenceColour)

	var candidates []int

	// Find all colours with similar hue (within ±5 degrees)
	for _, colour := range colours {
		h, _, _ := hsv(colour)

		// Calculate circular distance between hues
		hueDiff := math.Min(math.Abs(h-refH), 360-math.Abs(h-refH))

		// Include colours with very similar hue
		if hueDiff <= 5 {
			candidates = append(candidates, colour)
		}
	}

	// If we don't have enough candidates, expand the hue tolerance
	if len(candidates) < numColours {
		candidates = []int{}
		for _, colour := range colours {
			h, _, _ := hsv(colour)
			hueDiff := math.Min(math.Abs(h-refH), 360-math.Abs(h-refH))
			if hueDiff <= 15 {
				candidates = append(candidates, colour)
			}
		}
	}

	// Sort candidates by brightness (value) to create a proper sequence
	slices.SortFunc(candidates, func(a, b int) int {
		_, _, aV := hsv(a)
		_, _, bV := hsv(b)
		return cmp.Compare(aV, bV) // Darkest to brightest
	})

	// Select up to numColours, ensuring we include the reference colour
	result := []int{}
	referenceIncluded := false

	for _, colour := range candidates {
		if len(result) >= numColours {
			break
		}
		result = append(result, colour)
		if colour == referenceColour {
			referenceIncluded = true
		}
	}

	// If reference wasn't included, replace the middle colour with it
	if !referenceIncluded && len(result) > 0 {
		midIndex := len(result) / 2
		result[midIndex] = referenceColour

		// Re-sort to maintain brightness order
		slices.SortFunc(result, func(a, b int) int {
			_, _, aV := hsv(a)
			_, _, bV := hsv(b)
			return cmp.Compare(aV, bV)
		})
	}

	return result
}

// generateRGBGradient generates a 6-colour gradient by varying one RGB
// component.
func generateRGBGradient(referenceColour int, colours []int, numColours int) []int {
	if referenceColour < 16 || referenceColour > 231 {
		// Outside the 216-colour RGB cube, fallback to reference only
		return []int{referenceColour}
	}

	// Extract RGB components from the reference colour
	refR, refG, refB := rgb(referenceColour)

	// Generate all 6 variations by varying the component with the highest value
	// This creates the most noticeable brightness variation
	maxComponent := math.Max(math.Max(float64(refR), float64(refG)), float64(refB))

	var result []int

	if float64(refR) == maxComponent {
		// Vary red component from 0 to 5
		for r := 0; r < 6; r++ {
			colour := 16 + r*36 + refG*6 + refB
			result = append(result, colour)
		}
	} else if float64(refG) == maxComponent {
		// Vary green component from 0 to 5
		for g := 0; g < 6; g++ {
			colour := 16 + refR*36 + g*6 + refB
			result = append(result, colour)
		}
	} else {
		// Vary blue component from 0 to 5
		for b := 0; b < 6; b++ {
			colour := 16 + refR*36 + refG*6 + b
			result = append(result, colour)
		}
	}

	return result
}

// generateSplitComplementary generates base colour + 2 colours adjacent to its
// complement.
func generateSplitComplementary(referenceColour int, colours []int, numColours int) []int {
	if referenceColour < 16 || referenceColour > 231 {
		return []int{referenceColour}
	}

	refH, _, _ := hsv(referenceColour)
	result := []int{referenceColour}

	// Find complement's adjacent colours (complement ±30°)
	complementHue := math.Mod(refH+180, 360)
	splitHue1 := math.Mod(complementHue-30, 360)
	splitHue2 := math.Mod(complementHue+30, 360)

	// Find closest colours to the split-complement hues
	if split1 := findClosestColour(splitHue1, colours, referenceColour); split1 != -1 {
		result = append(result, split1)
	}
	if split2 := findClosestColour(splitHue2, colours, referenceColour); split2 != -1 {
		result = append(result, split2)
	}

	return result
}


func findHarmoniousColours(referenceColour int, colours []int) map[string][]int {
	harmony := make(map[string][]int)

	// Generate each harmony scheme using the DRY approach
	harmony["Complementary"] = generateHarmonyScheme(referenceColour, colours, 2)
	harmony["Triadic"] = generateHarmonyScheme(referenceColour, colours, 3)
	harmony["Tetradic"] = generateHarmonyScheme(referenceColour, colours, 4)
	harmony["Pentadic"] = generateHarmonyScheme(referenceColour, colours, 5)
	harmony["Hexadic"] = generateHarmonyScheme(referenceColour, colours, 6)
	harmony["Split-complementary"] = generateSplitComplementary(referenceColour, colours, 3)
	harmony["Monochrome sequential"] = generateMonochromeSequential(referenceColour, colours, 6)
	harmony["RGB gradient"] = generateRGBGradient(referenceColour, colours, 6)

	return harmony
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

func printColourHarmonies(referenceColour int, colours []int) {
	// Get harmonious colours for reference
	harmonies := findHarmoniousColours(referenceColour, colours)

	fmt.Printf("Colour harmonies for colour %d:\n\n", referenceColour)

	// Display each harmony type
	harmonyOrder := []string{
		"Complementary",
		"Split-complementary",
		"Triadic",
		"Tetradic",
		"Pentadic",
		"Hexadic",
		"Monochrome sequential",
		"RGB gradient",
	}

	for _, harmonyType := range harmonyOrder {
		colours := harmonies[harmonyType]
		if len(colours) <= 1 {
			continue // Skip if no harmonious colours found
		}

		fmt.Printf("%s (%d colours):\n", harmonyType, len(colours))

		// Print colours in this harmony
		for _, t := range []bool{false, true} {
			for _, c := range colours {
				printColour(c, t)
			}
		}
		fmt.Println()
		fmt.Println()
	}
}

func printColour(c int, fg bool) {
	if fg {
		fmt.Printf("\x1b[38;5;%[1]dm  |%03[1]d|\x1b[0m", c)
	} else {
		fmt.Printf("\x1b[48;5;%[1]dm  %03[1]d  \x1b[0m", c)
	}
}
