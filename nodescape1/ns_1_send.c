#include "nodescape.h"
#include "ns_1_send.h"

int ns_1_send(mesg_t msg, char *ip, unsigned short port)
{
	int sock;
	struct sockaddr_in server_addr;
	struct hostent *host;

	host = (struct hostent *) gethostbyname(ip);
	if (0 == host)
	{
		perror(ip);
		return NS_NOHOST;	
	}

	if (-1 == (sock = socket(AF_INET, SOCK_DGRAM, 0)))
	{
		perror("socket");
		return NS_NOSOCK;
	}

	server_addr.sin_family = AF_INET;
	server_addr.sin_port = htons(MYPORTNO);
	server_addr.sin_addr = *((struct in_addr *)host->h_addr);
	bzero(&(server_addr.sin_zero), 8);

	/*
		This next part only works for the NSK cluster. I need to get
		node number from the host name. The node names for this cluster
		all start with 's', followed by a single digit. So I'm just going
		to use atoi(&(hostname[1])) to get the number.

		This is not a general solution. 

		Hank, in the future, please write your code for the general case.

		And no, the else clause that looks for a number in the hostname
		doesn't count.
	*/

	char hostname[16]; // 16 will be enough for now.
	gethostname(hostname, 16);
	msg.node = atoi(&(hostname[1]));

	msg.status.when = dtime();

	sendto(sock, &msg, sizeof(msg), 0, (struct sockaddr *)&server_addr,
		sizeof(struct sockaddr));

	return NS_SUCCESS;		
}
