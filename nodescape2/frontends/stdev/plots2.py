#!/usr/local/bin/python3

from time import clock_gettime, CLOCK_REALTIME
from nsutil import *
from nsconfig import *
from stdev import stdev_analyze,stdev_analyze_multi

def main():
	start = clock_gettime(CLOCK_REALTIME)
	start_all = start

	config,groupcfg,hostcfg,propcfg = read_config(sys.argv[1])

	stop = clock_gettime(CLOCK_REALTIME)
	ns_print("debug", "Read configuration: %.3fs\n"%(stop-start), config)

	start = clock_gettime(CLOCK_REALTIME)

	propcfg = hosts_per_prop(groupcfg, propcfg)
	propcfg = hosts_per_prop(hostcfg, propcfg)

	stop = clock_gettime(CLOCK_REALTIME)
	ns_print("debug",
			"Sorted config by property: %.3fs\n"%(stop-start),config)

	con, curs = mysql_connect(config)

	stdev_anomalies = ()

	for property in propcfg.keys():
		start = clock_gettime(CLOCK_REALTIME)

		prop_data = get_propdata(propcfg[property], property, config, curs)

		stop = clock_gettime(CLOCK_REALTIME)
		ns_print("debug","Fetched %s: %.3fs\n"%(property,stop-start),config)

		start = clock_gettime(CLOCK_REALTIME)

		update_start = 0
		update_time = 0
		graph_start = 0
		graph_time = 0

		for host in prop_data.keys():
			update_start = clock_gettime(CLOCK_REALTIME)
			update_rrd(host, property, prop_data[host], config)
			update_time += clock_gettime(CLOCK_REALTIME) - update_start

			graph_start = clock_gettime(CLOCK_REALTIME)
			graph(host,property,days_ago(config["rrd_age"]),
					"#00FF00",config, rrd_name(host, property))
			graph_time += clock_gettime(CLOCK_REALTIME) - update_start

		stop = clock_gettime(CLOCK_REALTIME)
		ns_print("debug","update and graph:: %.3fs\n"%(stop-start),config)
		ns_print("debug","\tupdate time: %.3fs\n"%(update_time), config)
		ns_print("debug","\tgraph time: %.3fs\n"%(graph_time), config)

		stdev_anomalies += (
						stdev_analyze_multi(prop_data, property, config))

	for anomaly in stdev_anomalies:
		graph(anomaly[1], anomaly[2], days_ago(4), "#FF0000", config,
				rrd_name(anomaly[1], anomaly[2])+"4day")
		graph(anomaly[1], anomaly[2], days_ago(1), "#FF0000", config,
				rrd_name(anomaly[1], anomaly[2])+"1day")
		graph(anomaly[1], anomaly[2], days_ago(2/24), "#FF0000", config,
				rrd_name(anomaly[1], anomaly[2])+"2hour")
		gen_html_anomaly(anomaly[1], anomaly[2], config)

	stop = clock_gettime(CLOCK_REALTIME)
	ns_print("debug","Total program time: %.3fs\n"%(stop-start_all), config)	

main()
