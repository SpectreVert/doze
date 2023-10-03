/*
 * Created by Costa Bushnaq
 *
 * 26-08-2023 @ 19:51:36
*/

#include <stdio.h>
#include <string.h>

#include "doze.h"

static void usage(void);
static void version(void);

static s32 list(s32,char**);
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
    } else if (!strcmp(*av, "list") || !strcmp(*av, "li")) {
        return list(ac, av);
    } else if (!strcmp(*av, "stop") || !strcmp(*av, "st")) {
        return stop(ac, av);
    } else if (!strcmp(*av, "help") || !strcmp(*av, "?")) {
        return help(--ac, ++av);
    } else {
        return update(ac, av);
    }

    usage();
    return 0;
}

static s32
list(
    s32 ac,
    char **av
) {
    if (ac > 1) {
        (void)help(ac, av);
        return 1;
    }

    return 0;
}

static s32
stop(
    s32 ac,
    char **av
) {
    if (ac == 0) {
        /* targeting the local Dozefile */
        return 0;
    }

    /* targeting a list of other Dozefiles */
    return 0;
}

static s32
update(
    s32 ac,
    char **av
) {
    if (ac == 0) {
        /* targeting the local Dozefile */
        return 0;
    }

    /* targeting a list of other Dozefiles */
    return 0;
}

static s32
help(
    s32 ac,
    char **av
) {
    if (ac == 0) { usage(); return 0; }
    if (!strcmp(*av, "list") || !strcmp(*av, "li")) {
        printf("Usage: doze %s\n", *av);
        printf("\n");
        printf("Write me\n");
        printf("\n");
    } else if (!strcmp(*av, "stop") || !strcmp(*av, "st")) {
        printf("Usage: doze %s [DIRECTORIES...]\n", *av);
        printf("\n");
        printf("Write me\n");
        printf("\n");
    }

    return 0;
}

static void
usage(void) {
    version();
    printf("Usage:\n");
    printf("    doze [DIRECTORIES...]\n");
    printf("    doze [COMMAND]\n");
    printf("\n");
    printf("Commands:\n");
    printf("  help, ?     Print doze help.\n");
    printf("\n");
    printf("  list, li    Show a list of all monitored Dozefiles.\n");
    printf("  stop, st    Stop monitoring a Dozefile.\n");
    printf("\n");
    printf("\n");
    printf("Type 'doze help <COMMAND>' for command-specific help.\n");
}

static void
version(void)
{
    printf("doze %s\n", DOZE_MAJOR "." DOZE_MINOR "." DOZE_PATCH);
}
