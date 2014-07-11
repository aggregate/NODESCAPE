#!/usr/local/bin/python3

from time import clock_gettime, CLOCK_REALTIME
from nsutil import *
from nsconfig import *
from stdev import stdev_analyze_interval_multi

def main():
	read_start = now() 

	config = read_config_simple(sys.argv[1])

	report("Read config", read_start, now(), config)

	con, curs = mysql_connect(config)

	props = get_prop_list(curs, config)

	stdev_anomalies = ()

	query_total = 0
	graph_total = 0
	prop_data = {}
	for prop in props:
		temp = now()
		prop_data_tmp = get_propdata(prop, config, curs)
		prop_data = dict(list(prop_data.items())+
						list(prop_data_tmp.items()))
		query_total += now() - temp

		stdev_anomalies += (
				stdev_analyze_interval_multi(prop, prop_data_tmp, config))

	stdev_anomalies = sorted(stdev_anomalies, key=lambda anom: anom.sigs)

	frontpage = open(config["webdir"]+"/"+config["landing"], "w")

	frontpage.write("<html>\n<title>Nodescape</title>\n")
	frontpage.write("<meta http-equiv=\"refresh\" content\"%d\" "
						%(config["refresh"]))
	frontpage.write("url=\"http://super.ece.engr.uky.edu:8092/%s/%s\">"
						%(config["src_webdir"],config["landing"]))
	frontpage.write("<body>\n")
	frontpage.write("<table>\n")

	i = 0
	for anomaly in stdev_anomalies:
		if anomaly.odd:
			host = anomaly.host
			prop = anomaly.prop

			ns_print("debug", "Anomaly:", config)
			ns_print("debug", "Host: %s"%anomaly.host, config)
			ns_print("debug", "Property: %s"%anomaly.prop, config)
			ns_print("debug", "Mean: %.3f"%anomaly.mean, config)
			ns_print("debug", "Stdev: %.3f"%anomaly.stdev, config)
			ns_print("debug", "Most odd value: %.3f"%anomaly.max, config)
			ns_print("debug", "Sigmas from mean: %.3f"%anomaly.sigs, config)
			ns_print("debug", "-------------------------------------", config)

			update_rrd(host, prop, prop_data[host], config)

			temp = now()
			graph(host, prop, days_ago(4), "#FF0000", config,
					rrd_name(host, prop)+"4day")
			graph(host, prop, days_ago(1), "#FF0000", config,
					rrd_name(host, prop)+"1day")
			graph(host, prop, days_ago(2/24), "#FF0000", config,
					rrd_name(host, prop)+"2hour")
			graph(host, prop,days_ago(calc_days(config["graphint"], "second")), 
					"#FF0000", config, rrd_name(host, prop))

			graph_total += now() - temp

			gen_html_anomaly(host, prop, config)

			if i % 2 == 0:
				frontpage.write("\t<tr>\n")

			frontpage.write("\t\t<td>\n")
			frontpage.write("\t\t\t<a href=\"")
			frontpage.write("http://super.ece.engr.uky.edu:8092")
			frontpage.write("/ns-bin/hostprop.py?host=%s&prop=%s"
								%(host, web_strip_space(prop)))
			frontpage.write("\">\n")
			frontpage.write("\t\t\t<img src=%s/%s.png>\n"
							%(config["src_webdir"], rrd_name(host,prop)))
			frontpage.write("\t\t\t</a>\n")

			frontpage.write("<br>\n")
			frontpage.write("Mean: %.3f<br>\n"%anomaly.mean)
			frontpage.write("Stdev: %.3f<br>\n"%anomaly.stdev)
			frontpage.write("Most out of range value: %.3f<br>\n"%anomaly.max)
			frontpage.write("Sigmas from mean: %.3f<br>\n"%anomaly.sigs)


			gen_host_link(host, frontpage)
			gen_prop_link(prop, frontpage)
			gen_host_prop_link(host, prop, frontpage)

			frontpage.write("\t\t\t<hr>\n")

			frontpage.write("\t\t</td>\n")

			if i % 2 == 1:
				frontpage.write("\t</tr>\n")
			i += 1

	frontpage.write("</table>\n</body>\n</html>")	
	frontpage.close()

	stop = clock_gettime(CLOCK_REALTIME)

	ns_print("debug","Number of anomalies: %d"%len(stdev_anomalies),config)

	ns_print("debug","Query time: %.3fs\n"%(query_total), config)
	ns_print("debug","Graphing time: %.3fs\n"%(graph_total), config)

	report("Total program time", read_start, now(), config)

main()
