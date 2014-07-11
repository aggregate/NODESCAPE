#!/usr/bin/python3

from nsutil import *
from nsconfig import *

def main():

	if (len(sys.argv) != 4):
		print("Usage: %s <config file> <host> <property>"%(sys.argv[0]))
		sys.exit()
	

	config = read_config_simple(sys.argv[1])	

	host = sys.argv[2]
	prop = sys.argv[3]

	con, curs = mysql_connect(config)

	mydata = get_propdata_hosts((host,), prop, config, curs)

	fout = open(rrd_name(host,prop)+".txt", "w")

	for data, ctime in mydata[host]:
		fout.write(data)
		fout.write("\t")
		fout.write(str(date_to_epoch(ctime)))
		fout.write("\n")

	fout.close()
	curs.close()
	con.close()

main()
