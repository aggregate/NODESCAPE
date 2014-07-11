#!/usr/bin/python3

import sys

if len(sys.argv) < 3:
	print("Usage: %s <num samples> <file name>\n"%sys.argv[0])
	sys.exit(0)

out = open(sys.argv[2], "w")
num_samples = int(sys.argv[1])

for i in range(int(num_samples/2)):
	out.write(str(0)+"\t"+str(i)+"\n")

for i in range(int(num_samples/2), num_samples):
	out.write(str(1)+"\t"+str(i)+"\n")

out.close()
