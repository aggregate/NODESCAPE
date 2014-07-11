#include <stdio.h>
#include <stdlib.h>

#include <my_global.h>
#include <mysql.h>

#include "nodescape.h"

#define QUERYSZ	512

int main(int argc, char **argv)
{
	if (argc != 2)
	{
		printf("Usage: %s <config file>\n", argv[0]);
		exit(1);
	}

	FILE *conf = fopen(argv[1], "r");
	config_t config;
	read_config(conf, &config, NULL, NULL);		

	if (config.user[0] == '\0')
	{
		fprintf(stderr, 
            "Configuration file did not specify a MySQL user.\n");
        fprintf(stderr, "Exiting...\n");
        exit(1);
	}
	if (config.dbhost[0] == '\0')
	{
		fprintf(stderr, 
            "Configuration file did not specify a database host.\n");
        fprintf(stderr, "Exiting...\n");
        exit(1);
	}
	if (config.dbname[0] == '\0')
	{
		fprintf(stderr, 
            "Configuration file did not specify a database to use.\n");
        fprintf(stderr, "Exiting...\n");
        exit(1);
	}
	if (config.table[0] == '\0')
	{
		fprintf(stderr, 
            "Configuration file did not specify a table to use.\n");
        fprintf(stderr, "Exiting...\n");
        exit(1);
	}


	printf("Port Number: %d\n", config.port);

	int sock;
	struct sockaddr_in localAddr;
	
	if ((sock = socket(PF_INET, SOCK_DGRAM, IPPROTO_UDP)) < 0)
	{
		perror("Socket failed");
		exit(1);
	}

	memset(&localAddr, 0, sizeof(localAddr));
	localAddr.sin_family = AF_INET;
	localAddr.sin_addr.s_addr = htonl(INADDR_ANY);
	localAddr.sin_port = htons(config.port);

	if (bind(sock, (struct sockaddr *) &localAddr, sizeof(localAddr)) < 0)
	{
		perror("Bind failed");
		exit(1);
	}

	MYSQL *conn;
	conn = mysql_init(NULL);

	if (conn == NULL)
	{
		printf("Error %u: %s\n", mysql_errno(conn), mysql_error(conn));
		exit(1);
	}
	if (mysql_real_connect(conn, config.dbhost, config.user, 
			(strlen(config.passwd) == 0 ? NULL : config.passwd),
			config.dbname, 0, NULL, 0) == NULL)
	{
		printf(" Error %u: %s\n", mysql_errno(conn), mysql_error(conn));
		exit(1);
	}

	printf("Listening...\n");
	while (1) /* Run until server is killed? */
	{
		stat_t stat;
	
		struct sockaddr_in cliAddr;
		unsigned int cliAddrlen = sizeof(cliAddr);
	
		if ((recvfrom(sock, &stat, sizeof(stat), 0,
			(struct sockaddr *) &cliAddr, &cliAddrlen)
			!= sizeof(stat)))
		{
			perror("recvfrom() failed");
			continue;
		}
		else
		{
			char myquery[QUERYSZ];
			if (
				snprintf(myquery, sizeof(myquery),
					"insert into %s (host, label, data, status, mtime)"
					" values(\"%s\",\"%s\",\"%s\", %d, from_unixtime(%d));",
					config.table, 
					stat.host, stat.label, stat.data, stat.status, 
					stat.when)
				< 0)
			{
				printf("snprintf failed\n");
				continue;
			}
			printf("Query: %s\n", myquery);
			if (mysql_query(conn, myquery))
			{
				printf("Error %u: %s\n", mysql_errno(conn), 
						mysql_error(conn));
				continue;
			}

		} /* else */
	}
	/* not that we'll ever actually get here */
	mysql_close(conn);

	return 0;
}
