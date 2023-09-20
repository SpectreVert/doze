/*
 * Created by Costa Bushnaq
 *
 * 26-08-2023 @ 19:51:36
*/

#include <stdio.h>
#include <string.h>

#include "VERSION.H" 

static void usage(void);
static void version(void);

static s32 init(s32,char**);
static s32 stop(s32,char**);
static s32 update(s32,char**);
static s32 help(s32,char**);

s32
main(
    s32 ac,
    char *av[]
) {
    ac--;
    av++;

    if (ac == 0) {
        return update(ac, av);
    } else {
        if (!strcmp(*av, "init") || !strcmp(*av, "in")) {
            return init(ac, av);
        } else if (!strcmp(*av, "stop") || !strcmp(*av, "st")) {
            return stop(ac, av);
        } else if (!strcmp(*av, "help") || !strcmp(*av, "?")) {
            return help(--ac, ++av);
        }
    }

    usage();
    return 0;
}

static s32
init(
    s32 ac,
    char **av
) {
    if (ac == 1) {
        (void)help(ac, av);
        return 1;
    }

    /* look for the monitor */

    /* start the monitor if not found */

    /* get the list of files the monitor has found in the tree */

    /* create .doze/ and the DB if they are not here */

    /* parse Dozefile */

    /* store stuff in the DB (...) */

    return 0;
}

static s32
stop(
    s32 ac,
    char **av
) {
    if (ac == 1) {
        (void)help(ac, av);
        return 1;
    }

    return 0;
}

static s32
update(
    s32 ac,
    char **av
) {
    return 0;
}

static s32
help(
    s32 ac,
    char **av
) {
    if (ac == 0) { usage(); return 0; }
    if (!strcmp(*av, "init") || !strcmp(*av, "in")) {
        printf("Usage: doze %s <DIRECTORY>\n", *av);
        printf("\n");
        printf("Description:\n");
        printf("  Initialize a watch on the given directory. All files\n");
        printf("  recursively found inside it are monitored for changes.\n");
        printf("\n");
        printf("  A valid Dozefile must be present at the root of the\n");
        printf("  given directory.\n");
        printf("\n");
    } else if (!strcmp(*av, "stop") || !strcmp(*av, "st")) {
        printf("Usage: doze %s <DIRECTORY>\n", *av);
        printf("\n");
        printf("Description:\n");
        printf("  Stop watching for changes in the given directory.\n");
        printf("\n");
        printf("  The directory must have been previously registered\n");
        printf("  for watching with a `doze init` call.\n");
        printf("\n");
    }

    return 0;
}

static void
usage(void) {
    version();
    printf("Usage: doze [COMMAND]\n");
    printf("\n");
    printf("Commands:\n");
    printf("  help, ?     Print doze help.\n");
    printf("\n");
    printf("  init, in    Register a directory for monitoring.\n");
    printf("  stop, st    Stop monitoring a directory.\n");
    printf("\n");
    printf("If no command is used, doze executes the required\n");
    printf("orders with the files changed since the last run\n");
    printf("as specified in the current directory Dozefile.\n");
    printf("\n");
    printf("Type 'doze help <COMMAND>' for command-specific help.\n");
}

static void
version(void)
{
    printf("doze %s\n", DOZE_MAJOR "." DOZE_MINOR "." DOZE_PATCH);
}
