#include "nodescape.h"
int read_line(FILE *fin, char **line)
{
	int line_sz = 64; /* starting line size */
	int line_used = 0;
	*line = (char *)malloc(sizeof(char)*line_sz); 
	int nextc;
	while ('\n' != (nextc = fgetc(fin)) && nextc != EOF)
	{
		if (line_used < line_sz)
		{(*line)[line_used++] = nextc;}
		else
		{
			*line = (char *)realloc(*line, sizeof(char)*line_sz*2);
			line_sz *=2;
			(*line)[line_used++] = nextc;
		}
	}
	if (nextc == EOF)
	{return 0;}
	if (line_used < line_sz)
	{(*line)[line_used++] = '\0';}
	else
	{
		*line = (char *)realloc(*line, sizeof(char)*line_sz*2);
		line_sz *=2;
		(*line)[line_used++] = '\0';
	}
	return 1;
}

char *eat_space(char *p)
{
	while (isspace(*p))
	{p++;}
	return p;
}

#define COMS_START	8
#define SEP	':'
void read_config(FILE *fin, config_t *config, comm_t **coms, int *coms_used)
{
	/* initialize */
	config->host[0] = '\0';
	config->group[0] = '\0';
	config->user[0] = '\0';
	config->dbhost[0] = '\0';
	config->passwd[0] = '\0';
	config->dbname[0] = '\0';
	config->table[0] = '\0';
	config->port = DEFAULT_PORT;
	config->ageint = DEFAULT_AGE;

	int coms_sz = COMS_START;
	if (coms_used && coms)
	{
		*coms_used = 0;
		*coms = (comm_t*)malloc(sizeof(comm_t)*coms_sz);
	}

	char *line = NULL;
	while (read_line(fin, &line))
	{
		char *linep;
		linep = eat_space(line);	
		if (*linep == '#' || *linep == '\0')
		{
			free(line);
			line = NULL;	
			continue;
		}
		else if (linep == strstr(linep, "host"))
		{
			linep += strlen("host");
			linep = eat_space(linep);
			int i = 0;
			while (!isspace(*linep) && 
					i < STRMAX - 1 && 
					*linep != '\0')
			{config->host[i++] = *linep++;}
			config->host[i] = '\0';	
			free(line);
			line = NULL;
			continue;
		}
		else if (linep == strstr(linep, "group"))
		{
			linep += strlen("group");
			linep = eat_space(linep);
			int i = 0;
			while (!isspace(*linep) && 
					i < STRMAX - 1 && 
					*linep != '\0')
			{config->group[i++] = *linep++;}
			config->group[i] = '\0';	
			free(line);
			line = NULL;
			continue;
		}
		else if (linep == strstr(linep, "port"))
		{
			linep += strlen("port");
			linep = eat_space(linep);
			config->port = atoi(linep);	
			free(line);
			line = NULL;
			continue;
		}
		else if (linep == strstr(linep, "ageint"))
		{
			linep += strlen("ageint");
			linep = eat_space(linep);
			config->ageint = atoi(linep);
			free(line);
			line = NULL;
			continue;
		}
		else if (linep == strstr(linep, "user"))
		{
			linep += strlen("user");
			linep = eat_space(linep);
			int i = 0;
			while (!isspace(*linep) &&
				i < STRMAX - 1 &&
				*linep != '\0')
			{config->user[i++] = *linep++;}
			config->user[i] = '\0';
			free(line);
			line = NULL;
			continue;
		}
		else if (linep == strstr(linep, "dbhost"))
		{
			linep += strlen("dbhost");
			linep = eat_space(linep);
			int i = 0;
			while (!isspace(*linep) &&
				i < STRMAX - 1 &&
				*linep != '\0')
			{config->dbhost[i++] = *linep++;}
			config->dbhost[i] = '\0';
			free(line);
			line = NULL;
			continue;
		}
		else if (linep == strstr(linep, "passwd"))
		{
			linep += strlen("passwd");
			linep = eat_space(linep);
			int i = 0;
			while (!isspace(*linep) &&
				i < STRMAX - 1 &&
				*linep != '\0')
			{config->passwd[i++] = *linep++;}
			config->passwd[i] = '\0';
			free(line);
			line = NULL;
			continue;
		}
		else if (linep == strstr(linep, "dbname"))
		{
			linep += strlen("dbname");
			linep = eat_space(linep);
			int i = 0;
			while (!isspace(*linep) &&
				i < STRMAX - 1 &&
				*linep != '\0')
			{config->dbname[i++] = *linep++;}
			config->dbname[i] = '\0';
			free(line);
			line = NULL;
			continue;
		}
		else if (linep == strstr(linep, "table"))
		{
			linep += strlen("table");
			linep = eat_space(linep);
			int i = 0;
			while (!isspace(*linep) &&
				i < STRMAX - 1 &&
				*linep != '\0')
			{config->table[i++] = *linep++;}
			config->table[i] = '\0';
			free(line);
			line = NULL;
			continue;
		}
		/* This starts the specification of a command to be run */
		else if(*linep == SEP && coms && coms_used)
		{
			if (!(*coms_used < coms_sz))
			{
				*coms = (comm_t*)realloc(*coms, sizeof(comm_t)*coms_sz*2);
				coms_sz *= 2;
			}
			linep = eat_space(++linep);	
			int i = 0;
			while (*linep != SEP && *linep != '\0' && i < STATLAB - 1)
			{
				(*coms)[*coms_used].label_str[i++] = *linep++;
			}
			(*coms)[*coms_used].label_str[i] = '\0';
			while (isspace((*coms)[*coms_used].label_str[--i]))
			{/*Find end of label */}
			(*coms)[*coms_used].label_str[++i] = '\0';
			if (*linep == '\0')
			{continue;}

			while (*linep != SEP && *linep != '\0')
			{linep++;}
			if (*linep == '\0')
			{continue;}

			linep++; /* advance past SEP */
			i = 0;
			strncpy((*coms)[*coms_used].comm_str, linep, STRMAX);
			(*coms_used)++;
			free(line);
			line = NULL;
		}
		else
		{
			free(line);
			line = NULL;
		}

	} /* while */
}
