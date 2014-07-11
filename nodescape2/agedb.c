#include <stdio.h>
#include <stdlib.h>

#include <my_global.h>
#include <mysql.h>

#include "nodescape.h"

int main(int argc, char **argv)
{
	if (argc != 2)
	{
		printf("Usage: %s <config file>\n", argv[0]);
	}

	FILE *conf = fopen(argv[1], "r");
	config_t config;
	read_config(conf, &config, NULL, NULL);

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
		printf("Error %u: %s\n", mysql_errno(conn), mysql_error(conn));
		exit(1);
	}

	char myquery[256];

	if (0 > snprintf(myquery, 256, 
			"delete from ukanstats where ctime < date_sub(now(), "
			" interval %d day);", config.ageint))
	{
		printf("Error: sprintf failed.\n");
		exit(1);
	}
	printf("%s\n",myquery);

	if (mysql_query(conn, myquery))
	{
		printf("Error %u: %s\n", mysql_errno(conn),
				mysql_error(conn));
	}
	mysql_close(conn);

	return 0;
}
