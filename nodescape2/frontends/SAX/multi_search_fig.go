package main

import (
	"os"
	"fmt"
	"nodescape/nsutil"
	"nodescape/saxutil"
	"strings"
	"runtime"
	"os/exec"
	"bitbucket.org/binet/go-gnuplot/pkg/gnuplot"
)

func pre_name(host, property string) (name string) {
	name = host+property
	name = strings.Replace(name, " ", "", -1)
	name = strings.Replace(name, "\t", "", -1)
	return
}

/* These offsets work and I can actually explain them */
func lead_lag_offset(end int, config nsutil.Config_t) (pos int) {
	offset := (config.Fwin_len * config.Symbol_size)/2.0
	return end - config.Lead_length*config.Symbol_size + int(offset)
}

func profile_offset(end int, config nsutil.Config_t) (pos int) {
	offset := float32(config.Symbol_size * (config.Lead_length + config.Fwin_len)-1)/2.0
	return end-config.Lead_length*config.Symbol_size+int(offset)
}

type proc_res struct {
	dist map[int] float32
	dist_mul float64
	pos int
	key int
}

func proc_window(lag_subct, lead_subct map[string] float32, lead_end int,
					config nsutil.Config_t, subword_lists map[int] []string,
					ch chan proc_res, get_offset func(int, nsutil.Config_t) int) {

	var res proc_res

	res.key = lead_end - (config.Lead_length * config.Symbol_size)
	res.pos = get_offset(lead_end, config)
	res.dist = make(map[int] float32)
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

	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr,
		"Usage: %s <config file> (-f <filename> | -d <host> <property>)\n",
		os.Args[0])
		os.Exit(0)
	}

	config, err := nsutil.Read_config(os.Args[1])

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading configuration file.\n")
	}

	var values []float64
	var times []int
	var host string
	var property string

	switch (os.Args[2]) {
		case "-d":
			if len(os.Args) > 4 {
				host = os.Args[3]
				property = os.Args[4]
				rows, res, err := nsutil.Get_data(os.Args[3], os.Args[4], config)
				if err != nil {
					fmt.Fprintf(os.Stderr,
						"Could not get data from database for %s:%s\n",
						os.Args[3], os.Args[4])
					os.Exit(0)
				}

				values, times = saxutil.Arrays_from_rows(rows, res)

			} else {
				fmt.Fprintf(os.Stderr,
					"Option -d requires a host and a property.\n")
				os.Exit(0)
			}

		case "-f":
			if len(os.Args) > 3 {
				host = os.Args[3]
				values, times, _ = saxutil.Arrays_from_file(os.Args[3])
			} else {
				fmt.Fprintf(os.Stderr, "Option -f requires a file name.\n")
				os.Exit(0)
			}

		default:
			fmt.Fprintf(os.Stderr, "Unrecognized option %s\n", os.Args[2])
			os.Exit(0)
	} // switch(os.Args)


//	words := saxutil.Gen_words(saxutil.Normalize(values), config)
	words := saxutil.Gen_words(values, config)

	fmt.Fprintf(os.Stderr, "Words: %d\n", len(words))

	subword_lists := saxutil.Get_subword_lists(config)

	var profiles map[string] map[string] float32
	profiles = make(map[string] map[string] float32)

	for _, profile := range config.Profiles {
		profiles[profile], _ = saxutil.Read_profile(profile)
	}

	base_subct, _ := saxutil.Read_profile(config.Baseline)

	lag_start := 0
	lead_start := lag_start + config.Lag_length * config.Symbol_size
	lead_end := lead_start + config.Lead_length * config.Symbol_size

	var lag_subct []map[string] float32
	var lead_subct []map[string] float32

	lag_subct = append(lag_subct, saxutil.Count_subwords(
									words[lag_start:lead_start], config))

	lead_subct = append(lead_subct, saxutil.Count_subwords(
									words[lead_start:lead_end], config))

	llch := make(chan proc_res)
	basech := make(chan proc_res)

	profch := make(map[string] chan proc_res)
	for profile := range profiles {
		profch[profile] = make(chan proc_res)
	}

	NCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(NCPU)

	for lag_start = 0; lead_end <= len(words); {

		/* process lead/lag */
		go proc_window(lag_subct[lag_start], lead_subct[lag_start],
					lead_end, config, subword_lists, llch, lead_lag_offset)

		/* search for each profile */
		for name, profile := range profiles {
			go proc_window(profile, lead_subct[lag_start], lead_end, config,
					subword_lists, profch[name], profile_offset)
		}

		/* compute distance for baseline */
		go proc_window(base_subct, lead_subct[lag_start], lead_end,
						config, subword_lists, basech, profile_offset)

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

	base_dist := make(map[int] map[int] float32)
	base_dist_mul := make(map[int] float64)
	base_dist_key := make(map[int] int)
	base_max := 0.0

	ll_dist := make(map[int] map[int] float32)
	ll_dist_mul := make(map[int] float64)
	ll_max := 0.0

	prof_dist := make(map[string] map[int] map[int] float32)
	prof_dist_mul := make(map[string] map[int] float64)
	prof_max := make(map[string] float64)

	for profile := range profiles {
		prof_dist[profile] = make(map[int] map[int] float32)
		prof_dist_mul[profile] = make(map[int] float64)
		prof_max[profile] = 0.0
	}

	for i := 0; i < lag_start; i++ {
		baseres := <-basech
		llres := <-llch

		base_dist[baseres.pos - 1] = baseres.dist
		base_dist_mul[baseres.pos - 1] = baseres.dist_mul
		base_dist_key[baseres.pos - 1] = baseres.key

		ll_dist[llres.pos - 1] = llres.dist
		ll_dist_mul[llres.pos - 1] = llres.dist_mul

		if ll_max < llres.dist_mul {
			ll_max = llres.dist_mul
		}
		if base_max < baseres.dist_mul {
			base_max = baseres.dist_mul
		}

		for i := range profiles {
			profres := <-profch[i]
			prof_dist[i][profres.pos - 1] = profres.dist
			prof_dist_mul[i][profres.pos - 1] = profres.dist_mul
			if prof_max[i] < profres.dist_mul {
				prof_max[i] = profres.dist_mul
			}
		}
	} // for i < lag_start

	for name := range profiles {
		for i := range prof_dist_mul[name] {
			prof_dist_mul[name][i] = -1.0 *
								(prof_dist_mul[name][i] - prof_max[name])
		}
	}
	var dump_arr [][]float64

	dump_arr = append(dump_arr, values)

/*
	for _, arr := range trans_many(ll_dist, len(values), config) {
		dump_arr = append(dump_arr, arr)
	}

	for _, arr := range trans_many(base_dist, len(values), config) {
		dump_arr = append(dump_arr, arr)
	}

	for profile := range profiles {
		for _, arr := range trans_many(prof_dist[profile], len(values), config) {
		dump_arr = append(dump_arr, arr)
		}
	}
*/

	dump_arr = append(dump_arr, trans_one(ll_dist_mul, len(values)))
	dump_arr = append(dump_arr, trans_one(base_dist_mul, len(values)))

	for _, profile := range config.Profiles {
		dump_arr = append(dump_arr, trans_one(prof_dist_mul[profile], len(values)))
	}

	fname := fmt.Sprintf("%s", pre_name(host, property))
	nsutil.Full_dump(dump_arr, times, fname+".dat", words, base_dist_key, config)

	num_cols := 3 + len(profiles)
	fmt.Printf("Number of columns: %d\n", num_cols)


	plotter, err := gnuplot.NewPlotter("", true, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting plotter.\n")
		fmt.Fprintf(os.Stderr, "Exiting...\n")
		os.Exit(0)
	}

	plotter.CheckedCmd("set terminal png small size 600,145")
	plotter.CheckedCmd("set size 1,1")

/*	
	plotter.CheckedCmd("set xdata time")
	plotter.CheckedCmd("set timefmt \"%%s\"")
	plotter.CheckedCmd("set format x \"%%b %%d %%R\"")
*/

//	err = plotter.SetLabels("Time", "Sample value")
	err = plotter.SetLabels("Sample number", "Sample value")
	plotter.CheckedCmd("set output \"values.png\"")
	plotter.CheckedCmd("plot \"%s.dat\" u %d:1 t 'values' w lines",
						fname, num_cols + 1)

	err = plotter.SetLabels("Sample number", "Distance Score")
	plotter.CheckedCmd("set output \"ll.png\"")
	plotter.CheckedCmd("plot \"%s.dat\" u %d:2 t 'word-pair analysis' w lines",
			fname, num_cols + 1)

	err = plotter.SetLabels("Sample number", "Distance Score")
	plotter.CheckedCmd("set output \"baseline.png\"")
	plotter.CheckedCmd("plot \"%s.dat\" u %d:3 t 'baseline analysis' w lines",
			fname, num_cols + 1)

	for i, profile := range config.Profiles {
		err = plotter.SetLabels("Sample number", "Distance Score")
		plotter.CheckedCmd("set output \"%s.png\"", profile)
		plotter.CheckedCmd(
				"plot \"%s.dat\" u %d:%d t '%s profile analysis' w lines",
					fname, num_cols + 1, i+4, profile)
	}
	plotter.Close()

	path, err := exec.LookPath("convert")
	args := []string{"convert"}
	args = append(args, "values.png")
	args = append(args, "ll.png")
	args = append(args, "baseline.png")

	for _, profile := range config.Profiles {
		args = append(args, profile+".png")
	}

	args = append(args, "-append")
	args = append(args, fname+".png")

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Args:\n")
		for _, arg := range args {
			fmt.Fprintf(os.Stderr, "\t%s\n", arg)
		}
	}

	var sprocAttr os.ProcAttr
	sprocAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}

	spath , _ := exec.LookPath("sync")
	sprocess,_ := os.StartProcess(spath, []string{"sync",}, &sprocAttr)
	_, err = sprocess.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Wait failed: %v\n", err)
	}

	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}

	process, err := os.StartProcess(path, args, &procAttr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Start process failed %v\n", err)
	}
	_, err = process.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Wait failed: %v\n", err)
	}

} // main

func trans_many(dist map[int] map[int] float32, length int,
					config nsutil.Config_t) (dump_arr [][]float64) {

	for _, sub_len := range config.Subword_lengths {
		var next_column []float64
		for i := 0; i < length; i++ {
			if dist[i] != nil {
				next_column = append(next_column, float64(dist[i][sub_len]))
			} else {
				next_column = append(next_column, 0.0)
			}
		} // for i
		dump_arr = append(dump_arr, next_column)
	}
	return
}

//func trans_one(dist map[int] float64, max float64, length int) []float64 {
func trans_one(dist map[int] float64, length int) []float64 {
	var next_column []float64
	for i := 0; i < length; i++ {
		//dist[i] /= max
		next_column = append(next_column, dist[i])
	}
	return next_column
}
