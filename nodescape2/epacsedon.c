#include <stdio.h>		/* fgetc, popen */
#include <ctype.h>		/* isspace */
#include <stdlib.h>		/* exit */
#include <unistd.h>		/* gethostname */
#include <string.h>		/* strncpy, strlen, strstr */

#include "nodescape.h"

int send_message(stat_t stat, char *ip, unsigned short port);

int main(int argc, char **argv)
{
	if (argc != 2)
	{
		printf("Usage: %s <config file>\n", argv[0]);
		exit(1);
	}
	FILE *conf = fopen(argv[1], "r");
	config_t config;
	comm_t *commands;
	int ncommands;
	read_config(conf, &config, &commands, &ncommands);
	fclose(conf);

	if (config.host[0] == '\0')
	{
		fprintf(stderr, 
			"Configuration file did not specify a remote host.\n");
		fprintf(stderr, "Exiting...\n");
		exit(1);
	}

	printf("IP: %s\n", config.host);
	printf("Port: %d\n", config.port);

	stat_t stat;
	/* Set hostname, group. This information only needs to be set once. */
	gethostname(stat.host, STATID);
	if (config.group[0] != '\0')
	{
		if (strlen(stat.host) + strlen(config.group) + 2 < STATID)
		{
			strcat(stat.host, ".");
			strcat(stat.host, config.group);
		}
		else
		{
			printf("Length of group is too much to append to hostname.\n");
			printf("Max length for hostname + group is %d.\n", STATID);
		}
	}

	int i;
	for (i = 0; i < ncommands; i++)
	{
		printf("Label: |%s|\n", commands[i].label_str);
		printf("Command: |%s|\n", commands[i].comm_str);

		strncpy(stat.label, commands[i].label_str, STATLAB);	
		stat.status = 0; /* An as yet unimplemented feature */
		stat.when = itime();

		FILE *comm_out = popen(commands[i].comm_str, "r");
		int nextc, i;
		i = 0;
		while (EOF != (nextc = fgetc(comm_out)) && isspace(nextc))
		{ /* Eat leading space */ }
		do {
			if (nextc != '\n')
			{stat.data[i++] = (char)nextc;}
			printf("%c", nextc);
		} while (EOF != (nextc = fgetc(comm_out)) && i < STATDAT - 1);
		pclose(comm_out);
		stat.data[i] = '\0';

		if (!send_message(stat, config.host, config.port))
		{
			fprintf(stderr, "Error: send_message failed!\n");
		}
	}
	return 0;
}



int send_message(stat_t stat, char *ip, unsigned short port)
{
	int sock;
	struct sockaddr_in server_addr;

	struct hostent *host;
	host = (struct hostent *)gethostbyname(ip);
	if (host == 0)
	{
		perror("DNS Lookup failed");
		return 0;
	}
	
	if ((sock = socket(PF_INET, SOCK_DGRAM, IPPROTO_UDP)) < 0)
	{
		perror("Socket failed");
		return 0;
	}

	memset(&server_addr, 0, sizeof(server_addr));
	server_addr.sin_family = AF_INET;
	server_addr.sin_port = htons(port);
	server_addr.sin_addr = *((struct in_addr *)host->h_addr);

	if (sendto(sock, &stat, sizeof(stat), 0, (struct sockaddr *)
		&server_addr, sizeof(server_addr)) != sizeof(stat))
	{
		perror("sendto");
		return 0;
	}
	close(sock);
	return 1;
}
