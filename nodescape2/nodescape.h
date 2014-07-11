#ifndef NODESCAPE_H
#define NODESCAPE_H
#include <sys/time.h>
#include <string.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <netdb.h>

#define STATID 24
#define STATLAB 80
#define STATDAT 80

typedef struct
{
	char host[STATID];
	char label[STATLAB];
	char data[STATDAT];
	int status;
	int when;
} stat_t;

static inline double dtime()
{
	struct timeval tv;
	gettimeofday(&tv, NULL);
	return (tv.tv_sec + (tv.tv_usec / 1000000.0));
}

static inline int itime()
{
	struct timeval tv;
	gettimeofday(&tv, NULL);
	return (int)tv.tv_sec;
}

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define STRMAX	256	
typedef struct
{
	char comm_str[STRMAX];
	char label_str[STATLAB];	
} comm_t;

#define DEFAULT_PORT 45231
#define DEFAULT_AGE 14
typedef struct
{
	char host[STRMAX];
	char group[STRMAX];
	char user[STRMAX];
	char dbhost[STRMAX];
	char passwd[STRMAX];
	char dbname[STRMAX];
	char table[STRMAX];
	unsigned short port;
	unsigned short ageint;
} config_t;

void read_config(FILE *fin, config_t *config, comm_t **coms, int *coms_used);

#endif
