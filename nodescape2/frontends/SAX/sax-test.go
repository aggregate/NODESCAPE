package main

import (
	"nodescape/nsutil"
	"fmt"
	"nodescape/saxutil"
	"os"
	"bitbucket.org/binet/go-gnuplot/pkg/gnuplot"
)

func main() {

	if len(os.Args) < 8 {
		fmt.Fprintf(os.Stderr,
			"Usage: %s <config file> -f|d <host> <property> "+
			"-f|d <host> <property>\n", os.Args[0])
		os.Exit(0)
	}

	config, err := nsutil.Read_config(os.Args[1])

	var values_arr [][]float64
	values_arr = append(values_arr, []float64{0.0})
	values_arr = append(values_arr, []float64{0.0})
	var times []int
	for i := 2; i + 2 < len(os.Args); i += 3 {

		if os.Args[i] == "-f" {

			filename := nsutil.Pre_name(os.Args[i+1], os.Args[i+2])+".txt"
			values_arr[(i-2)/3], times, err = saxutil.Arrays_from_file(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"Could not get data from file for %s:%s\n",
					os.Args[i+1], os.Args[i+2])
				os.Exit(0)
			}

		} else if os.Args[i] == "-d" {

			rows, res, err := nsutil.Get_data(os.Args[i+1], os.Args[i+2], config)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"Could not get data from database for %s:%s\n",
					os.Args[i+1], os.Args[i+2])
				os.Exit(0)
			}
			values_arr[(i-2)/3], times = saxutil.Arrays_from_rows(rows, res)

		} else {

			fmt.Fprintf(os.Stderr, "Unrecognized option %s\n", os.Args[i])
			os.Exit(0)

		}
	} // End argument processing for loop

	min := len(times)
	if min > len(values_arr[0]) { min = len(values_arr[0])}
	if min > len(values_arr[1]) { min = len(values_arr[1])}
	times = times[0:min]
	values_arr[0] = values_arr[0][0:min]
	values_arr[1] = values_arr[1][0:min]


	/*
		Now, what I'd like to do is take two different series 
			(for the same parameter), and get bitmaps of both. I also want
			to compare the bitmaps to get distance. I might also look at
			distance between SAX words. I want to find a way to plot the
			series one above the other.

		I need primitives for:
			X getting SAX words	
			X getting lists of SAX words 
			X getting subword counts
			X getting bitmaps from lists of SAX words
			X getting distance between two bitmaps
			getting distance between two SAX words

		I'll also need image manipulation primitives:
			bitmap to rbg
			write ppm from rbg
	*/

	words := saxutil.Gen_words(values_arr[0], config)
	words_old := saxutil.Gen_words(values_arr[1], config)

	subword_ct := saxutil.Count_subwords(words, config)
	subword_ct_old := saxutil.Count_subwords(words_old, config)

	bitmaps := make(map[int] map[string] float32)
	bitmaps_old := make(map[int] map[string] float32)

	for _, sub_len := range config.Subword_lengths {
		bitmaps[sub_len] = saxutil.Gen_bitmap(subword_ct, sub_len, config)
		bitmaps_old[sub_len] = saxutil.Gen_bitmap(subword_ct_old, sub_len, config)
	}

	for _, sub_len := range config.Subword_lengths {
		fmt.Printf("Distance between bitmaps at level %d: %f\n", sub_len,
			saxutil.Bitmap_distance(bitmaps[sub_len], bitmaps_old[sub_len],
				sub_len, config))

		image_name := fmt.Sprintf("%s_%d.ppm",
				nsutil.Pre_name(os.Args[3], os.Args[4]), sub_len)
		image := saxutil.PPM_from_bitmap(bitmaps[sub_len], sub_len, config)
		saxutil.Write_PPM(image, image_name)

		image_name = fmt.Sprintf("%s_%d.ppm",
				nsutil.Pre_name(os.Args[6], os.Args[7])+"_old", sub_len)
		image = saxutil.PPM_from_bitmap(bitmaps_old[sub_len], sub_len, config)
		saxutil.Write_PPM(image, image_name)
	}

	plotter, err := gnuplot.NewPlotter("", true, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting plotter.\n")
		fmt.Fprintf(os.Stderr, "Exiting...\n")
		os.Exit(0)
	}
	defer plotter.Close()

	nsutil.Text_dump(values_arr, times, "sax-test.dat")

	err = plotter.SetLabels("Time", "sample value and distance")

    plotter.CheckedCmd("set terminal png size 1500,500")
    plotter.CheckedCmd("set size 1,1")
    plotter.CheckedCmd("set output \"sax-test.png\"")
    plotter.CheckedCmd("set log y")
    plotter.CheckedCmd("plot \"sax-test.dat\" u 3:1 t 'series 1' w lines, "+
                        "\"sax-test.dat\" u 3:2 t 'series 2' w lines")

}
