#!/usr/bin/python3

def main():

	if (len(sys.argv) != 2):
		print("Usage: %s <config file>"%(sys.argv[0]))
		sys.exit()
	
	config = read_config_simple(sys.argv[1])	

	fin = open("out.txt", "r")

	reals = []
	imags = []

	for line in fin:
		real, imag = line.split()
		reals.append(float(real))
		imags.append(float(imag))	
		

	for real in reals:
		update_rrd()

	fin.close()

main()
