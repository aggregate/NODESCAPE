package main

import (
	"os"
	"io/ioutil"
	"fmt"
	"bytes"
	"strconv"
	"errors"
	"math"
	"bitbucket.org/binet/go-gnuplot/pkg/gnuplot"
	"github.com/GaryBoone/GoStats"
	/* 
		This package is named incorrectly. It should be named
		"github.com/GaryBoone/stats". Unfortunatedly, the author didn't 
		feel it was necessary to follow the Go package naming conventions.
	*/
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <file>\n", os.Args[0])
		os.Exit(0)
	}

	nums, times := read_data(os.Args[1])

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


	var l = 8 // number of samples per symbol
	var n = 4*l // width of feature window
	var averages []float64
	var norm_data []float64
	var words []string
	var mean float64

	/* This computes the average for each section */
	for win := 0; win < len(nums)-n; win++ {
		word := ""
		norm_data = normalize(nums[win:win+n-1])
		for sec := 0; sec < len(norm_data)-l; sec += l {
			mean = stats.StatsMean(norm_data[sec:sec+l-1])
			word += get_symbol_4(mean, [3]float64{-0.675, 0, 0.675})
			averages = append(averages, mean)
		}
		words = append(words, word)
		word = ""
	}

	var distances []float64

	for i := 0; i < len(words)-1; i++ {
		distances = append(distances, distance(words[i], words[i+1], l))
	}

	/* Before plotting, we want to dump our data to a file */

	fout, err := os.Create("out.dat")
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	for i := 0; i < len(distances); i++ {
		fmt.Fprintf(fout, "%f\t%f\t%d\n", nums[i], distances[i], times[i])
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

	plotter.CheckedCmd("set terminal png size 2000,200")
	plotter.CheckedCmd("set size 1,1")
	plotter.CheckedCmd("set output \"value.png\"")
	plotter.CheckedCmd("plot \"out.dat\" using 3:1 t 'value' w lines ")
	plotter.CheckedCmd("reset")

	err = plotter.SetLabels("Time", "sample value and distance")

	plotter.CheckedCmd("set terminal png size 2000,200")
	plotter.CheckedCmd("set size 1,1")
	plotter.CheckedCmd("set output \"distance.png\"")
	plotter.CheckedCmd("plot \"out.dat\" using 3:2 t 'distance' w lines")

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

func average(nums []float64) (error, float64) {
	if len(nums) == 0 {
		return errors.New("Length of list is zero"), 0
	}

	var total float64 = 0
	for _,num := range nums {
		total += num;
	}

	return nil, total/float64(len(nums))
}

func read_data(filename string) ([]float64, []int) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	buf := string(raw)

	var part1 bytes.Buffer
	var part2 bytes.Buffer

	var nums []float64
	var times []int

	var toff = -1

	for i := 0; i < len(buf); i++ {
		i = eatws(buf, i)

		for ; i < len(buf) && !isws(buf[i]); i++ {
			part1.WriteString(string(buf[i]))
		}

		i = eatws(buf, i)

		for ; i < len(buf) && !isws(buf[i]); i++ {
			part2.WriteString(string(buf[i]))
		}

		num, err := strconv.ParseFloat(part1.String(), 64)
		time, err := strconv.ParseInt(part2.String(), 10, 32)
		if err != nil {
			fmt.Printf("We're ignoring this\n")
		}
		if toff == -1 {
			toff = int(time)
		}
		nums = append(nums, num)
		times = append(times, int(time)-toff)
		part1.Reset()
		part2.Reset()
	}

	return nums, times
}


func isws(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func eatws(buf string, i int) int {
	for ; i < len(buf) && isws(buf[i]); i++ {
		// eat whitespace
	}
	return i
}
