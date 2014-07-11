all: nodescape epacsedon agedb

SQLFLAGS:=$(shell mysql_config --cflags --libs) 
nodescape: nodescape.o config.o
	gcc nodescape.o config.o $(SQLFLAGS) -o nodescape
epacsedon: epacsedon.o config.o
	gcc epacsedon.o config.o -o epacsedon
agedb: agedb.o config.o
	gcc agedb.o config.o $(SQLFLAGS) -o agedb
epacsedon.o: epacsedon.c nodescape.h
	gcc -c epacsedon.c -o epacsedon.o
config.o: config.c nodescape.h 
	gcc -c config.c -o config.o
nodescape.o: nodescape.c nodescape.h
	gcc -c nodescape.c $(SQLFLAGS) -o nodescape.o
agedb.o: agedb.c nodescape.h
	gcc -c agedb.c $(SQLFLAGS) -o agedb.o

clean:
	rm *.o nodescape epacsedon agedb

export:
	svn export svn+sshmud://super.ece.engr.uky.edu/raid/svn/NODESCAPE/jfr-fork/
	mv jfr-fork NODESCAPE2-`date +%Y%m%d`
	tar -czvf NODESCAPE2-`date +%Y%m%d`.tgz NODESCAPE2-`date +%Y%m%d`
	rm -r NODESCAPE2-`date +%Y%m%d`

