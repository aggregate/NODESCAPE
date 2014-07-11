#!/usr/bin/python3

import sys
import random

if len(sys.argv) < 3:
	print("usage gendata2.py <num samples> <range> <output name>")
	sys.exit(0)

out = open(sys.argv[3], "w")

samples = int(sys.argv[1])
end = int(sys.argv[2])

for i in range(samples):
	out.write(str(random.randrange(0,end))+"\t"+str(i)+"\n")

out.close()
