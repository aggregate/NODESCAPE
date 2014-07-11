package config

import (
	"io"
	"os"
	"fmt"
	"strings"
)


type Property interface {
	Measure() string
	Name() string
	Poll() string
	Options() map[string] string
}

type ExternalProperty struct {
	name string
	command string
	poll string
	options map[string] string
}

func (e* ExternalProperty) Name() string {
	return e.name
}

func (e* ExternalProperty) Poll() string {
	return e.poll
}

func (e* ExternalProperty) Options() map[string] string {
	return e.options
}

func (e* ExternalProperty) Measure() string {
/*
	cmd = some_shell
	args := []some_string{cmd, "-c", e.command}
	command := exec.Command(cmd, args)
	stdout, err := command.StdoutPipe()
	stderr, err = command.StderrPipe()
	// Perhaps I should provide buffers for these instead?
	err = command.Run()
*/

	return "external-stub"
}

type BuiltinProperty struct {
	name string
	poll string
	options map[string] string
}

func (b* BuiltinProperty) Measure() string {
	return "internal-stub"
}

func (b* BuiltinProperty) set_name(section sect_node) {
// Use the below just as soon as it exists on all your systems (Go 1.1?).
//	b.Name = strings.TrimPrefix(section.name, "builtin_")
// I've replicated the code here for now.
	if strings.HasPrefix(section.name, "builtin_") {
		b.name = section.name[len("builtin_"):]
	} else {
		b.name = section.name
	}
}

func (b* BuiltinProperty) Name() string {
	return b.name
}

func (b* BuiltinProperty) Poll() string {
	return b.poll
}

func (b* BuiltinProperty) Options() map[string] string {
	return b.options
}

type NSConfig struct {
	Servers []string
	NetUpdate string
	Hostname string
	Group string
	Properties []Property

}

func ReadConfig(in io.Reader) {

	cp := parse_config(in)
	var config NSConfig

	for _, section := range cp.config.sects {
		switch {
			case section.name == "epacsedon":
				config.main_config(section)
			case strings.HasPrefix(section.name, "builtin_"):
				config.builtin_config(section)
				fmt.Fprintf(os.Stderr, "Found builtin property\n")
			default:
				config.external_config(section)
				fmt.Fprintf(os.Stderr, "Found external property\n")
		}
	}

	fmt.Fprintf(os.Stderr, "Server: %s\n", config.Servers[0])
	fmt.Fprintf(os.Stderr, "NetUpdate: %s\n", config.NetUpdate)
	fmt.Fprintf(os.Stderr, "Hostname: %s\n", config.Hostname)
	fmt.Fprintf(os.Stderr, "Group: %s\n", config.Group)

	fmt.Fprintf(os.Stderr, "%d\n", len(config.Properties))
	for _, p := range config.Properties {
		fmt.Fprintf(os.Stderr, "Name: %s\n", p.Name())
		fmt.Fprintf(os.Stderr, "Poll: %s\n", p.Poll())
		for key := range p.Options() {
			fmt.Fprintf(os.Stderr, "%s: %s\n", key, p.Options()[key])
		}
	}

}

func (nsc *NSConfig) main_config(section sect_node) {

	for _, opt := range section.opts {
		switch opt.key {
			case "hostname":
				if len(opt.vals) > 0 {
					nsc.Hostname = opt.vals[0].value
				} else {
					fmt.Fprintf(os.Stderr,
						"Option \"%s\" requires a value", opt.key)
				}
			case "net_update":
				if len(opt.vals) == 2 {
					nsc.NetUpdate = opt.vals[0].value + " " + opt.vals[1].value
				} else if len(opt.vals) == 1 {
					nsc.NetUpdate = opt.vals[0].value
				} else {
					fmt.Fprintf(os.Stderr,
						"Option \"%s\" requires a value", opt.key)
				}
			case "group":
				if len(opt.vals) > 0 {
					nsc.Group = opt.vals[0].value
				} else {
					fmt.Fprintf(os.Stderr,
						"Option \"%s\" requires a value", opt.key)
				}
			case "server":
				if len(opt.vals) > 0 {
					nsc.Servers = append(nsc.Servers, opt.vals[0].value)
				} else {
					fmt.Fprintf(os.Stderr,
						"Option \"%s\" requires a value", opt.key)
				}
			default:
				fmt.Fprintf(os.Stderr, "Unrecognized option: %s", opt.key)
		}
	}
}

func (nsc *NSConfig) builtin_config(section sect_node) {

	var prop BuiltinProperty
	prop.options = make(map[string] string)
	prop.set_name(section)
	for _, opt := range section.opts {
		switch opt.key {
			case "hostname":
				if len(opt.vals) > 0 {
					prop.options["hostname"] = opt.vals[0].value
				}
			case "poll":
				if len(opt.vals) > 0 {
					prop.poll = opt.vals[0].value
				} else {
					fmt.Fprintf(os.Stderr,
						"Option \"%s\" requires a value", opt.key)
				}
			default:
				fmt.Fprintf(os.Stderr, "Unrecognized option: %s\n", opt.key)
		}
	}
	nsc.Properties = append(nsc.Properties, Property(&prop))
}

func (nsc *NSConfig) external_config(section sect_node) {

	var prop ExternalProperty
	prop.name = section.name
	prop.options = make(map[string] string)
	for _, opt := range section.opts {
		switch opt.key {
			case "hostname":
				if len(opt.vals) > 0 {
					prop.options["hostname"] = opt.vals[0].value
				}
			case "poll":
				if len(opt.vals) > 0 {
					prop.poll = opt.vals[0].value
				} else {
					fmt.Fprintf(os.Stderr,
						"Option \"%s\" requires a value", opt.key)
				}
			case "command":
				if len(opt.vals) > 0 {
					prop.command = opt.vals[0].value
				} else {
					fmt.Fprintf(os.Stderr,
						"Option \"%s\" requires a value", opt.key)
				}
			default:
				fmt.Fprintf(os.Stderr, "Unrecognized option: %s", opt.key)
		}
	}
	nsc.Properties = append(nsc.Properties, Property(&prop))
}
