package nsenter

/*
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
__attribute__((constructor)) void enter_namespace(void) {
	char *xperiMoby_pid;
	xperiMoby_pid = getenv("xperiMoby_pid");
	if (xperiMoby_pid) {
		//fprintf(stdout, "got xperiMoby_pid=%s\n", xperiMoby_pid);
	} else {
		//fprintf(stdout, "missing xperiMoby_pid env skip nsenter");
		return;
	}
	char *xperiMoby_cmd;
	xperiMoby_cmd = getenv("xperiMoby_cmd");
	if (xperiMoby_cmd) {
		//fprintf(stdout, "got xperiMoby_cmd=%s\n", xperiMoby_cmd);
	} else {
		//fprintf(stdout, "missing xperiMoby_cmd env skip nsenter");
		return;
	}
	int i;
	char nspath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };
	for (i=0; i<5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", xperiMoby_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);
		if (setns(fd, 0) == -1) {
			//fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			//fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	int res = system(xperiMoby_cmd);
	exit(0);
	return;
}
*/
import "C"
