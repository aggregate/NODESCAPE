#!/usr/bin/python3

import sys
import math

if len(sys.argv) < 3:
	print("usage %s <num samples> <out_name>"%sys.argv[0])
	sys.exit(0)

out = open(sys.argv[2], "w")

for i in range(int(sys.argv[1])):
		out.write(str(math.sin(i*0.125))+"\t"+str(i)+"\n")

out.close()
