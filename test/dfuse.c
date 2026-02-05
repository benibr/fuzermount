/*
 * This is a simple mock of dfuse
 * to test the calling of the fusermount3
 * wrapper.
  */

#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <errno.h>
#include <sys/wait.h>

int main(int argc, char *argv[]) {
    argv[0] = "/usr/bin/fusermount3";

    pid_t pid = fork();
    if (pid == -1) {
        perror("fork");
    }
 
    if (pid == 0) {  // Child process
        if (execvp(argv[0], argv) == -1) {
            perror("execvp");
            exit(EXIT_FAILURE);
        }
    } else {  // Parent process
        int status;
        waitpid(pid, &status, 0);
    }
    return 0;
}
