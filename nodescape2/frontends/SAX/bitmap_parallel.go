package main

import (
	"nodescape/nsutil"
	"fmt"
	"nodescape/saxutil"
	"os"
	"os/exec"
	"time"
	"runtime"
	"bitbucket.org/binet/go-gnuplot/pkg/gnuplot"
)

func now() float64 {
    return float64(float64(time.Now().UnixNano()) / float64(1e9))
}

func report(task string, start float64, end float64) {
    fmt.Fprintf(os.Stderr, "%s: %0.6f\n", task, end-start)
}

type proc_res struct {
	dist map[int] float32
	dist_mul float64
	pos int
}

func proc_window(lag_subct, lead_subct map[string] float32, lead_end int,
					config nsutil.Config_t, subword_lists map[int] []string,
					ch chan proc_res) {

	var res proc_res

	res.pos = lead_end
	res.dist = make(map[int] float32)
/*
	lag_subct := saxutil.Count_subwords(words[lag_start:lead_start], config)

	lead_subct := saxutil.Count_subwords(words[lead_start:lead_end], config)
*/


	res.dist_mul = float64(1)
	for _, sub_len := range config.Subword_lengths {
		lag_map := saxutil.Gen_bitmap(lag_subct, sub_len,
										subword_lists[sub_len])
		lead_map := saxutil.Gen_bitmap(lead_subct, sub_len,
										subword_lists[sub_len])

		res.dist[sub_len] = saxutil.Bitmap_distance(
									lag_map, lead_map, sub_len, config)
		res.dist_mul *= float64(res.dist[sub_len])
	}

	ch <- res
}

func main() {

	prog_start := now()
	var start float64

	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr,
			"Usage: %s <config file> -f|d <host> <property>\n", os.Args[0])
		os.Exit(0)
	}

	config, err := nsutil.Read_config(os.Args[1])

	var values_arr []float64
	var times []int
	for i := 2; i + 2 < len(os.Args); i += 3 {

		if os.Args[i] == "-f" {

			filename := nsutil.Pre_name(os.Args[i+1], os.Args[i+2])+".txt"
			values_arr, times, err = saxutil.Arrays_from_file(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"Could not get data from file for %s:%s\n",
					os.Args[i+1], os.Args[i+2])
				os.Exit(0)
			}

		} else if os.Args[i] == "-d" {

			start = now()
			rows, res, err := nsutil.Get_data(os.Args[i+1], os.Args[i+2], config)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"Could not get data from database for %s:%s\n",
					os.Args[i+1], os.Args[i+2])
				os.Exit(0)
			}
			values_arr, times = saxutil.Arrays_from_rows(rows, res)
			report("Database Query", start, now())

		} else {

			fmt.Fprintf(os.Stderr, "Unrecognized option %s\n", os.Args[i])
			os.Exit(0)

		}
	} // End argument processing for loop

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

	fmt.Fprintf(os.Stderr, "Last time: %d\n", times[len(times)-1])
	fmt.Fprintf(os.Stderr, "len(values_arr): %d\n", len(values_arr))
	start = now()
	var words []string
	words = saxutil.Gen_words(values_arr, config)
	fmt.Fprintf(os.Stderr, "len(words): %d\n", len(words))
	report("Gen_words", start, now())


	lag_start := 0
	lead_start := lag_start + config.Lag_length * config.Symbol_size
	lead_end := lead_start + config.Lead_length * config.Symbol_size

	var dist map[int] map[int] float32
	dist = make(map[int] map[int] float32)
	dist_mul := make(map[int] float64)

	start = now()

	subword_lists := make(map[int] []string)

	for _, sub_len := range config.Subword_lengths {
		subword_lists[sub_len] = saxutil.Gen_subwords(sub_len, config)
	}

	NCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(NCPU)

	start = now()

	var lead_subct []map[string] float32
	var lag_subct []map[string] float32

    lag_subct = append(lag_subct, saxutil.Count_subwords(
                            words[lag_start:lead_start], config))

    lead_subct = append(lead_subct, saxutil.Count_subwords(
                            words[lead_start:lead_end], config))


	ch := make(chan proc_res)
	for lag_start = 0; lead_end <= len(words); {

		go proc_window(lag_subct[lag_start], lead_subct[lag_start],
						lead_end, config, subword_lists, ch)

        lag_subct = append(lag_subct,
						saxutil.Adj_count(lag_subct[lag_start],
							words[lag_start], words[lead_start], config))
        if lead_end < len(words) {
            lead_subct = append(lead_subct,
							saxutil.Adj_count(lead_subct[lag_start],
								words[lead_start], words[lead_end], config))
        }

		lag_start += 1
		lead_start += 1
		lead_end += 1
	}

	for i := 0; i < lag_start; i++ {
		res := <-ch
		dist[res.pos] = res.dist
		dist_mul[res.pos] = res.dist_mul
	}

	report("proc_window", start, now())

	start = now()
	var dump_arr [][]float64
	/* each row: value, distance, distance, ... , time */

	dump_arr = append(dump_arr, values_arr)

	/* transform the distance maps into arrays of float64 */
	for _, sub_len := range config.Subword_lengths {

		var next_column []float64
		for i := 0; i < len(values_arr) + 1; i++ {
			if dist[i] != nil {
				next_column = append(next_column, float64(dist[i][sub_len]))
			} else {
				next_column = append(next_column, 0.0)
			}
		} // for i

		dump_arr = append(dump_arr, next_column)
	} // for _, sub_len

	var next_column []float64
	for i := 0; i < len(values_arr) + 1; i++ {
		next_column = append(next_column, float64(dist_mul[i]))
	} // for i

	dump_arr = append(dump_arr, next_column)
	report("Transform", start, now())

	start = now()
	nsutil.Text_dump(dump_arr, times,"window.dat", words, config)
	report("Write", start, now())

	start = now()
	plotter, err := gnuplot.NewPlotter("", true, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting plotter.\n")
		fmt.Fprintf(os.Stderr, "Exiting...\n")
		os.Exit(0)
	}


    plotter.CheckedCmd("set terminal png size 1500,200")
    plotter.CheckedCmd("set size 1,1")
    plotter.CheckedCmd("set output \"window_values.png\"")
    //plotter.CheckedCmd("set log y")

	err = plotter.SetLabels("Time", "sample value")
	plotstr := fmt.Sprintf("plot \"window.dat\" u %d:1 t 'values' w lines",
							len(config.Subword_lengths)+3)
    plotter.CheckedCmd(plotstr)

	plotter.CheckedCmd("set output \"dist_mul.png\"")
    //plotter.CheckedCmd("set log y")

	err = plotter.SetLabels("Time", "multipled distance")
	plotstr = fmt.Sprintf("plot \"window.dat\" u %d:%d t 'values' w lines",
							len(config.Subword_lengths)+3, len(config.Subword_lengths)+2)
    plotter.CheckedCmd(plotstr)


	path, err := exec.LookPath("convert")
	args := []string{"window_values.png"}
	args = append(args, "window_values.png")
	args = append(args, "dist_mul.png")

	for i := 0; i < len(config.Subword_lengths); i++ {
		subsize := config.Subword_lengths[i]

		label := fmt.Sprintf("subword size %d", subsize)

		err = plotter.SetLabels("Time", label)
		/* new output image */
		plotstr = fmt.Sprintf("set output \"distances_%d.png\"", subsize)
		plotter.CheckedCmd(plotstr)
		args = append(args, fmt.Sprintf("distances_%d.png", subsize))

		/* plot the image */
		plotstr = fmt.Sprintf(
					"plot \"window.dat\" u %d:%d t 'distance %d' w lines",
					len(config.Subword_lengths)+3, i+2, subsize)
		plotter.CheckedCmd(plotstr)
	}

	plotter.Close()
	/* Generate big png? */
	report("Plot", start, now())

	start = now()
	args = append(args, "-append")
	args = append(args, "all.png")

	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}

	for _, arg := range args {
		fmt.Printf("%s\n", arg)
	}

	process, err := os.StartProcess(path, args, &procAttr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Start process filed: %v\n", err)
	}

	_, err = process.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Wait failed: %v\n", err)
	}
	report("Convert", start, now())
	report("Total", prog_start, now())
}
