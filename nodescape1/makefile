all: nodescape epacsedon

nodescape: nodescape.o
	gcc nodescape.o -o nodescape
epacsedon: epacsedon.o
	gcc epacsedon.o -o epacsedon 
nodescape.o: nodescape.c nodescape.h
	gcc -c nodescape.c -o nodescape.o
epacsedon.o: epacsedon.c nodescape.h
	gcc -c epacsedon.c -o epacsedon.o 

clean:
	rm *.o nodescape epacsedon

export:
	svn export svn+sshmud://super.ece.engr.uky.edu/raid/svn/NODESCAPE/
	rm -r NODESCAPE/trunk/ClusterImages NODESCAPE/trunk/ClusterScripts
	rm -r NODESCAPE/jfr-fork
	mv NODESCAPE NODESCAPE-`date +%Y%m%d`
	tar -czvf NODESCAPE-`date +%Y%m%d`.tgz NODESCAPE-`date +%Y%m%d`
	rm -r NODESCAPE-`date +%Y%m%d`
