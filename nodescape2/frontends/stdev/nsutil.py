import time
import pymysql
import datetime
import os
import rrdtool

def now():
	return time.clock_gettime(time.CLOCK_REALTIME)

def report(task, start, end, config):
	ns_print("debug", "%s: %0.3f<br>"%(task,end-start), config)

def calc_days(value, unit):
	days = {
			"second": (1/24/60/60),
			"minute": (1/24/60),
			"hour": (1/24),
			"day": (1),
			"week": (7),
			"month": (7*30), # I know, they've got 31 and 29 and 28.
									# I don't care.
			"year": (7*52)
			}
	return (days[unit] * value)



# Convert to minutes from some other unit
def calc_minutes(value, unit):
	minutes = {
			"second": (1/60),
			"minute": (1),
			"hour": (60),
			"day": (60*24),
			"week": (60*24*7),
			"month": (60*24*7*30), # I know, they've got 31 and 29 and 28.
									# I don't care.
			"year": (60*24*7*52)
			}
	return (minutes[unit] * value)

def calc_seconds(value, unit):
	seconds = {
			"second": (1),
			"minute": (60),
			"hour": (60*60),
			"day": (60*60*24),
			"week": (60*60*24*7),
			"month": (60*60*24*7*30), # I know, they've got 31, 29 and 28.
									# I don't care.
			"year": (60*60*24*7*52)
			}
	return (seconds[unit] * value)

# generate the pre-extension part of a file name
def rrd_name(host, stat):
	return (host + "_" + stat).replace(' ','')

def file_age(path):
	mtime = int(os.path.getmtime(path))
	return time.time() - mtime

# Convert to seconds from a date object
def date_to_epoch(adate):
	return int(time.mktime(adate.timetuple()))

def epoch_to_date(seconds):
	return datetime.datetime.fromtimestamp(seconds)	

# Get the epoch numdays days ago
def days_ago(numdays):
	return int(time.time()-(numdays*24*3600))

# Replace all spaces in a string
# Doesn't touch tabs, at least not now
def strip_space(astring):
	return astring.replace(' ','')

def web_strip_space(astring):
	return astring.replace(' ','%20')

# get a mysql connection
def mysql_connect(config):
	con = None
	con = pymysql.connect(host = config["host"],
							user = config["user"],
							passwd = config["passwd"],
							port = config["port"],
							db = config["db"])
	mycursor = con.cursor()
	return (con, mycursor)

def create_rrd(host, property, config):
	rrdtool.create("%s/%s.rrd"%(config["datadir"],rrd_name(host,property)),
					"--step", "600", "--start", "%d" 
						% (days_ago(config["rrd_age"])),
					"DS:%s:GAUGE:1200:0:50000"
						%(strip_space(property)),
					"RRA:AVERAGE:0.5:1:1000")


# Insert data into an round-robin-database
# host: a string, representing a hostname
# property: a string, representing a property being monitored
# data: A list of (record, time) pairs
# config: Our configuration options
def update_rrd(host, property, data, config):
	if (not os.path.exists(rrd_name(host,property)+".rrd")):
		create_rrd(host, property, config)

	for record,ctime in data:
		try:
			rrdtool.update(
					config["datadir"]+"/%s.rrd"%(rrd_name(host,property)),
					"%d:%s"%(date_to_epoch(ctime), record))

		except ValueError:
			ns_print("error",
					"ValueError: %s<br>"%(rrd_name(host,property)),config)
		except rrdtool.OperationalError as oe:
			ns_print("error","rrdtool Operational Error: \n%s:%s\n<br>"
					%(oe,rrd_name(host,property)), config)

	
# Convert a list of lists of properties, indexed by host, into
# a list of lists of hosts, indexed by property
# This does handle both nodescape groups and individual hosts
# reasonably. The host list for a given property may contain some
# redundancies, in that both a host and its group may be listed.
# Care should be taken to make sure that this doesn't cause a problem
# in the way the returned data is used.
def hosts_per_prop(props_per_host, props):
	for host in props_per_host.keys():
		if host[0] == ".": # groups must be inserted differently
			for prop in props_per_host[host]:
				if (prop in props.keys() 
					and "nsgroup "+host not in props[prop]):
					props[prop] += ("nsgroup "+host,)
				else:
					props[prop] = ("nsgroup "+host,)
		else: # dealing with a single host
			for prop in props_per_host[host]:
				if (prop in props.keys() 
					and host not in props[prop]):
					props[prop] += (host,)
				else:
					props[prop] = (host,)

	return props

def get_prop_list(curs, config):
	query_str = "select distinct label from %s"%(config["table"])
	curs.execute(query_str)
	props = []
	for prop in curs:
		props += prop

	return props

def get_host_list(curs, config):
	query_str = "select distinct host from %s"%(config["table"])	
	curs.execute(query_str)
	hosts = []
	for host in curs:
		hosts += host

	return hosts

# Get all data for prop_name sorted by host name, going back
# config["fetchint"] minutes.
def get_propdata_hosts(hosts, prop_name, config, curs):
	# Set fields and table, open parentheses for host specification
	query_str = "select host, data, ctime from "+config["table"]+"  where ("

	add_or = False 
	for host in hosts:
		if add_or:
			query_str += " or "
		else:
			add_or = True 

		if host.split()[0] == "nsgroup":
			# we're dealing with a nodescape group by this name
			# we're going to require that the group name include the leading
			# dot
			query_str += "host like \"%%%s\" " % (host.split()[1])

		else:
			# Just a hostname
			query_str += "host = \"%s\"" % (host)

	# Close parentheses for host spec
	# Specify label that we care about
	query_str += ") and label=\"%s\" " %(prop_name)

	# How far back we should look
	query_str += (" and ctime > date_sub(now(), interval %d minute) "
					% (calc_minutes(config["fetchint"], "second")))

	# Return results starting with the oldest
	query_str += "order by ctime asc;"


	curs.execute(query_str)	

	prop_by_host = {}

	# sort prop by host
	for host, data, ctime in curs:
		if host in prop_by_host.keys():
			prop_by_host[host] += ((data,ctime),)
		else:
			prop_by_host[host] = ((data,ctime),)

	return prop_by_host

def get_host_prop_data(host, prop, config, curs):
	query_str = ("select data,ctime from %s where host=\"%s\""
					%(config["table"],host))

	query_str = (" and label=\"%s\" "%(prop))

	query_str += (" and ctime > date_sub(now(), interval %d minute);"
					%(calc_minutes(config["fetchint"], "second")))
	query_str += "order by ctime asc;"

	curs.execute(query_str)

	data = []
	for record in curs:
		data += record	

	data = sorted(data, key=lambda record: data_to_epoch[record[1]])

	return data

def get_hostdata(host, config, curs):
	query_str = ("select label,data,ctime from %s where host=\"%s\""
					%(config["table"],host))
	query_str += (" and ctime > date_sub(now(), interval %d minute);"
					%(calc_minutes(config["fetchint"], "second")))
	query_str += "order by ctime asc;"

	curs.execute(query_str)

	host_data = {}

	# sort by label
	for label, data, ctime in curs:
		if label in host_data.keys():
			host_data[label] += ((data, ctime),)
		else:
			host_data[label] = ((data, ctime),)

	for label in host_data.keys():		
		host_data[label] = sorted(host_data[label],
								key=lambda record: date_to_epoch(record[1]))

	return host_data			

def get_propdata(prop, config, curs):
	query_str = ("select host,data,ctime from %s where label=\"%s\""
					%(config["table"],prop))
	query_str += (" and ctime > date_sub(now(), interval %d minute) "
					%(calc_minutes(config["fetchint"], "second")))
	query_str += "order by ctime asc;"

	curs.execute(query_str)

	prop_data = {}

	# sort by host 
	for host, data, ctime in curs:
		if host in prop_data.keys():
			prop_data[host] += ((data, ctime),)
		else:
			prop_data[host] = ((data, ctime),)

	for host in prop_data.keys():		
		prop_data[host] = sorted(prop_data[host],
								key=lambda record: date_to_epoch(record[1]))

	return prop_data			


def graph(host, property, start, color, config, png_name):
	rrdtool.graph("%s/%s.png"%(config["webdir"], png_name),
					"--start", "%d"%(start),
					"DEF:%s=%s/%s.rrd:%s:AVERAGE"
						%(strip_space(property), config["datadir"],
							rrd_name(host, property),strip_space(property)),
					"LINE1:%s%s"
						%(strip_space(property), color),
					"--title", "\"%s: %s\"" % (host, property),
					"-v", "\"%s\"" % (property),
					"-l", "0", "-X", "0", "-r")

def gen_html_anomaly(host, property, config):
	fout = open("%s/%s.html"
					%(config["webdir"],rrd_name(host,property)), 'w')
	fout.write("<html>\n<body>\n<table>")
	fout.write("<tr>")
	fout.write("<td><img src=%s.png></td>\n"
				%(rrd_name(host,property)+"2hour"))
	fout.write("<td><img src=%s.png></td>\n"
				%(rrd_name(host,property)+"1day"))
	fout.write("</tr>\n<tr>")
	fout.write("<td><img src=%s.png></td>\n"
				%(rrd_name(host,property)))
	fout.write("<td><img src=%s.png></td>\n"
				%(rrd_name(host,property)+"4day"))
	fout.write("</tr>\n")
	fout.write("</table>\n</body>\n</html>")
	fout.close()

def cgi_host_prop_html(host, property, config):
	fout = open("%s/%s.html"
					%(config["webdir"],rrd_name(host,property)), 'w')
	fout.write("<html>\n")
	fout.write("<meta http-equiv=\"refresh\" content=\"%d\" "
				%(config["refresh"]))
	fout.write("url=\"http://super.ece.engr.uky.edu:8092/ns-bin/")
	fout.write("hostprop.py?host=%s?prop=%s\">"
				%(host,rrd_name(host,property.replace(" ", "%20"))))
	fout.write("\n<body>\n<table>")
	fout.write("<tr>")

	fout.write("<td><img src=%s/%s.png></td>\n"
				%(config["src_webdir"], rrd_name(host,property)))
	fout.write("</tr>\n<tr>")
	fout.write("<td><img src=%s/%s.png></td>\n"
				%(config["src_webdir"], rrd_name(host,property)+"2hour"))
	fout.write("</tr>\n<tr>")
	fout.write("<td><img src=%s/%s.png></td>\n"
				%(config["src_webdir"], rrd_name(host,property)+"1day"))
	fout.write("</tr>\n<tr>")
	fout.write("<td><img src=%s/%s.png></td>\n"
				%(config["src_webdir"], rrd_name(host,property)+"4day"))
	fout.write("</tr>\n")
	fout.write("</table>\n")

	gen_prop_link(property, fout)
	gen_host_link(host, fout)	


	fout.write("</body>\n</html>")
	fout.close()

def cgi_host_html(host, properties, config):
	fout = open("%s/%s.html"
					%(config["webdir"],host), 'w')
	fout.write("<html>\n")
	fout.write("<meta http-equiv=\"refresh\" content=\"%d\" "
				%(config["refresh"]))
	fout.write("url=\"http://super.ece.engr.uky.edu:8092/ns-bin/")
	fout.write("host.py?host=%s\""
				%(host))
	fout.write("\n<body>\n<table>")

	i = 0
	for property in properties:
		if i == 0:
			fout.write("\n<tr>")
		fout.write("<td>")
		fout.write("<a href=\"")
		fout.write("http://super.ece.engr.uky.edu:8092")
		fout.write("/ns-bin/hostprop.py?host=%s&prop=%s"
					%(host, web_strip_space(property)))
		fout.write("\">")
		fout.write("<img src=%s/%s.png>"
					%(config["src_webdir"], rrd_name(host,property)))
		fout.write("</a>")

		gen_prop_link(property, fout)
		gen_host_prop_link(host, property, fout)
		fout.write("<hr>\n")

		fout.write("</td>\n")
		if i == 1:
			fout.write("</tr>\n")
		i = (i + 1) % 2;

	fout.write("</table>\n</body>\n</html>")
	fout.close()

def cgi_prop_html(property, hosts, config):
	fout = open("%s/%s.html"
					%(config["webdir"],property), 'w')
	fout.write("<html>\n")
	fout.write("<meta http-equiv=\"refresh\" content=\"%d\" "
				%(config["refresh"]))
	fout.write("url=\"http://super.ece.engr.uky.edu/8092/ns-bin/")
	fout.write("prop.py?prop=%s\""
				%(property))
	fout.write("\n<body>\n<table>")

	i = 0
	for host in hosts:
		if i == 0:
			fout.write("\n<tr>")

		fout.write("<td>")

		fout.write("<a href=\"")
		fout.write("http://super.ece.engr.uky.edu:8092")
		fout.write("/ns-bin/hostprop.py?host=%s&prop=%s"
					%(host, web_strip_space(property)))
		fout.write("\">")

		fout.write("<img src=%s/%s.png>"
					%(config["src_webdir"], rrd_name(host,property)))

		fout.write("</a>")

		gen_host_link(host, fout)
		gen_host_prop_link(host, property, fout)
		fout.write("<hr>\n")

		fout.write("</td>\n")
		if i == 1:
			fout.write("</tr>\n")
		i = (i + 1) % 2;

	fout.write("</table>\n</body>\n</html>")
	fout.close()

def gen_host_link(host, fout):

	fout.write("<br>\n")
	fout.write("<a href=\"")
	fout.write("http://super.ece.engr.uky.edu:8092")
	fout.write("/ns-bin/host.py?host=%s"%(host))
	fout.write("\">")

	fout.write("All properties for %s"%(host))

	fout.write("</a>")


def gen_prop_link(property, fout):

	fout.write("<br>\n")
	fout.write("<a href=\"")
	fout.write("http://super.ece.engr.uky.edu:8092")
	fout.write("/ns-bin/prop.py?prop=%s"%(web_strip_space(property)))
	fout.write("\">")

	fout.write("%s for all hosts"%(property))

	fout.write("</a>")

def gen_host_prop_link(host, property, fout):

	fout.write("<br>\n")
	fout.write("<a href=\"")
	fout.write("http://super.ece.engr.uky.edu:8092")
	fout.write("/ns-bin/hostprop.py?host=%s&prop=%s"
				%(host,web_strip_space(property)))
	fout.write("\">")

	fout.write("More detail for %s on %s"%(property, host))

	fout.write("</a>")



def ns_print(level, msg, config):
	if level == "normal":
		print(msg)
	elif level == "debug":
		if config["debug"] == "on":
			print("DEBUG:", msg)
	elif level == "error":
		if config["print_err"] == "on" or config["debug"] == "on":
			print("ERROR:", msg)

def needs_regen(filename, config):
	return (not os.path.exists(filename) 
			or file_age(filename) > config["max_age"])


