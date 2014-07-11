#!/usr/local/bin/python3
import sys
from nsutil import *

int_configs = ("port", "indexhist")

def set_defaults():
	config = {
			"graphint": (calc_minutes(7, "day")),
			"fetchint": (15),
			"indexhist": (50),
			"host": (""),
			"user": ("nsfront"),
			"passwd": (""),
			"port": (13000),
			"db": ("nodescape"),
			"webdir": ("/var/nodescape/")
			}
	return config

def read_config(filename):
	config = set_defaults()
	groupcfg = {}
	hostcfg = {}

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
						config[name] = calc_minutes(int(value), unit)
					else:
						if name in int_configs:
							config[name] = int(value)
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

	#print(config, groupcfg, hostcfg, sep="\n")
	return (config, groupcfg, hostcfg)


#read_config(sys.argv[1])
		
