package main

import (
	"nodescape/nsutil"
	"fmt"
	"nodescape/saxutil"
	"os"
	"time"
)

func now() float64 {
    return float64(float64(time.Now().UnixNano()) / float64(1e9))
}

func report(task string, start float64, end float64) {
    fmt.Fprintf(os.Stderr, "%s: %0.6f\n", task, end-start)
}

func main() {

	var start float64

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr,
			"Usage: %s <config file> <filename>\n", os.Args[0])
		os.Exit(0)
	}

	config, err := nsutil.Read_config(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read configuration.\n")
		os.Exit(1)
	}

	var values_arr []float64
	filename := os.Args[2]
	values_arr, _, err = saxutil.Arrays_from_file(filename)

	fmt.Fprintf(os.Stderr, "%d values\n", len(values_arr))


	start = now()
	var words []string
	words = saxutil.Gen_words(values_arr, config)
	report("Gen_words", start, now())

	fmt.Fprintf(os.Stderr, "%d words\n", len(words))

	subword_lists := make(map[int] []string)

	for _, sub_len := range config.Subword_lengths {
		subword_lists[sub_len] = saxutil.Gen_subwords(sub_len, config)
	}

	var subct map[string] float32

	subct = saxutil.Count_subwords(words, config)

	for _, sub_len := range config.Subword_lengths {
		for _, subword := range subword_lists[sub_len] {
			fmt.Printf("%s\t%.0f\n", subword, subct[subword]);
		} // for range subword_lists[sublen]
	} // for range subword lengths

}
