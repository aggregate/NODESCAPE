#!/usr/bin/python3

import sys
import math

if len(sys.argv) < 2:
	print("usage gensine2f.py out_name")
	sys.exit(0)

out = open(sys.argv[1], "w")

for i in range(200):
		out.write(str(math.sin(i*0.125))+"\t"+str(i)+"\n")

for i in range(200, 400):
		out.write(str(math.sin(i*0.25))+"\t"+str(i)+"\n")

out.close()
