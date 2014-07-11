package main

import (
	"os"
	"fmt"
	"math"
	"time"
	"bitbucket.org/binet/go-gnuplot/pkg/gnuplot"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"github.com/GaryBoone/GoStats"
	/* 
		This package is named incorrectly. It should be named
		"github.com/GaryBoone/stats". Unfortunatedly, the author didn't 
		feel it was necessary to follow the Go package naming conventions.
	*/
)

type image struct {
	r int
	g int
	b int
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <host> <property>\n", os.Args[0])
		os.Exit(0)
	}

	db := mysql.New("tcp", "", "super.ece.engr.uky.edu:13092", "nsfront",
					"frontpass", "nodescape")

	err := db.Connect()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, res, err := db.Query("select data, ctime from ukanstats where "+
						"host = \""+os.Args[1]+"\" and label = \""+
						os.Args[2]+"\" order by ctime asc;")

	if err != nil {
		panic(err)
	}

	var nums []float64
	var times []int
	var toff = -1

	/* Now, pull the data out of the results structure */
	for _, row := range rows {
		val := row.Float(res.Map("data"))
		loc, _ := time.LoadLocation("Local")
		ts := row.Time(res.Map("ctime"), loc).Unix()
		//fmt.Printf("%f\t%d\n", val, ts)
		if toff == -1 {
			toff = int(ts)
		}
		nums = append(nums, val)
		times = append(times, int(ts)-toff)
	}

	fmt.Printf("Got %d samples\n", len(nums))

	/* Get words for the new data */
	var sec_len = 6 // number of samples per symbol
	var word_len = 6*sec_len // width of feature window (word length in samples)
	var win_len = 4

	fmt.Printf("Window length is %d samples.\n", win_len*word_len)

	words := get_sax_words(nums, sec_len, word_len)

	fmt.Printf("Generated %d words\n", len(words))

	/* move lead and lag windows across the set of words */
	lag_ct := make(map[string] float32)
	lead_ct := make(map[string] float32)
	var distances []float32

	/* initialize the first part of distances to shift the graph over
		and be more indicative of where the error is actually high */

	for i := 0; i < word_len*win_len; i++ {
		distances = append(distances, float32(0.0))
	}

	l2_subwords := []string{ "aa", "ab", "ac", "ad",
					"ba", "bb", "bc", "bd",
					"ca", "cb", "cc", "cd",
					"da", "db", "dc", "dd",}

	/* count level 2 subword occurrences */
	level := 2
	max := 0
	maxval := float32(0.0)
	for i := 0; i < len(words)-(2*win_len*word_len); i++ {
		/* count subwords in lag window */
		/* Lag window is [i ..i+3*word_len] */

		for j := i; j < i+win_len*word_len; j++ {
			for k:= 0; k < len(words[j])-level+1; k++ {
				lag_ct[words[j][k:k+level]] += 1
			}
		}

		/* count subwords in lead window */
		/* Lead window is [i+3*word_len+1 .. i+6*word_len] */

		for j := i+win_len*word_len+1; j < i+2*win_len*word_len; j++ {
			for k:= 0; k < len(words[j])-level+1; k++ {
				lead_ct[words[j][k:k+level]] += 1
			}
		}

		/* now, compute the bitmaps */
		lag_ct = compute_bitmap(lag_ct)
		lead_ct = compute_bitmap(lead_ct)

		/* now, compute the distance between them */
		distances = append(distances, bitmap_distance(lag_ct, lead_ct, l2_subwords))
		if distances[i] > maxval {
			max = i
		}

		for _, subword := range l2_subwords {
			delete(lag_ct, subword)
			delete(lead_ct, subword)
		}
	}

	var subword_ct = make(map[string] float32)

	for _, word := range words {
		for i := 0; i < len(word)-level+1; i++ {
			subword_ct[word[i:i+level]] += 1
		}
	}

	subword_ct = compute_bitmap(subword_ct)

	write_ppm(subword_ct, l2_subwords, "all.ppm")

	fout, err := os.Create("tsb_out.dat")
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	for i := 0; i < len(distances); i++ {
		fmt.Fprintf(fout, "%f\t%f\t%d\n", nums[i], distances[i], i)
	}


	/* Now, let's do some plotting */
	plotter, err := gnuplot.NewPlotter("", true, false)
	if err != nil {
		fmt.Printf("Error getting plotter.\n")
		fmt.Printf("Exiting...\n")
		panic(err)
	}
	defer plotter.Close()

	err = plotter.SetLabels("Time", "sample value and distance")

	plotter.CheckedCmd("set terminal png size 1500,500")
	plotter.CheckedCmd("set size 1,1")
	plotter.CheckedCmd("set output \"tsb_dist.png\"")
	plotter.CheckedCmd("set log y")
	plotter.CheckedCmd("plot \"tsb_out.dat\" using 3:1 t 'value' w lines, "+
						"\"tsb_out.dat\" using 3:2 t 'distance' w lines")

	maxout, err := os.Create("tsb_max.dat")
	if err != nil {
		panic(err)
	}
	defer maxout.Close()

	for i:= max; i < max+win_len*word_len && i < len(distances); i++ {
		fmt.Fprintf(maxout, "%f\t%f\t%d\n", nums[i], distances[i], i)
	}

	plotter.CheckedCmd("reset")
	plotter.CheckedCmd("set terminal png size 1500,500")
	plotter.CheckedCmd("set size 1,1")
	plotter.CheckedCmd("set output \"tsb_max.png\"")
	plotter.CheckedCmd("set log y")
	plotter.CheckedCmd("plot \"tsb_max.dat\" using 3:1 t 'value' w lines, "+
						"\"tsb_max.dat\" using 3:2 t 'distance' w lines")
}

func bitmap_distance(map_1, map_2 map[string] float32, l2_subwords []string) (dist float32) {
	dist = 0.0
	for _, subword := range l2_subwords {
		dist += float32(math.Pow(float64(map_1[subword]-map_2[subword]), 2))
	}
	return
}

func compute_bitmap(subword_ct map[string] float32) (subword_map map[string] float32) {
	/* get the max count */
	subword_map = make(map[string] float32)
	max := float32(0)
	for _, count := range subword_ct {
		if max < count {
			max = count
		}
	}

	for key := range subword_ct {
		subword_map[key] = subword_ct[key] / max
	}

	return
}

func write_ppm(subword_ct map[string] float32, subword_list []string,
		filename string) {

	side := 4 * 50 /* pixels per side of final image */
	var bitmap [4*50][4*50]image

	for i := 0; i < side; i++ {
		for j := 0; j < side; j++ {
			subword := subword_list[(i/50)*4+(j/50)]
			bitmap[i][j].r = int(subword_ct[subword]*255.0)
			bitmap[i][j].g = int((1.0 - subword_ct[subword])*255.0)
			bitmap[i][j].b = 0
		}
	}

	fout, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	fmt.Fprintf(fout, "P3\n")
	fmt.Fprintf(fout, "200 200\n")
	fmt.Fprintf(fout, "255\n")

	for _, row := range bitmap {
		for _, pixel := range row {
			fmt.Fprintf(fout, "%d\n", pixel.r)
			fmt.Fprintf(fout, "%d\n", pixel.g)
			fmt.Fprintf(fout, "%d\n", pixel.b)
		}
	}
}

func get_sax_words(nums []float64, l int, n int) (words []string) {
	/*
		l = 6? l = 1 hour? units are samples
		n = feature window length = 1 hour? x hours? 1 day? 
			units are samples
		for win_front from samples[0] to samples[len(samples)-n]:
			for lstart from win_front up to win_front+n-l counting by l: 
				avg = average(samples[lstart:lstart+l-1]
				append(averages, avg)
		
			sort averages to determine lettering?
			letter averages in order, this is our word.
	
		Now, we'll be normalizing the data, and we're assuming a normal 
		distribuition, so breakpoints should be at:

		-0.675 sigma
		0 sigma (the mean)
		0.675 sigma

		A: up to -0.675
		B: from -0.675 to 0
		C: from 0 to 0.675
		D: from 0.675 and up

		We will perform normalization for each word, instead of normalizing
		the entire series at once.
	*/

	var norm_data []float64

	/* This computes the average for each section */
	for win := 0; win < len(nums)-n; win++ {
		norm_data = normalize(nums[win:win+n-1])
		words = append(words, get_sax_word(norm_data, l))
	}
	return
}

func get_sax_word(data []float64, sec_length int) (word string) {
	var mean float64

	word = ""
	for sec := 0; sec < len(data); sec += sec_length {
		mean = stats.StatsMean(data[sec:sec+sec_length-1])
		word += get_symbol_4(mean, [3]float64{-0.675, 0, 0.675})
	}

	return
}

func distance(w1 string, w2 string, samp_per_sym int) float64 {
	/* My lookup table for distance between letters: */
	var dist = map[string] float64 {
		"aa": 0,	"ab": 0,	"ac": 0.67,	"ad": 1.34,
		"ba": 0,	"bb": 0,	"bc": 0,	"bd": 0.67,
		"ca": 0.67,	"cb": 0,	"cc": 0,	"cd": 0,
		"da": 1.34,	"db": 0.67,	"dc": 0,	"dd": 0,
	}

	sum := float64(0)
	for i := 0; i < len(w1); i++ {
		sum += math.Pow(dist[string(w1[i])+string(w2[i])], 2)
	}
	return math.Sqrt(float64(samp_per_sym)) * math.Sqrt(sum)
}

func normalize(data []float64) []float64 {
/* 
Before building the SAX words, we need to normalize the data. For now,
we'll normalize to a normal distribution, and perhaps change this later.

	For each data point:
		1) Subtract the mean
		2) Divide by the standard deviation

*/

	mean := stats.StatsMean(data)
	stdev := stats.StatsSampleStandardDeviation(data)

	var norm_data []float64

	for _, val := range data {
		newval := val - mean
		newval /= stdev
		norm_data = append(norm_data, newval)
	}

	return norm_data
} /* normalize */

func get_symbol_4(num float64, dist [3]float64) string {

	if num < dist[0] {
		return "a"
	} else if num >= dist[0] && num < dist[1] {
		return "b"
	} else if num >= dist[1] && num < dist[2] {
		return "c"
	}
	/* Must be greater than dist[2] */
	return "d"
}

