package nsutil
import (
	"io/ioutil"
	"strings"
	"fmt"
	"os"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"strconv"
	//"math"
	//"time"
	//"bitbucket.org/binet/go-gnuplot/pkg/gnuplot"
	//"github.com/GaryBoone/GoStats"
	/* 
		This package is named incorrectly. It should be named
		"github.com/GaryBoone/stats". Unfortunatedly, the author didn't 
		feel it was necessary to follow the Go package naming conventions.
	*/
)

type Config_t struct {
	Sql_host string
	Sql_port string
	Sql_user string
	Sql_passwd string
	Sql_db string
	Sql_table string
	Symbol_size int
	Dist [3]float64
	Alphabet string
	Fwin_len int
	Subword_lengths []int
//	Pixel_width int
	Lead_length int
	Lag_length int
	Verbose bool
//	Dist_offset int
	DB_hist int
	Profiles []string
	Baseline string
	Tempdir string
	Webdir string
	Profdir string
}

func Read_config(filename string) (config Config_t, err_out error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		err_out = err
		return
	}

	file_str := string(raw)

	/*
		Parse the string one line at a time.
	*/

	var tokens [][]string // list of lines broken into tokens
	var next_line []string // tokens on next line
	var next_tok []string // building next token

	for i := 0; i < len(file_str); {

		next_line = []string{}
		for i < len(file_str) &&
			file_str[i] != '\n' {

			next_tok = []string{}

			for i < len(file_str) &&
				file_str[i] != '\n' &&
				file_str[i] != '#' &&
				white(file_str[i]) {

				i += 1
			}

			for i < len(file_str) &&
				file_str[i] != '\n' &&
				file_str[i] != '#' &&
				!white(file_str[i]) {

				next_tok = append(next_tok, string(file_str[i]))
				i += 1
			}

			/* Prevents adding an empty token on a comment line */
			if len(next_tok) > 0 {
				next_line = append(next_line, strings.Join(next_tok, ""))
			}

			/* handle a comment if we found one */
			if file_str[i] == '#' {
				for i < len(file_str) &&
					file_str[i] != '\n' {
					i += 1
				}
			}

		}
		tokens = append(tokens, next_line)
		i += 1

	} // for c:= range file_str


	lnum := 1
	for _, line := range tokens {
		if len(line) > 0 {
			switch line[0] {
			case "Sql_host":
				if len(line) < 2 {
					fmt.Fprintf(os.Stderr,
						"%d: Option Sql_host requires a value.\n", lnum)
				} else {
					config.Sql_host = line[1]
				}
			case "Sql_port":
				if len(line) < 2 {
					fmt.Fprintf(os.Stderr,
						"%d: Option Sql_port requires a value.\n", lnum)
				} else {
					config.Sql_port = line[1]
				}
			case "Sql_user":
				if len(line) < 2 {
					fmt.Fprintf(os.Stderr,
						"%d: Option Sql_user requires a value.\n", lnum)
				} else {
					config.Sql_user = line[1]
				}
			case "Sql_passwd":
				if len(line) < 2 {
					fmt.Fprintf(os.Stderr,
						"%d: Option Sql_passwd requires a value.\n", lnum)
				} else {
					config.Sql_passwd = line[1]
				}
			case "Sql_db":
				if len(line) < 2 {
					fmt.Fprintf(os.Stderr,
						"%d: Option Sql_db requires a value.\n", lnum)
				} else {
					config.Sql_db = line[1]
				}
			case "Sql_table":
				if len(line) < 2 {
					fmt.Fprintf(os.Stderr,
						"%d: Option Sql_table requires a value.\n", lnum)
				} else {
					config.Sql_table = line[1]
				}

			case "Symbol_size":
				config.Symbol_size = get_int(line, lnum)

			case "Fwin_len":
				config.Fwin_len = get_int(line, lnum)

			case "Dist":
				if len(line) < 2 {
					fmt.Fprintf(os.Stderr,
						"%d: Option Dist Requires a value.\n", lnum)
				} else {
					switch line[1] {
					case "normal":
						config.Dist = [3]float64{-0.675, 0, 0.675}
					default:
						fmt.Fprintf(os.Stderr,
							"%d: Distribution %s not supported. ",
							line[1])
						fmt.Fprintf(os.Stderr,
							"Defaulting to Normal distribution.\n")
						config.Dist = [3]float64{-0.675, 0, 0.675}
					}
				}
			case "Alphabet":
				if len(line) < 2 {
					fmt.Fprintf(os.Stderr,
						"%d: Option Alphabet requires a value.\n", lnum)
				} else {
					if len(line[1]) < 4 {

						config.Alphabet = "abcd"
						fmt.Fprintf(os.Stderr,
							"Only alpbabets of length 4 are supported. ")
						fmt.Fprintf(os.Stderr,
							"Substituting \"%s\" as the alpbabet.\n",
							config.Alphabet)

					} else if len(line[1]) > 4 {

						config.Alphabet = line[1][0:4]
						fmt.Fprintf(os.Stderr,
							"Only alpbabets of length 4 are supported. ")
						fmt.Fprintf(os.Stderr,
							"Truncating the alpbabet to \"%s\"\n",
							config.Alphabet)

					} else {

						config.Alphabet = line[1]

					}
				} // if len(line) < 2
			case "Subword_lengths":
				if len(line) < 2 {
					fmt.Fprintf(os.Stderr,
						"%d: Option Subword_lengths requires at least "+
						"one value.\n", lnum)
				} else {
					for i := 1; i < len(line); i++ {
						length, err := strconv.ParseInt(line[i], 10, 32)
						if err != nil {
							fmt.Fprintf(os.Stderr,
								"%d: Could not convert value to integer.\n",
								lnum)
						}
						config.Subword_lengths =
							append(config.Subword_lengths, int(length))
					}
				}
/*
			case "Pixel_width":
				config.Pixel_width = get_int(line, lnum)
*/

			case "Lead_length":
				config.Lead_length = get_int(line, lnum)

			case "Lag_length":
				config.Lag_length = get_int(line, lnum)

			case "Verbose":
				config.Verbose = (len(line) > 1 && line[1] == "on")
/*
			case "Dist_offset":
				config.Dist_offset = get_int(line, lnum)
*/

			case "DB_hist":
				if len(line) >= 3 {
					config.DB_hist = get_seconds(get_int(line,lnum),line[2])
				} else {
					fmt.Fprintf(os.Stderr,
						"%d: Option DB_hist requires at quantity "+
						"and a unit\n", lnum)
				}

			case "Profile":
				if len(line) >= 2 {
					for i := 1; i < len(line); i++ {
						config.Profiles = append(config.Profiles, line[i])
					}
				} else {
					fmt.Fprintf(os.Stderr,
						"%d: Option Profiles requires at least one profile"+
						" file name\n", lnum)
				}
			case "Baseline":
				if len(line) >= 2 {
					config.Baseline = line[1]
				} else {
					fmt.Fprintf(os.Stderr,
						"%d: Option Baseline requires a profile file name\n",
						lnum)
				}
			case "Tempdir":
				if len(line) >= 2 {
					config.Tempdir = line[1] + "/"
				} else {
					fmt.Fprintf(os.Stderr,
						"%d: Option Tempdir requires a directory\n", lnum)
				}
			case "Profdir":
				if len(line) >= 2 {
					config.Profdir = line[1] + "/"
				} else {
					fmt.Fprintf(os.Stderr,
						"%d: Option Profdir requires a directory\n", lnum)
				}
			case "Webdir":
				if len(line) >= 2 {
					config.Webdir = line[1] + "/"
				} else {
					fmt.Fprintf(os.Stderr,
						"%d: Option Webdir requires a directory\n", lnum)
				}
			/*
				To extend this config file parser, just add
				a case to process the line (or lines) for your option.
				Don't forget to check for fields before you access them,
				and make sure to print to os.Stderr, not os.Stdout.
			*/

			default:
				fmt.Fprintf(os.Stderr,
					"%d: Option %s not implemented. Ignoring.\n",
					lnum, line[0])
			} // switch line[0]
		} // if len(line) > 0
		lnum += 1;
	}

	err_out = nil
	return
} // func Read_config

func get_int(line []string, lnum int) (val int) {
	if len(line) < 2 {
		fmt.Fprintf(os.Stderr,
			"%d: Option %s requires a value.\n", lnum, line[0])
	} else {
		val64, err := strconv.ParseInt(line[1], 10, 32)
		val = int(val64)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"%d: Could not convert value to integer.\n", lnum)
		}
	}
	return
}

func get_seconds(quantity int, unit string) int {

	secs_per := map[string] int {
				"second" : (1),
				"minute" : (60),
				"hour"  : (60*60),
				"day"   : (60*60*24),
				"week"  : (60*60*24*7),
				"month" : (60*60*24*7*30), // I know, they've got 31, 29, 28
				"year"	: (60*60*24*7*52),
			}

	return (secs_per[unit] * quantity)
}

/* Is the character c a whitespace character? */
func white(c byte) bool {
	if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
		return true
	}
	return false
}

func Get_data(host string, property string, config Config_t) (rows []mysql.Row, res mysql.Result, err error) {

	db := mysql.New("tcp", "", config.Sql_host+":"+config.Sql_port,
					config.Sql_user, config.Sql_passwd, config.Sql_db)

	err = db.Connect()
	if err != nil {
		return
	}
	query_str := fmt.Sprintf("select data, ctime from %s "+
						"where host = \"%s\" and label = \"%s\" "+
						"and ctime > date_sub(now(), interval %d second) "+
						"order by ctime asc;",
						config.Sql_table, host, property, config.DB_hist)

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "DB_hist: %d\n", config.DB_hist)
		fmt.Fprintf(os.Stderr, "%s\n", query_str)
	}

	rows, res, err = db.Query(query_str)

	return
}

/*
	Build a string that can be used as a good name for creating files.
	Concatenate the host and property with an underscore, and remove
	whitespace characters (assumes whitespace is only \t and space).
*/
func Pre_name(host string, property string) (name string) {
	name = host + "_" + property
	name = strings.Replace(name, " ", "", -1)
	name = strings.Replace(name, "\t", "", -1)
	return
}

/*
	Dump data to a text file. A list of samples (or processed values) and
	a list of x values. Put them in nice columns.
*/
/* 
	TODO: Add some error handling for when the user gets their dimensions 
	mismatched.
*/
func Text_dump(values [][]float64, xvals []int, filename string, words []string, config Config_t) {
	fout, _ := os.Create(filename)

	length := len(words)
	for i := range values {
		if length > len(values[i]) {
			length = len(values[i])
		}
	}

	/* The indexing here is definitely sub-optimal, but it's not natural
		to have the series stored in row major order */
	for i := 0; i < length; i++ {
		for j := range values {
			fmt.Fprintf(fout, "%f\t", values[j][i])
		}
		fmt.Fprintf(fout, "%d", xvals[i])

		if config.Verbose {
			fmt.Fprintf(fout, "\t%s\n", words[i])
		} else {
			fmt.Fprintf(fout, "\n")
		}
	}
}

func Full_dump(values [][]float64, xvals []int, filename string, words []string, key map[int] int, config Config_t) {
	fout, _ := os.Create(filename)

	length := len(words)
	for i := range values {
		if length > len(values[i]) {
			length = len(values[i])
		}
	}

	/* The indexing here is definitely sub-optimal, but it's not natural
		to have the series stored in row major order */
	for i := 0; i < length; i++ {
		for j := range values {
			fmt.Fprintf(fout, "%f\t", values[j][i])
		}
		fmt.Fprintf(fout, "%d", xvals[i])

		if config.Verbose {
			fmt.Fprintf(fout, "\t%s", words[i])
			fmt.Fprintf(fout, "\t%d\n", key[i])
		} else {
			fmt.Fprintf(fout, "\n")
		}
	}
}
