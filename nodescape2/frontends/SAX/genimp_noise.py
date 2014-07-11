#!/usr/bin/python3

import sys
import numpy
from random import randrange

if len(sys.argv) < 5:
	print("usage gendata2.py <num samples> <stdev> <pulse width> <output name>")
	sys.exit(0)


num_samples = int(sys.argv[1])
stdev = float(sys.argv[2])
pulse_width = int(sys.argv[3])
out = open(sys.argv[4], "w")


for i in range(int((num_samples-pulse_width)/2)):
	out.write(str(50+numpy.random.normal(0,stdev))+"\t"+str(i)+"\n")

for i in range(int((num_samples-pulse_width)/2),int((num_samples-pulse_width)/2) + pulse_width):
	out.write(str(100+numpy.random.normal(0,stdev))+"\t"+str(i)+"\n")

for i in range(int((num_samples-pulse_width)/2)+pulse_width, num_samples):
	out.write(str(50+numpy.random.normal(0,stdev))+"\t"+str(i)+"\n")

out.close()
