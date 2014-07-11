package main

import (
	"os"
	"fmt"
	"math"
	"bitbucket.org/binet/go-gnuplot/pkg/gnuplot"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"github.com/GaryBoone/GoStats/stats"
	/* 
		This package is named incorrectly. It should be named
		"github.com/GaryBoone/stats". Unfortunatedly, the author didn't 
		feel it was necessary to follow the Go package naming conventions.
	*/
)

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

	rows, res, err := db.Query("select data from ukanstats where "+
						"host = \""+os.Args[1]+"\" and label = \""+
						os.Args[2]+"\" order by ctime asc;")

	if err != nil {
		panic(err)
	}

	var histogram = make(map[float64] int)

	/* Build histogram */
	for _, row := range rows {
		val := row.Float(res.Map("data"))
		histogram[val] += 1
	}

	fout, err := os.Create("histogram.dat")
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	for key, val := range histogram {
		fmt.Fprintf(fout, "%f\t%d\n", key, val)
	}

	/* Now, let's do some plotting */

	plotter, err := gnuplot.NewPlotter("", true, false)
	if err != nil {
		fmt.Printf("Error getting plotter.\n")
		fmt.Printf("Exiting...\n")
		panic(err)
	}
	defer plotter.Close()

	err = plotter.SetLabels("Value", "Count")

	plotter.CheckedCmd("set terminal png size 800,800")
	plotter.CheckedCmd("set size 1,1")
	plotter.CheckedCmd("set output \"histogram.png\"")
	plotter.CheckedCmd("binwidth=1")
	plotter.CheckedCmd("set boxwidth binwidth")
	plotter.CheckedCmd("plot \"histogram.dat\" using 1:2 t 'Histogram' smooth freq w boxes ")

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

