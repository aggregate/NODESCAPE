package main

import (
	"fmt"
	"nodescape/saxutil"
	"nodescape/nsutil"
)

func main() {

	var word string
	var words []string

	fmt.Scanf("%s", &word)
	for word != "stop" {
		words = append(words, word)
		fmt.Scanf("%s", &word)
	}

	fmt.Println("Now enter subword lengths")

	var length int
	var lengths []int

	fmt.Scanf("%d", &length)
	for length != 0 {
		lengths = append(lengths, length)
		fmt.Scanf("%d", &length)
	}

	var config nsutil.Config_t

	config.Subword_lengths = lengths
	config.Alphabet = "abcd"

	counts := saxutil.Count_subwords(words, config)

	mono := map[string] float32{"c": 1.0, "cc": 1.0, "ccc": 1.0,
								"cccc": 1.0, "ccccc": 1.0, "cccccc": 1.0,}

	mulsum := float32(1)
	for _, length := range lengths {
		subwords := saxutil.Gen_subwords(length, config)
		bitmap := saxutil.Gen_bitmap(counts, length, subwords)
		distance := saxutil.Bitmap_distance(bitmap, mono, length, config)

		for _, subword := range subwords {
			if count, ok := counts[subword]; ok {
				fmt.Printf("%s:\t%.0f\t%.3f\n", subword, count, bitmap[subword])
			}
		}

		fmt.Printf("Dist:\t\t%f\n", distance)
		mulsum *= distance
		fmt.Println("------------------")
	}
	fmt.Printf("Multiplied sums: %f\n", mulsum)
}
