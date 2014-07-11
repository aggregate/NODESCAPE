#!/usr/local/bin/python3

from cgi import FieldStorage

from nsutil import *
from nsconfig import *


from time import *

def main():
	getdata = FieldStorage()
	host = getdata["host"].value.strip()
	prop = getdata["prop"].value.strip()

	# need to deal with landing here w/o any arguments.

	config = read_config_simple("dynamic.conf")

	filepre = rrd_name(host,prop)
	filename = "%s/%s.html"%(config["webdir"],filepre)

	print("Content-type: text/html\r\n\r\n", end="")
	if (not os.path.exists(filename) or 
		file_age(filename) > config["max_age"]):

		propcfg = {}
		propcfg[prop] = (host,)

		con,curs = mysql_connect(config)
		prop_data = get_propdata(propcfg[prop], prop, config, curs)
		update_rrd(host, prop, prop_data[host], config)

		graph(host, prop, days_ago(4), "#FF0000", config, filepre+"4day")
		graph(host, prop, days_ago(1), "#FF0000", config, filepre+"1day")
		graph(host, prop, days_ago(2/24),"#FF0000",config, filepre+"2hour")
		graph(host, prop, days_ago(calc_days(config["graphint"], "second")),
				"#FF0000", config, filepre) 

		cgi_host_prop_html(host, prop, config)

	# now, send the file
	fin = open(filename)
	for line in fin:
		print(line, end="")


main()
