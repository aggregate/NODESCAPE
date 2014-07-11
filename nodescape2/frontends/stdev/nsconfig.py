#!/usr/local/bin/python3
import sys
from nsutil import *

int_configs = ("port", "indexhist", "rrd_age")
float_configs = ("bound",)

def set_defaults():
	config = {
			"graphint": (calc_seconds(7, "day")),
			"fetchint": (calc_seconds(15, "minute")),
			"max_age": (calc_seconds(5, "minute")),
			"host": (""),
			"user": ("nsfront"),
			"passwd": (""),
			"port": (13000),
			"db": ("nodescape"),
			"table": ("nsstats"),
			"webdir": ("/var/nodescape/"),
			"datadir": ("./data"),
			"bound": (5),
			"rrd_age": (3),
			"debug": ("off"),
			"print_err": ("on"),
			"refresh": ("5 minute"),
			"landing": "index.html"
			}
	return config

# Build a configuration from a file
# The file should contain one or more of the sections:
#	config
#	groups
#	hosts
# These last two are how we configure a nodescape frontend to know about
# a host and which properties we're monitoring on it.

def read_config(filename):
	config = set_defaults()
	groupcfg = {}
	hostcfg = {}
	propcfg = {}

	fin = open(filename, 'r')
	confstring = ""
	# Read the entire file into one string
	for line in fin:
		confstring += line

	# Parse one section at a time
	for section in confstring.split("endsection"):
		sname, confdata = section.split(':')
		sname = sname.strip()

		if (sname == "config"):
			# The config section is line-ending delimited
			confdata = confdata.strip()
			for option in confdata.split('\n'):
				if (option != "\n"):
					s_option = option.split()
					name = s_option[0].strip()
					value = (s_option[1].strip())
					if (len(s_option) == 3):
						# this is to take care of fetchint and graphint
						unit = s_option[2].strip()
						config[name] = calc_seconds(int(value), unit)
					else:
						if name in int_configs:
							config[name] = int(value)
						elif name in float_configs:
							config[name] = float(value)
						else:
							config[name] = value

		elif (sname == "groups"):
			# The groups section is semicolon delimited
			for group in confdata.split(";"):
				stats = group.split("||");
				name = stats[0].strip()
				stats = stats[1:]
				for stat in stats:
					if not (name in groupcfg.keys()):
						groupcfg[name] = (stat.strip(),)
					else:
						groupcfg[name] += (stat.strip(),)
				
		elif (sname == "hosts"):
			# The hosts section is semicolon delimited
			for host in confdata.split(";"):
				stats = host.split("||");
				name = stats[0].strip()
				stats = stats[1:]
				for stat in stats:
					if not (name in hostcfg.keys()):
						hostcfg[name] = (stat.strip(),)
					else:
						hostcfg[name] += (stat.strip(),)

		elif (sname == "properties"):
			for property in confdata.split(";"):
				hosts = property.split("||");
				name = hosts[0].strip()
				hosts = hosts[1:]
				for host in hosts:
					if (name in propcfg.keys()):
						propcfg[name] += (host.strip(),)
					else:
						propcfg[name] = (host.strip(),)

	return (config, groupcfg, hostcfg, propcfg)


def read_config_simple(filename):
	config = set_defaults()

	fin = open(filename, 'r')
	confstring = ""
	# Read the entire file into one string
	for line in fin:
		confstring += line

	for section in confstring.split("endsection"):
		if (len(section.split(':')) > 1):
			sname, confdata = section.split(':')
			sname = sname.strip()

			if (sname == "config"):
				# The config section is line-ending delimited
				confdata = confdata.strip()
				for option in confdata.split('\n'):
					if (option != "\n"):
						s_option = option.split()
						name = s_option[0].strip()
						value = (s_option[1].strip())
						if (len(s_option) == 3):
							# this is to take care of fetchint and graphint
							unit = s_option[2].strip()
							config[name] = calc_seconds(int(value), unit)
						else:
							if name in int_configs:
								config[name] = int(value)
							elif name in float_configs:
								config[name] = float(value)
							else:
								config[name] = value

	return config
		
