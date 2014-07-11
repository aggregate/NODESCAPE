package main

import (
	"os"
	"fmt"
	"nodescape/nsutil"
	"nodescape/saxutil"
	"os/exec"
	"runtime"
	"bitbucket.org/binet/go-gnuplot/pkg/gnuplot"
)

type proc_res struct {
	dist map[int] float32
	dist_mul float64
	pos int
}

func main() {

	if len(os.Args) < 6 {
		fmt.Fprintf(os.Stderr,
		"Usage: %s <config file> -f|p <filename> -d <host> <property>\n",
		os.Args[0])
		os.Exit(0)
	}

	config, err := nsutil.Read_config(os.Args[1])

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error readig config.\n")
	}

	var profiles map[string] map[string] float32
	profiles = make(map[string] map[string] float32)

	var values []float64
	var times []int

	for i := 2; i < len(os.Args); {

		fmt.Fprintf(os.Stderr, "Process: %s\n", os.Args[i])
		switch (os.Args[i]) {
		case "-p":
			if i+1 < len(os.Args) {
				profiles[os.Args[i+1]], _ =  saxutil.Read_profile(os.Args[i+1])
				i += 2
			} else {
				i += 1
			}
		case "-f":
			if i+1 < len(os.Args) {
				values, times, _ = saxutil.Arrays_from_file(os.Args[i+1])
				i += 2
			} else {
				i += 1
			}

		case "-d":
			if i+2 < len(os.Args) {
				rows, res, err := nsutil.Get_data(os.Args[i+1], os.Args[i+2], config)
				if err != nil {
					fmt.Fprintf(os.Stderr,
						"Could not get data from database for %s:%s\n",
						os.Args[i+1], os.Args[i+2])
					os.Exit(0)
				}
				values, times = saxutil.Arrays_from_rows(rows, res)
				i += 3
			} else {
				i += 2
			}
		default:
			fmt.Fprintf(os.Stderr, "Unrecognized option %s\n", os.Args[i])
			i += 1
		}
	} // for i < len(os.Args)

	/* get the list of words for our time series */
	//words := saxutil.Gen_words(saxutil.Normalize(values), config)
	words := saxutil.Gen_words(values, config)

	/* 
		Get lists of the subwords.
		Use them to control iteration over the maps containing the subwords.
	*/
	subword_lists := make(map[int] []string)
	for _, sub_len := range config.Subword_lengths {
		subword_lists[sub_len] = saxutil.Gen_subwords(sub_len, config)
	}

	lead_start := 0
	lead_end := lead_start + config.Lead_length * config.Symbol_size

	/* Get subword counts for the initial window */
	var lead_subct []map[string] float32
	lead_subct = append(lead_subct, saxutil.Count_subwords(
										words[lead_start:lead_end], config))
	/* set number of CPUs */
	NCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(NCPU)


	for name, profile := range profiles {
		ch := make(chan proc_res)

		for lead_end <= len(words) {
			/* compute distance between the subword maps */
			go proc_window(profile, lead_subct[lead_start],
							lead_end, config, subword_lists, ch)

			/* Adjust subword counts for the next iteration */
			if lead_end < len(words) {
				lead_subct = append(lead_subct,
								saxutil.Adj_count(lead_subct[lead_start],
								words[lead_start], words[lead_end], config))
			} // if lead_end < len(words)

			lead_start += 1
			lead_end += 1
		} // for lead_end < len(words)

		dist := make(map[int] map[int] float32)
		dist_mul := make(map[int] float64)

		/* Gather results from the goroutines */
		maxdist_mul := 0.0
		for i := 0; i < lead_start; i++ {
			res := <-ch
			dist[res.pos - 1] = res.dist
			dist_mul[res.pos - 1] = res.dist_mul
			if maxdist_mul < res.dist_mul {
				maxdist_mul = res.dist_mul
			}
		}

		for _, sub_len := range config.Subword_lengths {
			max := float32(0.0)
			for i := 0; i < len(values); i++ {
				if dist[i] != nil {
					if dist[i][sub_len] > max {
						max = dist[i][sub_len]
					}
				}
			}
			for i := 0; i < len(values); i++ {
				if dist[i] != nil {
				dist[i][sub_len] = -1.0 * (dist[i][sub_len] - max)
				}
			}
		} // for _, sub_len := range config.Subword_lengths

		var dump_arr [][]float64
		dump_arr = append(dump_arr, values)

		for _, sub_len := range config.Subword_lengths {

			var next_column []float64
			for i := 0; i <len(values) + 1; i++ {
				if dist[i] != nil {
					next_column = append(next_column, float64(dist[i][sub_len]))
				} else {
					next_column = append(next_column, 0.0)
				}
			} // for i < len(values)

			dump_arr = append(dump_arr, next_column)

		} // for _, sub_len := range config.Subword_lengths

		var next_column []float64
		for i := 0; i < len(values) + 1; i++ {
			dist_mul[i] /= maxdist_mul
			next_column = append(next_column, dist_mul[i])
		}

		dump_arr = append(dump_arr, next_column)

		fname := fmt.Sprintf("%s", name)
		nsutil.Text_dump(dump_arr, times, fname+".dat", words, config)

		plotter, err := gnuplot.NewPlotter("", true, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting plotter.\n")
			fmt.Fprintf(os.Stderr, "Exiting...\n")
			os.Exit(0)
		}

		plotter.CheckedCmd("set terminal png size 1500,200")
		plotter.CheckedCmd("set size 1,1")
		plotter.CheckedCmd(fmt.Sprintf("set output \"%s.png\"", fname))
		//plotter.CheckedCmd("set log y")

		err = plotter.SetLabels("Time", "sample value")
		plotstr := fmt.Sprintf("plot \"%s.dat\" u %d:1 t 'values' w lines",
								fname, len(config.Subword_lengths)+3)
		plotter.CheckedCmd(plotstr)

		plotter.CheckedCmd("set output \"dist_mul.png\"")
		//plotter.CheckedCmd("set log y")

		err = plotter.SetLabels("Time", "multipled distance")
		plotstr = fmt.Sprintf(
			"plot \"%s.dat\" u %d:%d t 'values' w lines", fname,
			len(config.Subword_lengths)+3, len(config.Subword_lengths)+2)
		plotter.CheckedCmd(plotstr)


		path, err := exec.LookPath("convert")
		args := []string{fmt.Sprintf("%s.png", fname)}
		args = append(args, fmt.Sprintf("%s.png", fname))
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
						"plot \"%s.dat\" u %d:%d t 'distance %d' w lines",
						fname, len(config.Subword_lengths)+3, i+2, subsize)
			plotter.CheckedCmd(plotstr)
		}

		plotter.Close()
		/* Generate big png? */

		args = append(args, "-append")
		args = append(args, fmt.Sprintf("%s.png", fname))

		var procAttr os.ProcAttr
		procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}

		process, err := os.StartProcess(path, args, &procAttr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Start process failed: %v\n", err)
		}

		_, err = process.Wait()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Wait failed: %v\n", err)
		}


		lead_start = 0
		lead_end = lead_start + config.Lead_length * config.Symbol_size
	} // for _, profile := range profiles

} // main

func proc_window(lag_subct, lead_subct map[string] float32, lead_end int,
                    config nsutil.Config_t, subword_lists map[int] []string,
                    ch chan proc_res) {

    var res proc_res

    // Get the offset for distance (in symbols)
    offset := float32(config.Fwin_len)/2.0 - float32(config.Lead_length)/2.0

    // Set the position for this distance computation (in samples)
    res.pos = lead_end + int(float32(config.Symbol_size) * offset)

    //res.pos = lead_end + config.Dist_offset
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

