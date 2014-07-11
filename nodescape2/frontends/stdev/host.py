#!/usr/local/bin/python3

from cgi import FieldStorage

from nsutil import *
from nsconfig import *


from time import *
def main():
	getdata = FieldStorage()
	host = getdata["host"].value.strip()

	# need to deal with landing here w/o any arguments.

#	config = read_config_simple("dynamic.conf")
	config = read_config_simple("test.conf")

	html_filename = "%s/%s.html"%(config["webdir"],host)

	print("Content-type: text/html\r\n\r\n", end="")
	# Only do a db query if something is old.
	# We don't know which graphs we have for a given host, so we'll look
	# at when we last regenerated the html.
	if (needs_regen(html_filename, config)):
		con,curs = mysql_connect(config)

		query_start = now()
		host_data = get_hostdata(host, config, curs)
		report("query", query_start, now(), config)

		graph_start = now()
		for prop in host_data.keys():
			graphfi_filename = "%s/%s.png"%(config["webdir"],
											rrd_name(host,prop))
			update_rrd(host, prop, host_data[prop], config)

			# Only generate a new graph if we've updated the RRD and the
			# graph is old.
			if (needs_regen(graphfi_filename, config)):
				graph(host,prop,
					days_ago(calc_days(config["graphint"],"second")), 
					"#FF0000", config, rrd_name(host, prop)) 

		report("graph", graph_start, now(), config)

		cgi_host_html(host, host_data.keys(), config)

	# now, send the file
	fin = open(html_filename)
	for line in fin:
		print(line, end="")


main()
