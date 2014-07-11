#!/usr/local/bin/python3

from nsconfig import *
from nsutil import *

import os 		# using command line arguments; sys.argv
from math import sqrt
from numpy import *

class anomaly_t:
	def __init__(self, init_odd, init_host, init_prop, init_stdev, init_mean):
		self.odd = init_odd
		self.host = init_host
		self.prop = init_prop
		self.max = 0
		self.stdev = init_stdev 
		self.mean = init_mean
		self.sigs = 0

def stdev_analyze_multi(prop_data, prop_name, config):
	anomalies = ()
	for host in prop_data.keys():
		res = stdev_analyze(host, prop_name, prop_data[host], config)
		if res.odd:
			anomalies += (res,)

	return anomalies

def stdev_analyze_interval_multi(prop, prop_data, config):
	anomalies = ()
	for host in prop_data.keys():
		res = stdev_analyze_interval(host, prop, prop_data[host], config)
		if res.odd:
			anomalies += (res,)

	return anomalies

def stdev_analyze_interval(host, prop_name, data, config):
#	mean_prev = float(data[0][0])
#	mean = 0
#	variance_prev = 0
#	variance = 0
#	k = 1
#	last_update = epoch_to_date(0)
#
#	for record in data:
#		try:
#			record = float(record[0])
#			mean = mean_prev + (record - mean_prev)/(k)
#			variance = variance_prev + (record - mean_prev)*(record - mean)
#			mean_prev = mean
#			variance_prev = variance
#			k = k + 1
#		except ValueError as ve:
#			ns_print("error", "Value Error: \n%s:%s\n"
#						%(ve, rrd_name(host,prop)), config)
#
#	stdev = sqrt(variance)

	# Let's try this with numpy
	# first, build the array we want stdev of:
	prop_data = array(list(float(record[0]) for record in data))

	stdev = prop_data.std()
	mean = prop_data.mean()

	rtv = anomaly_t(False, host, prop_name, stdev, mean)
	bound = (config["bound"] * stdev)
	# We don't limit our search for anomalies to new data. We want this
	# to stay flagged, even if we've seen the anomaly before.
	for record in prop_data:
		if (bound < abs(record - mean)):
			rtv.odd = True
			if (rtv.max < abs(record - mean)):
				rtv.max = record 
				rtv.sigs = abs(record - mean) / stdev

	return rtv


def stdev_analyze(host, prop_name, data, config):
	mean_prev = float(data[0][0])
	variance_prev = 0
	k = 1
	last_update = epoch_to_date(0)

	if (os.path.exists("%s/%s.stdev"
						%(config["datadir"],rrd_name(host,prop_name)))):

		fin = open("%s/%s.stdev" 
					% (config["datadir"],rrd_name(host,prop_name)), 'r')

		mean_prev, std_prev, k, last_update = fin.readline().split()
		mean_prev = float(mean_prev)
		variance_prev = float(variance_prev)
		k = int(k) + 1
		last_update = epoch_to_date(int(last_update))

	# We only want to add new data to our calculation of variance
	i = 0
	while i < len(data) and data[i][1] <= last_update:
		i = i + 1

	mean = mean_prev
	variance = variance_prev
	for record in data[i:]:
		try:
			record = float(record[0])
			mean = mean_prev + (record - mean_prev)/(k)
			variance = variance_prev + (record - mean_prev)*(record - mean)
			mean_prev = mean
			variance_prev = variance
			k = k + 1
		except ValueError as ve:
			ns_print("error","Value Error: \n%s:%s\n"
				%(ve,rrd_name(host,prop_name)), config)

	fout = open("%s/%s.stdev"
					%(config["datadir"],rrd_name(host,prop_name)), 'w')
	fout.write("%f %f %d %d\n" 
				% (mean,variance, k, date_to_epoch(data[-1][1])))

	rtv = anomaly_t(False, host, prop_name)
	# We don't limit our search for anomalies to new data. We want this
	# to stay flagged, even if we've seen the anomaly before.
	for record in data:
		if (float(record[0]) >= config["bound"] * sqrt(variance)):
			rtv = anomaly_t(True, host, prop_name)

	return rtv
