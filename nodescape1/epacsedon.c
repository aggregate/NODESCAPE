/*	epacsedon.c

	NodeScape generic client.

	April 2011 by Hank Dietz
*/

#include "nodescape.h"

#define	SENSORSBYTES	(64 * 1024)

double
sense(char *name)
{
	/* Get value of name from sensors */
	int fds[2];
	char buf[SENSORSBYTES];
	register int bytes;
	register int i;
	register int len = strlen(name);
	double value = 0;

	if (pipe(fds) != 0) {
		perror("pipe");
		return(0.0);
	}
	switch (fork()) {
	case -1:
		perror("fork");
		exit(1);
	case 0:
		/* child */
		close(1);		/* close stdout */
		dup2(fds[1], 1);	/* connect stdout to pipe */
		execlp("sensors", "sensors", "-u", NULL);
		perror("sensors");
		write(1, " ", 1);	/* so parent doesn't block */
		exit(0);
	}

	/* Only parent gets here */
	bytes = read(fds[0], buf, SENSORSBYTES);
	for (i=0; i<(bytes-12); ++i) {
		if ((strncmp(name, &(buf[i]), len) == 0) &&
		    (buf[i + len] == ':')) {
			/* Found it! Get value... */
			i += len + 1;
			while ((i < bytes) && (buf[i] == ' ')) ++i;
			close(fds[0]);
			close(fds[1]);
			return(atof(&(buf[i])));
		}
	}

	/* No luck */
	if (bytes > 1) {
		fprintf(stderr, "\"%s\" not found in sensors output\n", name);
	}
	close(fds[0]);
	close(fds[1]);
	return(0.0);
}


int
main(register int argc,
register char **argv)
{
	int sock;
	struct sockaddr_in server_addr;
	struct hostent *host;
	mesg_t m;
	register int nxtarg = 2;
	register int repeat = 0;
	register int savedpos = 0;

	/* Process command line... */
	if (argc < 3) {
		fprintf(stderr,
"Usage: %s server_address {node_number} ((property_name {property_value}) | @{delay})+\n\n"
"If no node_number is specified, hostname is examined for it.\n"
"Each property_name must start with an alpha character.\n"
"Each property_value must be numeric, treated as 0 if omitted.\n"
"However, built-in properties compute their own value:\n"
"\tloadavg\tload average, as per getloadavg()\n"
"\tname:\tfirst value of \"name\" from \"sensors -u\" output\n"
"@delay causes a pause of approximately delay seconds;\n"
"@ by itself causes the sequence to repeat infinitely.\n",
			argv[0]);
		exit(1);
	}

	host = (struct hostent *) gethostbyname(argv[1]);
	if (host == 0) {
		perror(argv[1]);
		exit(1);
	}
	if ((sock = socket(AF_INET, SOCK_DGRAM, 0)) == -1) {
		perror("socket");
		exit(1);
	}

	server_addr.sin_family = AF_INET;
	server_addr.sin_port = htons(MYPORTNO);
	server_addr.sin_addr = *((struct in_addr *)host->h_addr);
	bzero(&(server_addr.sin_zero),8);

	if (isdigit(*(argv[2]))) {
		m.node = atoi(argv[2]);
		nxtarg = 3;
	} else {
		/* Get host name and parse it... */
		char myname[1024];
		register char *p = &(myname[0]);
		gethostname(&(myname[0]), 1024);
		myname[1023] = 0;
		while (*p && !isdigit(*p)) ++p;
		if (*p == 0) {
			fprintf(stderr,
				"%s: could not find node number in hostname '%s'\n",
				argv[0],
				&(myname[0]));
			exit(2);
		} else {
			m.node = atoi(p);
		}
	}

	savedpos = nxtarg;
	do {
		nxtarg = savedpos;

		while (nxtarg < argc) {
			register char *p = argv[nxtarg];
			strncpy(&(m.name[0]), p, PROPLEN);
			if (p[0] == '@') {
				/* Delay or repeat spec. */
				if (p[1] == 0) {
					repeat = 1;
				} else {
					register double d = atof(p + 1);
					printf("Sleeping for %1.2fs....\n", d);
					if (d >= 1.0) {
						sleep((int) d);
						d -= ((int) d);
					}
					if (d >= 0.001) {
						usleep((int) (d * 1000000.0));
					}
				}
				++nxtarg;
				goto skipsend;
			} else if (p[strlen(p)-1] == ':') {
				/* Parse from sensors -u output...
				   but remove the ':' from the name
				*/
				m.name[strlen(p)-1] = 0;
				m.status.value = sense(&(m.name[0]));
				++nxtarg;
			} else if (!strcmp(argv[nxtarg], "loadavg")) {
				/* Use internal load avgerage sampling */
				double loadavg[3];
				printf("Using built-in loadavg function\n");
				m.status.value = ((getloadavg(loadavg, 1) == 1) ?
						  loadavg[0] :
						  0.0);
				++nxtarg;
			} else if ((nxtarg+1 < argc) && !isalpha(*(argv[nxtarg+1]))) {
				m.status.value = atof(argv[nxtarg+1]);
				nxtarg += 2;
			} else {
				m.status.value = 0;
				++nxtarg;
			}

			/* Get as accurate a timestamp as possible */
			m.status.when = dtime();

			/* Write the message */
			sendto(sock,
			       &(m),
			       sizeof(m),
			       0,
			       (struct sockaddr *)&server_addr,
			       sizeof(struct sockaddr));

			printf("Sending to %s from node %d at time %1.2f: %s=%1.2f\n",
			       argv[1],
			       m.node,
			       m.status.when,
			       &(m.name[0]),
			       m.status.value);

skipsend:		;
		}
	} while (repeat);

	/* All's well that exits 0... */
	return(0);
}
