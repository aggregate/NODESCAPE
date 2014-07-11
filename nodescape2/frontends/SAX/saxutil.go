package saxutil

import (
	"nodescape/nsutil"
	"time"
	"io/ioutil"
	"bytes"
	"strconv"
	"math"
	"fmt"
	"runtime"
	"os/exec"
	"bitbucket.org/binet/go-gnuplot/pkg/gnuplot"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	stats "github.com/GaryBoone/GoStats/stats"
    /*  
        This package is named incorrectly. It should be named
        "github.com/GaryBoone/stats". Unfortunatedly, the author didn't 
        feel it was necessary to follow the Go package naming conventions.
    */
)
/*
type Pixel struct {
	R int
	G int
	B int
}
*/

/* Normalize our dataset, statistically speaking. */
func Normalize(data []float64) (norm_data []float64) {
	mean := stats.StatsMean(data)
	stdev := stats.StatsSampleStandardDeviation(data)

	if stdev == 0 {
		stdev = 1
	}
	for _, val := range data {
		newval := val - mean
		newval /= stdev
		norm_data = append(norm_data, newval)
	}

	return
}

/* Convert the structures we got from MySQL into slices */
func Arrays_from_rows(rows []mysql.Row, res mysql.Result) (values []float64, times []int) {

	loc, _ := time.LoadLocation("Local")

	for _, row := range rows {
		val, _ := row.FloatErr(res.Map("data"))
		ts, _ := row.TimeErr(res.Map("ctime"), loc)
		next := ts.Unix()
		values = append(values, val)
		times = append(times, int(next))
	}

	return
}

/* Get locally cached data, as opposed to data from a database */
func Arrays_from_file(filename string) (values []float64, times []int, err_out error) {
	raw, err_out := ioutil.ReadFile(filename)
	if err_out != nil {
		return
	}

	buf := string(raw)

	var part1 bytes.Buffer
	var part2 bytes.Buffer

	for i := 0; i < len(buf); i++ {
		i = eat_white(buf, i)

		for ; i < len(buf) && !white(buf[i]); i++ {
			part1.WriteString(string(buf[i]))
		}

		i = eat_white(buf, i)

		for ; i < len(buf) && !white(buf[i]); i++ {
			part2.WriteString(string(buf[i]))
		}

		num, err := strconv.ParseFloat(part1.String(), 64)
		if err != nil {
			err_out = err
			return
		}

		time, err := strconv.ParseInt(part2.String(), 10, 32)
		if err != nil {
			err_out = err
			return
		}

		values = append(values, num)
		times = append(times, int(time))
		part1.Reset()
		part2.Reset()
	}

	return
}

func Read_profile(filename string) (prof map[string] float32, err_out error) {

	prof = make(map[string] float32)

	raw, err_out := ioutil.ReadFile(filename)
	if err_out != nil {
		return
	}

	buf := string(raw)

	var part1 bytes.Buffer
	var part2 bytes.Buffer

	for i := 0; i < len(buf); i++ {
		i = eat_white(buf, i)

		for ; i < len(buf) && !white(buf[i]); i++ {
			part1.WriteString(string(buf[i]))
		}

		i = eat_white(buf, i)

		for ; i < len(buf) && !white(buf[i]); i++ {
			part2.WriteString(string(buf[i]))
		}

		subword := part1.String()

		count, err := strconv.ParseFloat(part2.String(), 32)
		if err != nil {
			err_out = err
			return
		}

		part1.Reset()
		part2.Reset()

		prof[subword] = float32(count)
	}

	return
}

/* Is the character c a whitespace character ? */
func white(c byte) bool {
	if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
		return true
	}
	return false
}

/* Increment i until white(string[i]) is false. */
func eat_white(buf string, i int) int {
	for ; i < len(buf) && white(buf[i]); i++ {
		// eat whitespace
	}
	return i
}

func Gen_words(values []float64, config nsutil.Config_t) (words []string) {
	/* 
		One day, we may support distributions other than Normal.
		For now, we're just doing Normal, and we're hardcoding it.
	*/
	var norm_values []float64

	win_len := config.Fwin_len * config.Symbol_size

	var win int
	for win = 0; win < len(values) - win_len; win++ {
		norm_values = Normalize(values[win:win + win_len])
		words = append(words, gen_word(norm_values, config))
	}
	for ; win < len(values); win++ {
		norm_values = Normalize(values[win:])
		words = append(words, gen_word(norm_values, config))
	}
	return
}

func gen_word(series []float64, config nsutil.Config_t) (word string) {
	var mean float64
	word = ""

	for sec := 0; sec+config.Symbol_size-1 < len(series); sec += config.Symbol_size {
		mean = stats.StatsMean(series[sec:sec+config.Symbol_size])
		word += get_symbol_4(mean, config)
	}

	return
}

func get_symbol_4(num float64, config nsutil.Config_t) string {

	if num < config.Dist[0] {

		return string(config.Alphabet[0])

	} else if num >= config.Dist[0] && num < config.Dist[1] {

		return string(config.Alphabet[1])

	} else if num >= config.Dist[1] && num < config.Dist[2] {

		return string(config.Alphabet[2])
	}

	return string(config.Alphabet[3])
}

func Count_subwords(words []string, config nsutil.Config_t) (subword_ct map[string] float32) {

	subword_ct = make(map[string] float32)

	for _, sub_len := range config.Subword_lengths {
		for _, word := range words {
			for i:= 0; i + sub_len < len(word) + 1; i++ {
				subword_ct[word[i:i+sub_len]] += 1
			}
		} // for range words
	} // for range config.Subword_lengths
	return
} // func Count_subwords

/* 
	Generate all possible words of length length from the alphabet
	config.Alphabet.
*/
func Gen_subwords(length int, config nsutil.Config_t) (subwords []string) {

	if length == 1 {
		for i := 0; i < len(config.Alphabet); i++ {
			subwords = append(subwords, string(config.Alphabet[i]))
		}
	} else {
		subsubwords := Gen_subwords(length - 1, config)
		for i := 0; i < len(config.Alphabet); i++ {
			for j := 0; j < len(subsubwords); j++ {
				subwords = append(subwords,
					string(config.Alphabet[i])+subsubwords[j])
			} // for j := ...
		} // for i := ...
	} // if length == 1
	return
} // func Get_subwords

/*
	Generate lists of lists of subwords
*/

func Get_subword_lists(config nsutil.Config_t) (
		subword_lists map[int] []string) {

	subword_lists = make(map[int] []string)
	for _, sub_len := range config.Subword_lengths {
		subword_lists[sub_len] = Gen_subwords(sub_len, config)
	}
	return
}

/* 
	Useful for moving the lead and lag windows over by one word.
	I can use this code instead of recounting everything.
*/
func Adj_count(subword_ct map[string] float32, remove string, add string, config nsutil.Config_t) map[string] float32 {

	new_ct := make(map[string] float32)

	for subword := range subword_ct {
		new_ct[subword] = subword_ct[subword]
	}

	for _, sub_len := range config.Subword_lengths {
		for i := 0; i + sub_len < len(add) + 1; i++ {
			new_ct[add[i:i+sub_len]] += 1
			new_ct[remove[i:i+sub_len]] -= 1
			/*
			subword_ct[add[i:i+sub_len]] += 1
			subword_ct[remove[i:i+sub_len]] -= 1
			*/
		}
	}

	return new_ct
	//return subword_ct
}

func Gen_bitmap(subword_ct map[string] float32, sub_len int, subwords []string) (bitmap map[string] float32) {

	bitmap = make(map[string] float32)

	max := float32(0)
	for _, subword := range subwords {
		if max < subword_ct[subword] {
			max = subword_ct[subword]
		}
	}

	if max != 0.0 {
		for _, subword := range subwords {
			bitmap[subword] = subword_ct[subword] / max
		}
	}
	return
}


func Bitmap_distance(map_1, map_2 map[string] float32, sub_len int, config nsutil.Config_t) (dist float32) {
	dist = 0.0
	for _, subword := range Gen_subwords(sub_len, config) {
		dist += float32(math.Pow(float64(map_1[subword]-map_2[subword]), 2))
	}
	return
}
/*
func PPM_from_bitmap(bitmap map[string] float32, sub_len int, config nsutil.Config_t) (image [][]Pixel) {

	pix_width := config.Pixel_width
	width := int(math.Pow(2, float64(sub_len)))
	subword_list := Gen_subwords(sub_len, config)
	side := width * pix_width

//	fmt.Fprintf(os.Stderr, "pix_width: %d\n", pix_width)
//	fmt.Fprintf(os.Stderr, "width: %d\n", width)
//	fmt.Fprintf(os.Stderr, "side: %d\n", side)
//	fmt.Fprintf(os.Stderr, "len(subword_list): %d\n", len(subword_list))

	for i := 0; i < side; i++ {
		var next_row []Pixel
		for j := 0; j < side; j++ {
			index := (i/pix_width)*width+(j/pix_width)
			//fmt.Fprintf(os.Stderr, "i: %d\tj: %d\tindex: %d\n", i, j, index)
			subword := subword_list[index]
			var next_pixel Pixel
			next_pixel.R = int(bitmap[subword]*255.0)
			next_pixel.G = int((1.0 - bitmap[subword])*255.0)
			next_pixel.B = 0
			next_row = append(next_row, next_pixel)
		}
		image = append(image, next_row)
	}
	return
}
*/

/*
func Write_PPM(image [][]Pixel, filename string) (err_out error) {

	fout, err := os.Create(filename)
	if err != nil {
		err_out = err
		return
	}
	defer fout.Close()

	fmt.Fprintf(fout, "P3\n")
	fmt.Fprintf(fout, "%d %d\n", len(image), len(image[0]))
	fmt.Fprintf(fout, "255\n")

	for _, row := range image {
		for _, pixel := range row {
			fmt.Fprintf(fout, "%d\n%d\n%d\n", pixel.R, pixel.G, pixel.B)
		}
	}
	return
}
*/

func Word_distance(w1 string, w2 string, config nsutil.Config_t) float64 {
	/* My lookup table for distance between letters: */
    var dist = map[string] float64 {
        "aa": 0,    "ab": 0,    "ac": 0.67, "ad": 1.34,
        "ba": 0,    "bb": 0,    "bc": 0,    "bd": 0.67,
        "ca": 0.67, "cb": 0,    "cc": 0,    "cd": 0,
        "da": 1.34, "db": 0.67, "dc": 0,    "dd": 0,
    }

    sum := float64(0)
    for i := 0; i < len(w1); i++ {
        sum += math.Pow(dist[string(w1[i])+string(w2[i])], 2)
    }
    return math.Sqrt(float64(config.Symbol_size)) * math.Sqrt(sum)

}

type proc_res struct {
	dist map[int] float32
	dist_mul float64
	pos int
}

func Proc_window(lag_subct, lead_subct map[string] float32, lead_end int,
				config nsutil.Config_t, subword_lists map[int] []string,
				ch chan proc_res) {

	var res proc_res

    // Get the offset for distance (in symbols)
    offset := float32(config.Fwin_len)/2.0 - float32(config.Lead_length)/2.0

    // Set the position for this distance computation (in samples)
    res.pos = lead_end + int(float32(config.Symbol_size) * offset)

    //res.pos = lead_end + config.Dist_offset
    res.dist = make(map[int] float32)

    res.dist_mul = float64(1)
    for _, sub_len := range config.Subword_lengths {
        lag_map := Gen_bitmap(lag_subct, sub_len,
                                        subword_lists[sub_len])
        lead_map := Gen_bitmap(lead_subct, sub_len,
                                        subword_lists[sub_len])

        res.dist[sub_len] = Bitmap_distance(lag_map, lead_map,
											sub_len, config)
        res.dist_mul *= float64(res.dist[sub_len])
   }

    ch <- res
}

func Baseline(words []string, series []map[string] float32,
	base map[string] float32, subword_lists map[int] []string,
	config nsutil.Config_t) (dist map[int] float64) {

	lead_start := 0
	lead_end := lead_start + config.Lead_length * config.Symbol_size

	NCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(NCPU)

	ch := make(chan proc_res)

	for lead_end <= len(words) {
		go Proc_window(base, series[lead_start], lead_end, config,
						subword_lists, ch)

		if lead_end < len(words) {
			series = append(series, Adj_count(series[lead_start],
							words[lead_start], words[lead_end], config))
		}
		lead_start += 1
		lead_end += 1
	}

	dist = make(map[int] float64)

	for i := 0; i < lead_start; i++ {
		res := <-ch
		dist[res.pos - 1] = res.dist_mul
	}

	return
}

func Baseline_norm(words []string, series []map[string] float32,
	base map[string] float32, subword_lists map[int] []string,
	config nsutil.Config_t) (dist map[int] float64) {

	lead_start := 0
	lead_end := lead_start + config.Lead_length * config.Symbol_size

	NCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(NCPU)

	ch := make(chan proc_res)

	for lead_end <= len(words) {
		go Proc_window(base, series[lead_start], lead_end, config,
						subword_lists, ch)

		if lead_end < len(words) {
			series = append(series, Adj_count(series[lead_start],
							words[lead_start], words[lead_end], config))
		}
		lead_start += 1
		lead_end += 1
	}

	dist = make(map[int] float64)

	maxdist := 0.0
	for i := 0; i < lead_start; i++ {
		res := <-ch
		dist[res.pos - 1] = res.dist_mul
		if res.dist_mul > maxdist {
			maxdist = res.dist_mul
		}
	}

	for i := range dist {
		dist[i] /= maxdist
	}

	return
}

func Lead_lag(words []string, lag_subct, lead_subct []map[string] float32,
		subword_lists map[int] []string, config nsutil.Config_t) (dist map[int] float64) {

	lag_start := 0
	lead_start := lag_start + config.Lead_length * config.Symbol_size
	lead_end := lead_start + config.Lead_length * config.Symbol_size

	NCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(NCPU)

	ch := make(chan proc_res)

	for lag_start = 0; lead_end <= len(words); {
		go Proc_window(lag_subct[lag_start], lead_subct[lag_start],
				lead_end, config, subword_lists, ch)
		lag_subct = append(lag_subct, Adj_count(lag_subct[lag_start],
							words[lag_start], words[lead_start], config))
		if lead_end < len(words) {
			lead_subct = append(lead_subct,Adj_count(lead_subct[lag_start],
								words[lead_start], words[lead_end], config))
		}

		lag_start += 1
		lead_start += 1
		lead_end += 1
	}

	dist = make(map[int] float64)

	for i := 0; i < lag_start; i++ {
		res := <-ch
		dist[res.pos - 1] = res.dist_mul
	}

	return
}

func Get_hosts(db mysql.Conn, config nsutil.Config_t) (hosts []string) {
	query_str := fmt.Sprintf("select distinct host from %s;",
					config.Sql_table)
	rows, res, err := db.Query(query_str)
	if err != nil {
		hosts = nil
	} else {
		for _, row := range rows {
			hosts = append(hosts, row.Str(res.Map("host")))
		}
	}
	return
}

func Get_labels(db mysql.Conn, host string, config nsutil.Config_t) (labels []string) {
	query_str := fmt.Sprintf(
					"select distinct label from %s where host = \"%s\";",
					config.Sql_table, host)

	rows, res, err := db.Query(query_str)
	if err != nil {
		labels = nil
	} else {
		for _, row := range rows {
			labels = append(labels, row.Str(res.Map("label")))
		}
	}
	return
}

func Plot2(datafile, out_prefix string, xcol, ycol int, xlabel, ylabel, legend string) (err error) {
	plotter, err := gnuplot.NewPlotter("", true, false)
	if err != nil {
		return
	}

	plotter.CheckedCmd("set terminal png small size 1000,150")
	plotter.CheckedCmd("set size 1,1")
	plotter.CheckedCmd("set xdata time")
	plotter.CheckedCmd("set timefmt \"%%s\"")
	plotter.CheckedCmd("set format x \"%%m/%%d %%R\"")

	plotter.SetLabels(xlabel, ylabel)
	plotter.CheckedCmd("set output \"%s.png\"", out_prefix)
	plotter.CheckedCmd("plot \"%s\" u %d:%d t '%s' w lines",
						datafile, xcol, ycol, legend)
	plotter.Close()

	err = exec.Command("convert", "-scale", "200x70!",
						out_prefix+".png", out_prefix+"-small.png").Run()
	return
}
