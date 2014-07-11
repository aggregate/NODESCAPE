import time

def calc_minutes(value, unit):
	minutes = {
			"minute": (1),
			"hour": (60),
			"day": (60*24),
			"week": (60*24*7),
			"month": (60*24*7*30), # I know, they've got 31 and 29 and 28.
									# I don't care.
			"year": (60*24*7*52)
			}
	return (minutes[unit] * value)

def rrd_name(host, stat):
	return (host + "_" + stat).replace(' ','')

def date_to_sec(adate):
	return int(time.mktime(adate.timetuple()))

def days_ago(numdays):
	return int(time.time()-(numdays*24*3600))

def strip_space(astring):
	return astring.replace(' ','')
