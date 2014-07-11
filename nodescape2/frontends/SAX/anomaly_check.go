/*
	Get list of conf files.
	For each conf file, get baselines
	Run analysis for each baseline
	Generate anomaly page for each anomaly?
		Give sample date
		Give baseline analysis for all baselines
		Perform and give lead/lag analysis
	Generate new landing page
		List of anomalies with scaled down images
	+-------------------+-----------+------------+
	| Machine name      | signal    | anomaly    |
	+-------------------+-----------+------------+
	| Machine name      | signal    | anomaly    |
	+-------------------+-----------+------------+
*/

package main

import (
	"nodescape/nsutil"
	"nodescape/saxutil"
	"fmt"
	"os"
	"os/exec"
	"bytes"
	"regexp"
	"runtime"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
)

func get_baselines(host, property string, config nsutil.Config_t) (baselines map[string] map[string] float32, err error){

	/*
		Check for profiles here. If no profiles, then
		don't do a DB query.
	*/
	baselines = make(map[string] map[string] float32)
	cmd := exec.Command("ls", "-1", config.Profdir)

	out, err := cmd.Output()
	if err != nil {
		return
	}

	line_read := bytes.NewBuffer(out)
	line, read_err := line_read.ReadString('\n')
	/* Perhaps I want to use bytes.Split?*/
	for read_err == nil {
		match, _ := regexp.MatchString(
						nsutil.Pre_name(host, property)+".*", line)
		if match {
			next, err := saxutil.Read_profile(
								config.Profdir+line[0:len(line)-1])
			if err != nil {
				fmt.Fprintf(os.Stderr,"Could not read profile %s\n",
					line[0:len(line)-1])
				fmt.Print(err)
				continue
			}
			baselines[line[0:len(line)-1]] = next
		} // if match
		line, read_err = line_read.ReadString('\n')
	}
	return
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config file>\n", os.Args[0])
		os.Exit(1)
	}

	config, err := nsutil.Read_config(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading configuration file.\n")
		os.Exit(1)
	}
	db := mysql.New("tcp", "", config.Sql_host+":"+config.Sql_port,
			config.Sql_user, config.Sql_passwd, config.Sql_db)

	err = db.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to database.\n")
		os.Exit(1)
	}

	subword_lists := make(map[int] []string)
	for _, sub_len := range config.Subword_lengths {
		subword_lists[sub_len] = saxutil.Gen_subwords(sub_len, config)
	}

	NCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(NCPU)

	for _, host := range saxutil.Get_hosts(db, config) {
		fmt.Fprintf(os.Stderr, "Handling %s\n", host);

		for _, property := range saxutil.Get_labels(db, host, config) {
			fmt.Fprintf(os.Stderr, "\t%s\n", property)

			/*
				Check for profiles here. If no profiles, then
				don't do a DB query.
			*/

			baselines, err := get_baselines(host, property, config)
			if len(baselines) <= 0 {
				continue
			}

			rows, res, err := nsutil.Get_data(host, property, config)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not get data for %s:%s\n",
						host, property)
				continue
			}

			values, times := saxutil.Arrays_from_rows(rows, res)
			var dump_arr [][]float64
			dump_arr = append(dump_arr, values)

//			words := saxutil.Gen_words(saxutil.Normalize(values),config)
			words := saxutil.Gen_words(values, config)

			fmt.Fprintf(os.Stderr, "Got words\n")

			lag_start := 0
			lead_st_ll := lag_start+config.Lag_length*config.Symbol_size
			lead_end_ll := lead_st_ll+config.Lead_length*config.Symbol_size
			var lag_subct []map[string] float32
			var lead_ll_subct []map[string] float32
			lag_subct = append(lag_subct, saxutil.Count_subwords(
									words[lag_start:lead_st_ll], config))
			lead_ll_subct = append(lead_ll_subct, saxutil.Count_subwords(
									words[lead_st_ll:lead_end_ll], config))

			lead_start := 0
			lead_end := lead_start + config.Lead_length * config.Symbol_size
			var lead_subct []map[string] float32
			lead_subct = append(lead_subct, saxutil.Count_subwords(
									words[lead_start:lead_end], config))


			dist_ll := saxutil.Lead_lag(words, lag_subct,lead_ll_subct,
							subword_lists, config)
			var col_ll []float64
			for i := 0; i < len(words) + 1; i++ {
				col_ll = append(col_ll, dist_ll[i])
			}
			dump_arr = append(dump_arr, col_ll)

			var anomalies []string
			for basename, baseline := range baselines {

				fmt.Fprintf(os.Stderr, "Considering %s\n", basename)
				dist_mul := saxutil.Baseline(words, lead_subct, baseline,
								subword_lists, config)

				var next_column []float64
				for i := 0; i < len(words) + 1; i++ {
					next_column = append(next_column, dist_mul[i])
				}

				dump_arr = append(dump_arr, next_column)
				anomalies = append(anomalies, basename)
			} // baseline in baselines
			// Baselines in dump_arr have the same ordering as their
			// names in anomalies.


			nsutil.Text_dump(dump_arr, times, nsutil.Pre_name(host, property)+".dat", words, config)
			html, err := os.Create(config.Tempdir+nsutil.Pre_name(host, property)+".html")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not create html file.\n")
				fmt.Print(err)
				continue
			}

			err = saxutil.Plot2(
					nsutil.Pre_name(host, property)+".dat",
					config.Tempdir+"/assets/"+nsutil.Pre_name(host, property),
					len(dump_arr)+1, 1,
					"Times", "Sample Values", "Sample Values")
			/* Build graphs */

			fmt.Fprintf(html, "<html>\n<body>\n")
			fmt.Fprintf(html, "<h1><center>%s %s</center></h1><br>\n",
								host, property)
			fmt.Fprintf(html, "<hr>")
			fmt.Fprintf(html, "<img src=\"assets/%s.png\">\n",
								nsutil.Pre_name(host, property))

			/* Plot Lead/Lag analysis */
			err = saxutil.Plot2(
					nsutil.Pre_name(host, property)+".dat",
					config.Tempdir+"/assets/"+
						nsutil.Pre_name(host, property)+"ll",
					len(dump_arr)+1, 2,
					"Times", "Distance Score", "Lead/Lag Analysis")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Plot failed\n")
				fmt.Print(err)
				continue
			}
			fmt.Fprintf(html, "%s\n", "Lead/Lag Analysis")
			fmt.Fprintf(html, "<img src=\"assets/%s\">\n",
							nsutil.Pre_name(host, property)+"ll"+".png")
			fmt.Fprintf(html, "<hr>\n")


			for i, anomaly := range anomalies {
				err = saxutil.Plot2(
						nsutil.Pre_name(host, property)+".dat",
						config.Tempdir+"/assets/"+anomaly,
						len(dump_arr)+1, i+3,
						"Times", "Distance Score", anomaly)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Plot failed\n")
					fmt.Print(err)
					continue
				}
				fmt.Fprintf(html, "%s\n", anomaly)
				fmt.Fprintf(html, "<img src=\"assets/%s\">\n",
								anomaly+".png")
				fmt.Fprintf(html, "<hr>\n")
			}
			fmt.Fprintf(html, "</body>\n</html>\n")
		}
	}


}

