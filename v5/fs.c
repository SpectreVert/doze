/*
 * Created by Costa Bushnaq
 *
 * 09-09-2023 @ 23:12:02
 *
 * see LICENSE
*/

#include <errno.h>

void
fs_init(
    struct fuse_conn_info *connection,
    struct fuse_config *config
) {
    (void)connection;
    (void)config;
    /* I don't know what to do here yet. So we do nothing. */
}

int
fs_getattr(
    char const *path,
    struct stat *fbuf,
    struct fuse_file_info *fi
) {
    (void)fi;

    return -ENOSYS;
}
