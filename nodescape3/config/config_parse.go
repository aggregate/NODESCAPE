package config

import (
	"io"
	"os"
	"bytes"
	"fmt"
)

type ConfigParser struct {
	scanner scanner_t
	next token
	config config_root
}

type config_root struct {
	sects []sect_node
}

type sect_node struct {
	name string
	opts []opt_node
}

type opt_node struct {
	key string
	vals []token
}

func init_parser(in io.Reader) (ConfigParser, error) {
	scanner, err := init_scanner(in)
	if err != nil {
		return ConfigParser{
					scanner,
					*new(token),
					*new(config_root),
					}, err
	}
	next := scanner.scan()

	return ConfigParser{scanner, next, *new(config_root)}, nil
}

type ConfigError struct {
	err string
}

func (ce *ConfigError) Error() string {
	return ce.err
}

func type_str(t int) string {
	switch t {
		case start: return "start"
		case lb: return "lb"
		case rb: return "rb"
		case nl: return "nl"
		case id: return "id"
		case arg: return "arg"
		case cm: return "cm"
		case eof: return "eof"
		default: return "<invalid type>"
	}
	return "<invalid type>"
}

func (cp *ConfigParser) match(term int) (token) {
	found := cp.next
	if term == found.t_type {
		cp.next = cp.scanner.scan()
	} else {
		fmt.Fprintf(os.Stderr, "%d:%d: Could not match token!\n",
					found.line, found.column)
		fmt.Fprintf(os.Stderr, "\tExpected type: %s\n", type_str(term))
		fmt.Fprintf(os.Stderr, "\tFound type: %s\n", type_str(found.t_type))
		os.Exit(1)
	}

	return found
}

func (cp *ConfigParser) peek(term int) bool {
	return cp.next.t_type == term
}


func parse_config(in io.Reader) ConfigParser {
	cp, err := init_parser(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize parser\n")
	}
	cp.config = cp.parse_C()

/*
	for _, sect := range cp.config.sects {
		fmt.Printf("Section: %s\n", sect.name)
		for _, opt := range sect.opts {
			fmt.Printf("\tkey: %s\n", opt.key)
			for _, val := range opt.vals {
				fmt.Printf("\t\tval: %s\n", val.value)
			}
		}
	}
*/

	return cp
}

func (cp *ConfigParser) parse_C() config_root {
	var config config_root
	config.sects = append(config.sects, cp.parse_P())
	for cp.peek(nl) {
		cp.match(nl)
	}
	if cp.peek(lb) {
		config_t := cp.parse_C()
		for _, sect := range config_t.sects {
			config.sects = append(config.sects, sect)
		}
	}

	return config
}

func (cp *ConfigParser) parse_P() sect_node {
	var sect sect_node
	sect.name = cp.parse_I()
	sect.opts = cp.parse_MK()

	return sect
}

func (cp *ConfigParser) parse_I() string {
	cp.match(lb)
	var prop_id bytes.Buffer
	for cp.peek(id) {
		prop_id.WriteString(cp.match(id).value+" ")
	}
	prop_id.Truncate(prop_id.Len()-1)
	cp.match(rb)
	cp.match(nl) // Require at least one newline here
	for cp.peek(nl) {
		cp.match(nl)
	}
	return prop_id.String()
}

func (cp *ConfigParser) parse_MK() []opt_node {
	var opts []opt_node
	opts = append(opts, cp.parse_K())
	for cp.peek(nl) {
		cp.match(nl)
	}
	if cp.peek(id) {
		opts_temp := cp.parse_MK()
		for _, opt := range opts_temp {
			opts = append(opts, opt)
		}
	}

	return opts
}

func (cp *ConfigParser) parse_K() opt_node {
	var opt opt_node
	opt.key = cp.match(id).value
	opt.vals = cp.parse_A()
	return opt
}

func (cp *ConfigParser) parse_A() []token {
	var args []token
	for !(cp.peek(cm) || cp.peek(nl) || cp.peek(eof)) {
		args = append(args, cp.match(cp.next.t_type))
	}
	if cp.peek(cm) {
		cp.match(cm)
		args_t := cp.parse_A()
		for _, arg := range args_t {
			args = append(args, arg)
		}
	}
	return args
}
