#!/usr/local/bin/python3

import pymysql
from subprocess import *
import cgi

from nsutil import *

config = {}

config["host"] = "super.ece.engr.uky.edu"
config["user"] = "nsfront"
config["passwd"] = "frontpass"
config["port"] = 13092
config["db"] = "nodescape"
config["webdir"] = "/var/nodescape/html/"
config["datadir"] = "/var/nodescape/data/"

def print_list_select(listname):
	print("<option value=\"none\" selected=\"selected\">no selection</option>")
	lines= open(config["datadir"]+"/"+listname, "r")
	for line in lines:
		print("<option value=\"%s\">%s</option>"%(line,line))

def send_form():
	print("<html>\n<body>")
	print("<center><h1>Custom Graph Creation</h1></center>")
	print("<br>")

	print("<form action=\"/ns-bin/custom.py\" method=\"get\">")

	print("Graph history period:")
	print("<input type=\"text\" name=\"history\" value=\"7\">")

	print("<select name=\"hist_units\">")
	print("\t<option value=\"minute\">minutes</option>")
	print("\t<option value=\"hour\">hours</option>")
	print("\t<option value=\"day\" selected=\"selected\">days</option>")
	print("\t<option value=\"week\">weeks</option>")
	print("</select>")
	print("<br>")

	print("Host 1: <select name=\"host1\">")
	print_list_select("hostlist")
	print("</select>")
	print("<br>")

	print("Host 2: <select name=\"host2\">")
	print_list_select("hostlist")
	print("</select>")
	print("<br>")

	print("Host 3: <select name=\"host3\">")
	print_list_select("hostlist")
	print("</select>")
	print("<br>")

	print("Host 4: <select name=\"host4\">")
	print_list_select("hostlist")
	print("</select>")
	print("<br>")

	print("Host 5: <select name=\"host5\">")
	print_list_select("hostlist")
	print("</select>")
	print("<br>")

	print("Stat to graph: <select name=\"stat\">")
	print_list_select("statlist")
	print("</select>")
	print("<br>")

	print("<input type=\"submit\" name=\"submit\">")

	print("</form>")
	print("</body>\n</html>")


def main():
	getdata = cgi.FieldStorage()
	if (len(getdata) == 0):
		# return the form
		print("Content-type: text/html\r\n\r\n", end="")
		send_form()

	else:
		con = None
		con = pymysql.connect(host = config["host"],
								user = config["user"],
								passwd = config["passwd"],
								port = config["port"],
								db = config["db"])
		curs = con.cursor()
		host1 = getdata["host1"].value.strip()
		host2 = getdata["host2"].value.strip()
		host3 = getdata["host3"].value.strip()
		host4 = getdata["host4"].value.strip()
		host5 = getdata["host5"].value.strip()
		stat = getdata["stat"].value.strip()
		history = int(getdata["history"].value)
		hist_units = getdata["hist_units"].value.strip()

		hosts = ()

		if (host1 != "none"):
			hosts += (host1,)
		if (host2 != "none"):
			hosts += (host2,)
		if (host3 != "none"):
			hosts += (host3,)
		if (host4 != "none"):
			hosts += (host4,)
		if (host5 != "none"):
			hosts += (host5,)

		print("Content-type: text/html\r\n\r\n", end="")
	
		stat_data = {}
		for host in hosts:	
			query = (("select host, data, ctime from ukanstats where "+
					("label = \"%s\""%(stat))+
					(" and host=\"%s\" "+
					"and ctime > date_sub(now(), interval %d minute) "+
					"order by ctime asc;")
					%(host, calc_minutes(history,hist_units))))

			curs.execute(query)

			for host,data,ctime in curs:
				if host in stat_data.keys():
					stat_data[host] += ((data,ctime),)
				else:
					stat_data[host] = ((data,ctime),)


			create_cmd = (("rrdtool create /var/nodescape/data/%s.rrd"+
							" --start %d --step 600"+ 
							" DS:%s:GAUGE:1200:0:50000"+
							" RRA:AVERAGE:0.5:1:1500 > /dev/null")
							%((rrd_name(host,stat)+"_custom"),
								date_to_sec(stat_data[host][0][1])-1, 
								strip_space(stat)
							)
						)
			call(create_cmd, shell=True)


			for data,ctime in stat_data[host]:
				update_cmd = ("rrdtool update "+
								"/var/nodescape/data/%s %d:%s > /dev/null"
								%((rrd_name(host,stat)+"_custom.rrd "),
									date_to_sec(ctime),
									data))
				call(update_cmd, shell=True)
		# end of for host in hosts

		graph_cmd = (
			("rrdtool graph --start %d "
				%(days_ago(calc_minutes(history,hist_units)/24/60)))+
			("/var/nodescape/html/custom.png "))

		ctr = 0
		color = ["FFF000", "CCC333", "999666", "666999", "333CCC"]
		for host in hosts:
			graph_cmd += (("DEF:%s%d=/var/nodescape/data/%s.rrd:%s:AVERAGE "
							%(strip_space(stat),ctr,
								rrd_name(host,stat)+"_custom",
								strip_space(stat)))+
						("LINE%d:%s%d#%s "
							%(ctr,strip_space(stat),ctr,color[ctr])))
			ctr = ctr + 1


		graph_cmd +=(("--title \"%s: %s\" -v \"%s\""
					%(str(hosts), stat, stat))+
					(" -l 0 -X 0 -r ")+("> /dev/null"))

		call(graph_cmd, shell=True)


		print("<br>")
		print("<img src=\"../nodescape/custom.png\">")
	
main()	
