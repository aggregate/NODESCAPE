#!/usr/bin/python3

import sys

if len(sys.argv) < 4:
	print("usage gen_impulse.py <pulse width> <series length> <output name>")
	sys.exit(0)

out = open(sys.argv[3], "w")

pulse_width = int(sys.argv[1])
length = int(sys.argv[2])

for i in range(int(length/2-pulse_width/2)):
	out.write(str(50)+"\t"+str(i)+"\n")

for i in range(int(length/2-pulse_width/2), int(length/2+pulse_width/2)):
	out.write(str(100)+"\t"+str(i)+"\n")

for i in range(int(length/2+pulse_width/2),length):
	out.write(str(50)+"\t"+str(i)+"\n")
