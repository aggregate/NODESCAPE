#!/usr/bin/python3

out = open("cooked_data.txt", "w")

for i in range(100):
	out.write(str(0)+"\t"+str(i)+"\n")

for i in range(100, 200):
	out.write(str(1)+"\t"+str(i)+"\n")

out.close()
