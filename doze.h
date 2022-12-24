/*
 * Created by Costa Bushnaq
 *
 * 11-12-2022 @ 10:56:43
 *
 * see LICENSE
*/

#ifndef ZEE_DOZE_H
#define ZEE_DOZE_H

#ifndef ZEE_BASIC_TYPES
#define ZEE_BASIC_TYPES
#include <stdint.h>
typedef int8_t   s8;
typedef int16_t  s16;
typedef int32_t  s32;
typedef int64_t  s64;
typedef uint8_t  u8;
typedef uint16_t u16;
typedef uint32_t u32;
typedef uint64_t u64;
#endif /* ZEE_BASIC_TYPES */

#ifdef __cplusplus
extern "C" {
#endif

typedef struct Doze_Context Doze_Context;

s32 bake_context(Doze_Context *ctx, s32 ac, char *av[]);

void set_executable_name(Doze_Context *ctx, char const *exe_name);
u32 add_batch(Doze_Context *ctx, char **files, u32 nb_files);

void link_exe(Doze_Context *ctx); 

#ifdef __cplusplus
}
#endif

/* -- No Man's Land between header and source -- */

#ifdef DOZE_IMPLEMENTATION

#ifndef DOZE_CONTEXT_BATCH_CAP
#define DOZE_CONTEXT_BATCH_CAP 256
#endif

#ifndef DOZE_BATCH_FILES_CAP
#define DOZE_BATCH_FILES_CAP 256
#endif

/* ------------- */

#ifndef ZEE_ASSERT
#include <assert.h>
#define ZEE_ASSERT(x) assert(x)
#endif

#ifndef ZEE_MALLOC
#include <stdlib.h>
#define ZEE_MALLOC(n, s) (calloc(n, s))
#define ZEE_FREE(p) (free(p))
#endif
#ifndef ZEE_REALLOC
#define ZEE_REALLOC(p, ns) (realloc(p, ns))
#endif

#include <openssl/md5.h>

#include <fcntl.h>
#include <string.h>
#include <stdio.h>
#include <sys/stat.h>
#include <time.h>
#include <unistd.h>

enum Doze_Option {
    DOZE_OPT_COMPILER_PATH = 0,
    DOZE_OPT_SOURCE_PATH,
    DOZE_OPT_OUTPUT_PATH,

    DOZE_OPT_INCLUDE_PATHS,
    DOZE_OPT_LIBRARY_PATHS,
    DOZE_OPT_LIBRARY_NAMES,
    DOZE_OPT_ADDITIONAL,

    DOZE_NB_OPTIONS
};

typedef struct Doze_Batch Doze_Batch;
struct Doze_Batch {
    u8  has_been_visited;
    u32 id;
    u32 nb_deps;
    u32 deps[DOZE_CONTEXT_BATCH_CAP];
    u32 nb_files;
    struct {
        char const *path;
        u8 is_header;
        u8 digest[MD5_DIGEST_LENGTH];
        u64 last_modified;
    } files[DOZE_BATCH_FILES_CAP];
};

struct Doze_Context {
    u32 nb_batches;
    Doze_Batch batches[DOZE_CONTEXT_BATCH_CAP];
    u32 nb_seen;
    u32 seen_cache[DOZE_CONTEXT_BATCH_CAP];
    u32 nb_resolved;
    u32 resolved_cache[DOZE_CONTEXT_BATCH_CAP];
    struct {
        const char *path;
        u64 last_modified;
    } exe;
    struct {
        u8 is_offset;
        char *opt;
    } opts[DOZE_NB_OPTIONS];
};

#define DOZE___ARRAY_APPEND_ELEM(arr, size, elem) do {\
    u32 __i = 0; \
    for (; __i < size; __i++); \
    arr[__i] = elem; \
    size += 1; \
    } while (0)

#define DOZE___ENSURE_DIRSLASH(str) \
    if (str[strlen(str) - 1] != '/') { \
        str = doze___merge(str, "/", 0x0); \
    }

static char *doze___merge(char *a, char const *b, char const *separator)
{
    u64 alen = strlen(a);
    u64 blen = strlen(b);
    u64 slen = 0;

    if (separator) { slen = strlen(separator); }
    a = ZEE_REALLOC(a, alen + blen + slen + 1);
    if (separator) { a = strcat(a, separator); }
    
    return strcat(a, b);
}

static s32 doze___is_file_header(const char *fname)
{
    static char *header_ext[3] = {
        ".h",
        ".hh",
        ".hpp",
    };

    char *s;
    for (u64 i = 0; header_ext[i]; i++) {
        if ((s = strstr(fname, header_ext[i]))) {
            if (*(s + strlen(header_ext[i])) == 0)
                return 1;
        }
    }
    
    return 0;
}

static s32 doze___is_file_source(const char *fname)
{
    static char *source_ext[3] = {
        ".c",
        ".cc",
        ".cpp",
    };

    char *s;
    for (u64 i = 0; source_ext[i]; i++) {
        if ((s = strstr(fname, source_ext[i]))) {
            if (*(s + strlen(source_ext[i])) == 0)
                return 1;
        }
    }

    return 0;
}

extern s32 doze___is_file_usable(const char *fpath)
{
    u64 len = strlen(fpath);

    if (len <= 2) {
        // file name not long enough
        return 0;
    } else if (access(fpath, R_OK)) {
        // file is not accessible
        return 0;
    }

    return 1;
}

static u64 doze___get_timestamp(const char *fpath)
{
    struct stat fs;

    if (stat(fpath, &fs)) {
        return 0;
    }

    return fs.st_mtime;
}

s32 doze___resolve_dependencies(Doze_Context *ctx, u32 id)
{
    u32 nb_deps = ctx->batches[id].nb_deps;
    u32 const *deps = ctx->batches[id].deps;

    DOZE___ARRAY_APPEND_ELEM(ctx->seen_cache, ctx->nb_seen, id);

    // we check all dependencies
    for (u32 i = 0; i < nb_deps; i++) {
        // check if this dependency has been resolved
        u8 is_resolved = 0;
        for (u32 c = 0; c < ctx->nb_resolved; c++) {
            if (ctx->resolved_cache[c] == deps[i]) {
                // it has been resolved already
                is_resolved = 1;
            }
        }
        if (!is_resolved) {
            // check if dependency is circular
            for (u32 s = 0; s < ctx->nb_seen; s++) {
                if (ctx->seen_cache[s] == deps[i]) {
                    printf("We got a circular dependency! %d <-> %d\n",
                        id, deps[i]);
                    return -1;
                }
            }
            if (doze___resolve_dependencies(ctx, deps[i])) {
                return -1;
            }
        }
    }

    DOZE___ARRAY_APPEND_ELEM(ctx->resolved_cache, ctx->nb_resolved, id);
    return 0;
}

#define DOZE___SPLIT_REPLACE(src, dest, delim, replace) do {\
    char *s = src; \
    char *n; \
    for (n = strchr(s, delim); n != 0x0; n = strchr(s, delim)) { \
        s[n - s] = '\0'; \
        dest = doze___merge(dest, s, replace); \
        s[n - s] = delim; \
        s = n + 1; \
    } \
    dest = doze___merge(dest, s, replace); \
    } while (0)

s32 doze___compile(Doze_Context *ctx, char const *fpath)
{
    ZEE_ASSERT(ctx->opts[DOZE_OPT_COMPILER_PATH].opt);

    char *cmd = ZEE_MALLOC(1, 1); // holds the compilation command
    char *tmp = ZEE_MALLOC(1, 1); // holds the object file constructed name

    // we form the compilation command:
    //
    // cmd = "<compiler> -c <source_file> -o <object_file>"
    //
    // Ex:
    //
    //    tcc -c src/main.c -o build/src_main.o
    //                               ^
    //                               here starts the contents of `tmp`
    //
    cmd = doze___merge(cmd, " -c ", ctx->opts[DOZE_OPT_COMPILER_PATH].opt);
    if (ctx->opts[DOZE_OPT_SOURCE_PATH].opt) {
        tmp = doze___merge(tmp, ctx->opts[DOZE_OPT_SOURCE_PATH].opt, 0x0);
        DOZE___ENSURE_DIRSLASH(tmp);
    }
    tmp = doze___merge(tmp, fpath, 0x0);
    cmd = doze___merge(cmd, " -o ", tmp);
    if (ctx->opts[DOZE_OPT_OUTPUT_PATH].opt) {
        cmd = doze___merge(cmd, ctx->opts[DOZE_OPT_OUTPUT_PATH].opt, 0x0);
        DOZE___ENSURE_DIRSLASH(cmd);
    }
    for (u32 i = 0; i < strlen(tmp); i++) {
        if (tmp[i] == '/') {
            tmp[i] = '_';
        }
    }
    tmp[strlen(tmp) - 1] = 'o';
    cmd = doze___merge(cmd, tmp, 0x0);

    // @Todo change the location of the linking flags
    if (ctx->opts[DOZE_OPT_INCLUDE_PATHS].opt) {
        DOZE___SPLIT_REPLACE(
            ctx->opts[DOZE_OPT_INCLUDE_PATHS].opt, cmd, ';', " -I"
        );
    }
    if (ctx->opts[DOZE_OPT_ADDITIONAL].opt) {
        DOZE___SPLIT_REPLACE(
            ctx->opts[DOZE_OPT_ADDITIONAL].opt, cmd, ';', " "
        );
     }

    puts(cmd);
    s32 res = system(cmd);

    ZEE_FREE(cmd);

    return res;
}

extern s32 bake_context(Doze_Context *ctx, s32 ac, char *av[])
{
    static char const *flags[DOZE_NB_OPTIONS] = {
        [DOZE_OPT_COMPILER_PATH] = "-C",
        [DOZE_OPT_SOURCE_PATH]   = "-S",
        [DOZE_OPT_OUTPUT_PATH]   = "-B",

        [DOZE_OPT_INCLUDE_PATHS] = "-I",
        [DOZE_OPT_LIBRARY_PATHS] = "-L",
        [DOZE_OPT_LIBRARY_NAMES] = "-l",
        [DOZE_OPT_ADDITIONAL]    = "-A",
    };

    memset(ctx, 0, sizeof(*ctx));

    for (u32 i = 0; i < ac; i++) {
        for (u32 f = 0; f < DOZE_NB_OPTIONS; f++) {
            if (strncmp(av[i], flags[f], 2)) {
                continue;
            }

            // found a flag, now we determine where exactly is the option
            // we also silently ignore wrong input
            if (strlen(av[i]) > 2) {
                // the option follows the flag
                ctx->opts[f].is_offset = 1;
                ctx->opts[f].opt = av[i] + 2;
            } else if (strlen(av[i]) == 2 && i + 1 < ac) {
                // the option should be in the next argument
                ctx->opts[f].is_offset = 0;
                ctx->opts[f].opt = av[i + 1];
            }
        }
    }

    return 0;
}

void set_executable_name(Doze_Context *ctx, char const *exe_name)
{
    ZEE_ASSERT(exe_name);
    ZEE_ASSERT(!ctx->exe.path);

    char *real_exe_path = ZEE_MALLOC(1, 1);
    char const *out_path = ctx->opts[DOZE_OPT_OUTPUT_PATH].opt;

    if (out_path) {
        real_exe_path = doze___merge(real_exe_path, out_path, 0x0);
        DOZE___ENSURE_DIRSLASH(real_exe_path);
    }
    real_exe_path = doze___merge(real_exe_path, exe_name, 0x0);
    ctx->exe.path = real_exe_path;
    if (doze___is_file_usable(real_exe_path)) {
        ctx->exe.last_modified = doze___get_timestamp(real_exe_path);
    } else {
        ctx->exe.last_modified = 0;
    }
}

extern u32 add_batch(Doze_Context *ctx, char **files, u32 nb_files)
{
    ZEE_ASSERT(ctx->nb_batches < DOZE_CONTEXT_BATCH_CAP);
    ZEE_ASSERT(nb_files < DOZE_BATCH_FILES_CAP);
    Doze_Batch *bat = ctx->batches + ctx->nb_batches;

    memset(bat, 0, sizeof(Doze_Batch));
    bat->id = ctx->nb_batches;
    
    for (u32 i = 0; i < nb_files; i++) {
        char *real_path = ZEE_MALLOC(1, 1);
        char const *src_path = ctx->opts[DOZE_OPT_SOURCE_PATH].opt;
        s32 is_valid;

        if (src_path) {
            real_path = doze___merge(real_path, src_path, 0x0);
            DOZE___ENSURE_DIRSLASH(real_path);
        }
        real_path = doze___merge(real_path, files[i], 0x0);
        if (!doze___is_file_usable(real_path) ||
           (!doze___is_file_header(real_path) &&
            !doze___is_file_source(real_path))) {
                // we silently drop bad files
                ZEE_FREE(real_path);
                continue;
        }
        bat->files[bat->nb_files].path = files[i]; // @Note store relative path
        bat->files[bat->nb_files].is_header = doze___is_file_header(real_path);
        bat->files[bat->nb_files].last_modified = doze___get_timestamp(real_path);
        bat->nb_files += 1;
        ZEE_FREE(real_path);
    }

    ZEE_ASSERT(bat->nb_files > 0);
    ctx->nb_batches += 1;

    return bat->id;
}

void link_exe(Doze_Context *ctx)
{
    ZEE_ASSERT(ctx->exe.path);
    ZEE_ASSERT(ctx->opts[DOZE_OPT_COMPILER_PATH].opt);

    u32 nb_compilations = 0;
    u64 last_exe_mod = ctx->exe.last_modified;
    char *cmd = ZEE_MALLOC(1, 1);

    cmd = doze___merge(cmd, ctx->opts[DOZE_OPT_COMPILER_PATH].opt, 0x0);
    cmd = doze___merge(cmd, ctx->exe.path, " -o ");

    doze___resolve_dependencies(ctx, 0);
    for (u32 b = 0; b < ctx->nb_resolved; b++) {
        // for each dependency batch check every file
        for (u32 f = 0; f < ctx->batches[b].nb_files; f++) {
            // for each file check if it has been updated later than the
            // last binary linking AND that it's not a header (important)
            if (ctx->batches[b].files[f].last_modified > last_exe_mod &&
                !ctx->batches[b].files[f].is_header) {
                if (!doze___compile(ctx, ctx->batches[b].files[f].path)) {
                    nb_compilations += 1;
                } else {
                    printf("doze: error compiling file '%s'\n",
                            ctx->batches[b].files[f].path);
                    exit(1);
                }
            }
        }
    }

    if (!nb_compilations) {
        ZEE_FREE(cmd);
        puts("doze: nothing to do");
        return;
    }

    // now we loop again to form the linking command
    for (u32 b = 0; b < ctx->nb_resolved; b++) {
        for (u32 f = 0; f < ctx->batches[b].nb_files; f++) {
            char *tmp = ZEE_MALLOC(1, 1);
            if (ctx->opts[DOZE_OPT_SOURCE_PATH].opt) {
                tmp = doze___merge(tmp, ctx->opts[DOZE_OPT_SOURCE_PATH].opt, 0x0);
                DOZE___ENSURE_DIRSLASH(tmp);
            }
            tmp = doze___merge(tmp, ctx->batches[b].files[f].path, 0x0);
            for (u32 i = 0; i < strlen(tmp); i++) {
                if (tmp[i] == '/') {
                    tmp[i] = '_';
                }
            }
            tmp[strlen(tmp) - 1] = 'o';
            if (ctx->opts[DOZE_OPT_OUTPUT_PATH].opt) {
                cmd = doze___merge(cmd, ctx->opts[DOZE_OPT_OUTPUT_PATH].opt, " ");
                DOZE___ENSURE_DIRSLASH(cmd);
            }
            cmd = doze___merge(cmd, tmp, 0x0);
            ZEE_FREE(tmp);
        }
    }

    if (ctx->opts[DOZE_OPT_LIBRARY_PATHS].opt) {
        DOZE___SPLIT_REPLACE(
            ctx->opts[DOZE_OPT_LIBRARY_PATHS].opt, cmd, ';', " -L"
        );
    }
    if (ctx->opts[DOZE_OPT_LIBRARY_NAMES].opt) {
        DOZE___SPLIT_REPLACE(
            ctx->opts[DOZE_OPT_LIBRARY_NAMES].opt, cmd, ';', " -l"
        );
    }

    puts(cmd);
    system(cmd);

    printf("doze: compiled %d file%s\n", nb_compilations,
        nb_compilations == 1 ? "" : "s");
    printf("doze: linked '%s'\n", ctx->exe.path);

    ZEE_FREE(cmd);
}

s32 add_dependency(Doze_Context *ctx, u32 child, u32 parent)
{
    ZEE_ASSERT(parent <= ctx->nb_batches);
    ZEE_ASSERT(child  <= ctx->nb_batches);
    ZEE_ASSERT(ctx->batches[child].nb_deps < DOZE_CONTEXT_BATCH_CAP);

    ctx->batches[child].deps[ctx->batches[child].nb_deps] = parent;
    ctx->batches[child].nb_deps += 1;

#ifdef DOZE_CHECK_RECURSIVE_DEPENDENCY
    ctx->nb_seen = 0;
    ctx->nb_resolved = 0;

    if (doze___resolve_dependencies(ctx, child)) {
        exit(1);
    }
#endif

    return 0;
}

#endif /* DOZE_IMPLEMENTATION */
#endif /* ZEE_DOZE_H */
