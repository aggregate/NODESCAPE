package main

import (
	"github.com/mjibson/go-dsp/fft"
	"os"
	"io/ioutil"
	"fmt"
	"bytes"
	"strconv"
	"bitbucket.org/binet/go-gnuplot/pkg/gnuplot"
)

func main() {
	if len(os.Args) != 2 {
		panic(fmt.Sprintf("Usage: ./%s <file>\n", os.Args[0]))
	}
	raw, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	buf := string(raw)

	fout, err := os.Create("out.txt")
	if err != nil {
		panic(err)
	}
	defer fout.Close()


	var part1 bytes.Buffer
	var part2 bytes.Buffer

	var nums []float64

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
		if err != nil {
			fmt.Printf("We're ignoring this\n")
		}
		nums = append(nums, num)
		part1.Reset()
		part2.Reset()
	}

	var fft_out []complex128

	fft_out = fft.FFTReal(nums)

	var min_real = real(fft_out[0])
	for i,num := range fft_out {
		fmt.Fprintf(fout, "%f\t%f\t%f\n", 
			float64(i)*float64(0.5)/float64(len(fft_out)), 
				real(num), imag(num))
		if real(num) < min_real {
			min_real = real(num)
		}
	}
	fout.Close()

	// get a gnuplot plotter
	plotter, err := gnuplot.NewPlotter("", true, false)
	if err != nil {
		fmt.Printf("Error getting plotter!\n")
		fmt.Printf("Exiting...\n")
		panic(err)
	}

	defer plotter.Close()

	// use Setlabels
	err = plotter.SetLabels("Frequency (fraction of sampling frequency",
							"Amplitude log((not sure of units))")

	// use PlotXY
	//err = plotter.PlotXY(freqs, amplitudes, "DFT of some data")

	plotter.CheckedCmd("set logscale xy")
	plotter.CheckedCmd("plot \"out.txt\" using 1:2 t 'real' w impulses")

	// use CheckedCmd to generate a PNG
	plotter.CheckedCmd("set terminal png")
	plotter.CheckedCmd("set output \"real.png\"")
	plotter.CheckedCmd("replot")

	plotter.CheckedCmd("reset")

	err = plotter.SetLabels("Fraction of sampling frequency",
							"Amplitude log((not sure of units))")
	plotter.CheckedCmd("set logscale y")
	plotter.CheckedCmd("plot \"out.txt\" using 1:3 t 'imag' w impulses")
	plotter.CheckedCmd("set terminal png")
	plotter.CheckedCmd("set output \"imag.png\"")
	plotter.CheckedCmd("replot")

	fmt.Printf("Min real freq: %f\n", min_real)

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
