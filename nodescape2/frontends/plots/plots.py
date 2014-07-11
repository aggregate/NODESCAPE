#!/usr/local/bin/python3

import pymysql
import time
import os
import rrdtool
from nsconfig import *
from nsutil import *

from subprocess import *
from multiprocessing import *

mydata = {}

def strip_space(astring):
	return astring.replace(' ','')

def rrd_name(host, stat):
	return (host + "_" + stat).replace(' ','')

def date_to_sec(adate):
	return int(time.mktime(adate.timetuple()))

def days_ago(numdays):
	return int(time.time()-(numdays*24*3600))

def process_rrd(args):
	global mydata
	#args = ((host, stat, config),)
	host, stat, config = args
	if (not os.path.exists(rrd_name(host,stat)+".rrd")):
		#create_cmd =(
		#("rrdtool create %s.rrd --start %d --step 600"+
		#" DS:%s:GAUGE:1200:0:50000 RRA:AVERAGE:0.5:1:1000")
		#%(rrd_name(host,stat),date_to_sec(mydata[host][0][1])-1,
		#	strip_space(stat)))
		rrdtool.create("%s.rrd"%(rrd_name(host,stat)),
						"--step", "600",
						"--start","%d"
							%(date_to_sec(mydata[host][0][1])-1),
						"DS:%s:GAUGE:1200:0:50000"
							%(strip_space(stat)),
						"RRA:AVERAGE:0.5:1:1000")
	
		#call(create_cmd,shell=True)
	# Populate the rrd
	total = 0
	ctr = 0
	for data,ctime in mydata[host]:
		try:
			#update_cmd = ("rrdtool update %s %d:%s"
			#	%((rrd_name(host,stat)+".rrd"),
			#		date_to_sec(ctime),
			#		data))
			total = total + float(data)
			ctr = ctr + 1
			rrdtool.update("%s.rrd"%(rrd_name(host,stat)),
							"%d:%s"%(date_to_sec(ctime),data))
		except ValueError:
			print("ValueError: %s"%(rrd_name(host,stat)))
		except rrdtool.OperationalError:
			print("rrdtool error")
		#print(update_cmd)
		#call(update_cmd, shell=True)

	data_average = total / ctr

	averages = []
	total = 0
	alarm = ()
	if (os.path.exists("%s.dat"%(rrd_name(host,stat)))):
		avgfile = open("%s.dat"%(rrd_name(host,stat)), "r")
		for line in avgfile:
			averages.append(float(line))
			total = total + float(line)

		old_average = total / len(averages)

		if (data_average > 0):
			a_ratio = old_average / data_average
		else:
			a_ratio = 1

		if (a_ratio < 0.5 or a_ratio > 1.5):
			alarm = (host, stat, a_ratio)

	averages.append(data_average)
	averages = averages[-config["indexhist"]:]

	avgfile = open("%s.dat"%(rrd_name(host,stat)), "w")	
	for avg in averages:
		avgfile.write("%f\n"%(avg))

	# graph the data
#	graph_cmd = (
#		("rrdtool graph --start %d "
#			%(days_ago(config["graphint"]/24/60)))+
#		("%s.png DEF:%s=%s.rrd:%s:AVERAGE "
#			%(rrd_name(host,stat),strip_space(stat),
#				rrd_name(host,stat),strip_space(stat)))+
#		("LINE2:%s#FF0000 "
#			%(strip_space(stat)))+
#		("--title \"%s: %s\" "
#			%(host, stat))+
#		(" -v \"%s\""
#			%(stat))+
#		(" -l 0 -r -X 0")
#		)
	#print(graph_cmd)
	#call(graph_cmd, shell=True)
	rrdtool.graph("%s.png"%(rrd_name(host,stat)),
				"--start", "%d"
					%(days_ago(config["graphint"]/24/60)),
				"DEF:%s=%s.rrd:%s:AVERAGE"
					%(strip_space(stat),rrd_name(host,stat),
						strip_space(stat)),
				"LINE2:%s#FF0000"
					%(strip_space(stat)),
				"--title", "\"%s: %s\""
					%(host, stat),
				"-v", "\"%s\""
					%(stat),
				"-l", "0", "-X", "0", "-r") 

	return alarm

# set up some global state
def mysql_connect(config):
	con = None
	con = pymysql.connect(host = config["host"],
							user = config["user"],
							passwd = config["passwd"],
							port = config["port"],
							db = config["db"])
	curs = con.cursor()
	return (con,curs)

def gengroups(curs, groupcfg, config, lists):
	hostlist,statlist,alarmlist = lists
	global mydata
	for gname, gcfg in groupcfg.items():
		for stat in gcfg:
			mydata = {}
			query = (("select host, data, ctime from ukanstats where "+
				"label = \"%s\" and host like \"%%%s\" "+
				"and ctime > date_sub(now(), interval %d minute) "+
				"order by ctime asc;")
				%(stat, gname, config["fetchint"]))
			#print(query)
			# query database for: 
			#	host, data, ctime; based on label and group
			curs.execute(query)

			# Build dictionary, using host as key,
			#	value is a tuple containing all of the data from the db
			for host,data,ctime in curs:
				if host in mydata.keys():
					mydata[host] += ((data,ctime),)
				else:
					mydata[host] = ((data,ctime),)

			args = ()
			for host in mydata.keys():
				if (len(args) == 0):
					args = ((host, stat, config),)
				else:
					args += ((host, stat, config),)
				if not (stat in statlist):
					statlist[stat] = (host, )
				else:
					statlist[stat] += (host, )

			for arg in args:
				if not (arg[0] in hostlist):
					hostlist[arg[0]] = (arg[1],)
				else:
					hostlist[arg[0]] += (arg[1],)
			if (len(mydata.keys()) >= 1):
				# create rrd for each host for this stat
				workers = Pool(len(mydata.keys()))
				try:
					alarms = workers.map(process_rrd, args)
					workers.close()
					workers.join()
					print(alarms)
					for alarm in alarms:
						if (alarm != ()):
							alarmlist += (alarm,)
					print(alarmlist)
				except ZeroDivisionError:
					print("ZeroDivisionError")
					workers.close()
					workers.join()
				except ValueError:
					print("ValueError")
					workers.close()
					workers.join()

	return (hostlist,statlist,alarmlist)

def genhosts(curs, hostcfg, config, lists):
	global mydata
	hostlist,statlist,alarmlist = lists
	for hname, hcfg in hostcfg.items():
		for stat in hcfg:
			mydata = {}
			query = (("select host, data, ctime from ukanstats where "+
				"label = \"%s\" and host = \"%s\" "+
				"and ctime > date_sub(now(), interval %d minute) "+
				"order by ctime asc;")
				%(stat, hname, config["fetchint"]))
			#print(query)
			# query database for: 
			#	host, data, ctime; based on label and group
			curs.execute(query)

			# Build dictionary, using host as key,
			#	value is a tuple containing all of the data from the db
			for host,data,ctime in curs:
				if host in mydata.keys():
					mydata[host] += ((data,ctime),)
				else:
					mydata[host] = ((data,ctime),)
			alarm = process_rrd((host, stat, config))

			if (alarm != ()):
				alarmlist += (alarm,)

			if not (stat in statlist):
				statlist[stat] = (hname,)
			else:
				statlist[stat] += (hname,)

			if not (hname in hostlist):
				hostlist[hname] = (stat,)
			else:
				hostlist[hname] += (stat,)

	return (hostlist, statlist, alarmlist)	


def genhost_html(hostlist):
	for host in hostlist.keys():
		fout = open("%s.html"%host, 'w')
		fout.write("<html>\n<body>\n<table>")

		ctr = 0
		for stat in hostlist[host]:
			if (ctr == 0):
				fout.write("<tr>")		
			fout.write("<td><a href=\"%s.html\"><img src=%s.png></a></td>\n"
						%(strip_space(stat), rrd_name(host,stat)))
			if (ctr == 1):
				fout.write("</tr>")
			ctr = (ctr + 1) % 2

		fout.write("</table>\n</body>\n</html>")

def genstat_html(statlist):
	for stat in statlist.keys():
		fout = open("%s.html"%strip_space(stat), 'w')
		fout.write("<html>\n<body>\n<table>\n")

		ctr = 0
		for host in statlist[stat]:
			if (ctr == 0):
				fout.write("<tr>")
			fout.write("<td><a href=\"%s.html\"><img src=%s.png></a></td>\n"
						%(host, rrd_name(host,stat)))
			if (ctr == 1):
				fout.write("</tr>")
			ctr = (ctr + 1) % 2

		fout.write("</table>\n</body>\n</html>\n")

def genlanding_html(alarmlist):

	alarmdict = {}

	for host,stat,ratio in alarmlist:
		alarmdict["%f%s%s"%(ratio,host,stat)] = (host,stat,ratio)

	fout = open("index.html", "w")
	fout.write("<html>\n<body>\n<table>\n")

	ctr = 0
	for key in sorted(alarmdict.keys())[-10:]:
		host,stat,alarm = alarmdict[key]
		if (ctr == 0):
			fout.write("<tr>")

		fout.write("<td><a href=\"%s.html\"><img src=%s.png></a></td>\n"
					%(host, rrd_name(host,stat)))
		if (ctr == 1):
				fout.write("<tr>")
		ctr = (ctr + 1) %2

	fout.write("</table>\n</body>\n</html>")

def copy_files(config):	
	copy_cmd = ("cp *.html *.png %s/."%(config["webdir"]))
	#print(copy_cmd)
	call(copy_cmd, shell=True)

def main():
	config,groupcfg,hostcfg = read_config(sys.argv[1])

	con,curs = mysql_connect(config)

	hostlist = {}
	statlist = {}
	alarmlist = ()
	lists = (hostlist,statlist,alarmlist)

	lists = genhosts(curs, hostcfg, config, lists)
	hostlist,statlist,alarmlist = gengroups(curs, groupcfg, config, lists)
	print("Generated pngs")
	genhost_html(hostlist)
	genstat_html(statlist)
	genlanding_html(alarmlist)

	split_hosts = {}
	for host in sorted(hostlist.keys()):
		split_hosts[host.split('.')[1]+host.split()[0]] = host

	hostfile = open(config["datadir"]+"/hostlist", "w")
	for host in sorted(split_hosts.keys()):
		hostfile.write(split_hosts[host]+"\n")	

	statfile = open(config["datadir"]+"/statlist", "w")
	for stat in sorted(statlist.keys()):
		statfile.write(stat+"\n")	

	copy_files(config)

main()
