#!/usr/bin/python3

import sys

if len(sys.argv) < 2:
	print("usage gendata2.py out_name")
	sys.exit(0)

out = open(sys.argv[1], "w")

for i in range(50):
	out.write(str(i)+"\t"+str(i)+"\n")

x = 100 
for i in range(50, 100):
	out.write(str(x-i)+"\t"+str(i)+"\n")
