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

	#config = read_config_simple("dynamic.conf")
	config = read_config_simple("test.conf")

	filepre = rrd_name(host,prop)
	html_filename = "%s/%s.html"%(config["webdir"],filepre)
	graph4d_filename = "%s/%s4day.png"%(config["webdir"],filepre)
	graph1d_filename = "%s/%s1day.png"%(config["webdir"],filepre)
	graph2h_filename = "%s/%s2hour.png"%(config["webdir"],filepre)
	graphfi_filename = "%s/%s.png"%(config["webdir"],filepre)

	print("Content-type: text/html\r\n\r\n", end="")
	# Only do a db query if something is old.
	if (needs_regen(graph4d_filename, config)
		or needs_regen(graph1d_filename, config)
		or needs_regen(graph2h_filename, config)
		or needs_regen(graphfi_filename, config)
		):

		con,curs = mysql_connect(config)
		prop_data = get_propdata_hosts((host,), prop, config, curs)
		update_rrd(host, prop, prop_data[host], config)

		# Only generate a new graph if we've updated the RRD and the
		# graph is old.
		if (needs_regen(graph4d_filename, config)):
			graph(host, prop, days_ago(4),"#FF0000", config, filepre+"4day")
		if (needs_regen(graph1d_filename, config)):
			graph(host, prop, days_ago(1),"#FF0000", config, filepre+"1day")
		if (needs_regen(graph2h_filename, config)):
			graph(host,prop,days_ago(2/24),"#FF0000",config,filepre+"2hour")
		if (needs_regen(graphfi_filename, config)):
			graph(host,prop,
				days_ago(calc_days(config["graphint"],"second")), 
				"#FF0000", config, filepre) 

	# Regenerate the html if it is old. Generally, the html won't be 
	# changing.
	if (needs_regen(html_filename, config)):
		cgi_host_prop_html(host, prop, config)

	# now, send the file
	fin = open(html_filename)
	for line in fin:
		print(line, end="")


main()
