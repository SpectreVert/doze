/*
 * Created by Costa Bushnaq
 *
 * 24-12-2022 @ 12:14:52
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
#endif

struct Doze_Context;

void context(struct Doze_Context *, s32, char **);
u32  batch(struct Doze_Context *, char **, u32);
void mark(struct Doze_Context *, u32, u32);
void exe(struct Doze_Context *, char const *);

/* BULLETIN ( DOZE )
 *
 * RELEASES
 *
 * 01-02-2023 - 0.1
 *
 * TODOLIST
 *
 *   <> `batch` should use the pointer to the array as the id.
 *      Thus `mark` would also be able to use the pointer in the same way.
 *   <> implement `lib` function for linking a library
 *   <> rewrite the comments so they are more helpful+better formatted
 *
*/

#ifdef DOZE_IMPLEMENTATION

#ifndef DOZE_BATCH_CAP
#define DOZE_BATCH_CAP 256
#endif

#ifndef DOZE_FILES_PER_BATCH
#define DOZE_FILES_PER_BATCH 256
#endif

#ifndef DOZE_OPTION_SEPARATOR
#define DOZE_OPTION_SEPARATOR ';'
#endif

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

struct Doze_Batch {
    u8  is_flagged;
    u32 id;

    u32 nb_childs;
    u32 child_ids[DOZE_BATCH_CAP];

    u32 nb_files;
    struct {
        u8  is_header;
        u64 last_modified;
        char const *path;
    } files[DOZE_FILES_PER_BATCH];
};

struct Doze_Context {
    struct {
        u8 is_offset;
        char *opt;
    } opts[DOZE_NB_OPTIONS];

    struct {
        u64 last_modified;
        const char *path;
    } exe;

    u32 nb_batches;
    struct Doze_Batch bucket[DOZE_BATCH_CAP];

    char *object_files;
    u32 nb_compiled_files;
    u32 nb_modified_files;

    u32 nb_seen;
    u32 seen_cache[DOZE_BATCH_CAP];
    u32 nb_resolved;
    u32 resolved_cache[DOZE_BATCH_CAP];
};

/* --- Private parts O.o --- */

#define DOZE___ENSURE_DIRSLASH(path)            \
    if (path[strlen(path) - 1] != '/') {        \
        path = doze___merge(path, "/", 0x0);    \
    }

#define DOZE___ARRAY_APPEND_ELEM(arr, size, elem) do {  \
    arr[size] = elem;                                   \
    size += 1;                                          \
    } while (0)

#define DOZE___SPLIT_REPLACE_MERGE_INTO(src, sep, replace, dest) do {   \
    char *__s = src; char *__n;                                         \
    for (__n = strchr(__s, sep); __n; __n = strchr(__s, sep)) {         \
        __s[__n - __s] = 0;                                             \
        dest = doze___merge(dest, __s, replace);                        \
        __s[__n - __s] = sep;                                           \
        __s = __n + 1;                                                  \
    }                                                                   \
    dest = doze___merge(dest, __s, replace);                            \
    } while (0)

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

static u8 doze___is_file_header(const char *fname)
{
    static char *ext[3] = {
        ".h",
        ".hh",
        ".hpp",
    };

    char *s;
    for (u64 i = 0; ext[i]; i++) {
        if ((s = strstr(fname, ext[i]))) {
            if (*(s + strlen(ext[i])) == 0) { return 1; }
        }
    }

    return 0;
}

static u8 doze___is_file_source(const char *fname)
{
    static char *ext[3] = {
        ".c",
        ".cc",
        ".cpp",
    };

    char *s;
    for (u64 i = 0; ext[i]; i++) {
        if ((s = strstr(fname, ext[i]))) {
            if (*(s + strlen(ext[i])) == 0) { return 1; }
        }
    }

    return 0;
}

static u8 doze___is_file_usable(const char *fpath)
{
    u64 len = strlen(fpath);

    // file name not long enough
    if (len <= 2) { return 0; }
    
    // file is not accessible
    if (access(fpath, R_OK)) { return 0; }

    return 1;
}

static u64 doze___get_timestamp(const char *fpath)
{
    struct stat fs;

    if (stat(fpath, &fs)) { return 0; }

    return fs.st_mtime;
}

s32 doze___resolve_dependencies(struct Doze_Context *ctx, u32 id)
{
    u32 nb_childs = ctx->bucket[id].nb_childs;
    u32 const *child_ids = ctx->bucket[id].child_ids;

    DOZE___ARRAY_APPEND_ELEM(ctx->seen_cache, ctx->nb_seen, id);

    // we check all dependencies
    for (u32 i = 0; i < nb_childs; i++) {
        // check if this dependency has been resolved
        u8 is_resolved = 0;
        for (u32 c = 0; c < ctx->nb_resolved; c++) {
            if (ctx->resolved_cache[c] == child_ids[i]) {
                // it has been resolved already
                is_resolved = 1;
                break;
            }
        }
        if (!is_resolved) {
            // check if dependency is circular
            for (u32 s = 0; s < ctx->nb_seen; s++) {
                if (ctx->seen_cache[s] == child_ids[i]) {
                    printf("doze: circular dependency! [%d] <-> [%d]\n",
                           id, child_ids[i]);
                    return -1;
                }
            }
            if (doze___resolve_dependencies(ctx, child_ids[i])) {
                return -1;
            }
        }
    }

    DOZE___ARRAY_APPEND_ELEM(ctx->resolved_cache, ctx->nb_resolved, id);
    return 0;
}

void doze___compile_file(struct Doze_Context *ctx, char const *fpath, u8 for_real)
{
    ZEE_ASSERT(ctx->opts[DOZE_OPT_COMPILER_PATH].opt);

    // holds the object file name which gets constructed from the source file path
    // slashes '/' get changed to underscores '_' and the extension switched to an .o
    char *obj = ZEE_MALLOC(1, 1);
    u32 last_dot = 0;
    u8  skip = 0;

    if (ctx->opts[DOZE_OPT_OUTPUT_PATH].opt) {
        obj = doze___merge(obj, ctx->opts[DOZE_OPT_OUTPUT_PATH].opt, 0x0);
        DOZE___ENSURE_DIRSLASH(obj);
        skip = 1;
    }
    obj = doze___merge(obj, fpath, 0x0);
    for (u32 i = 0; obj[i]; i++) {
        if (obj[i] == '/') {
            if (skip) {
                skip = 0;
                continue;
            }
            obj[i] = '_';
        } else if (obj[i] == '.') {
            last_dot = i;
        }
    }

    ZEE_ASSERT(last_dot); // this should never fail
    obj[last_dot + 1] = 'o';
    if (obj[last_dot + 2] != 0) {
        obj[last_dot + 2] = 0;
    }

    // here we add the object file to the final link command
    ctx->object_files = doze___merge(ctx->object_files, obj, " ");

    // verify that we actually want to compile the file
    if (!for_real) {
        ZEE_FREE(obj);
        return;
    }

    // now we construct the compile command
    //
    // cmd = <compiler> -c <source_file> -o <object_file> [options...]
    //
    // Ex:
    //     tcc -c src/main.c -o build/src_main.o [-I....]
    //
    char *cmd = ZEE_MALLOC(1, 1);
    cmd = doze___merge(cmd, ctx->opts[DOZE_OPT_COMPILER_PATH].opt, 0x0);
    cmd = doze___merge(cmd, fpath, " -c ");
    cmd = doze___merge(cmd, obj, " -o ");

    // we add the compilation flags
    if (ctx->opts[DOZE_OPT_INCLUDE_PATHS].opt) {
        DOZE___SPLIT_REPLACE_MERGE_INTO(
            ctx->opts[DOZE_OPT_INCLUDE_PATHS].opt,
            DOZE_OPTION_SEPARATOR, " -I ",
            cmd);
    }
    if (ctx->opts[DOZE_OPT_ADDITIONAL].opt) {
        DOZE___SPLIT_REPLACE_MERGE_INTO(
            ctx->opts[DOZE_OPT_ADDITIONAL].opt,
            DOZE_OPTION_SEPARATOR, " ",
            cmd);
    }

    puts(cmd);
    system(cmd);

    ctx->nb_compiled_files += 1;

    ZEE_FREE(obj);
    ZEE_FREE(cmd);
}

void doze___process_batch(struct Doze_Context *ctx, u32 id)
{
    ZEE_ASSERT(id < ctx->nb_batches);

    struct Doze_Batch *batch = ctx->bucket + id;

    // first we figure out if any parent batch has been updated
    for (u32 p = 0; p < batch->nb_childs; p++) {
        if (ctx->bucket[p].is_flagged) {
            batch->is_flagged = 1;
            break;
        }
    }

    for (u32 f = 0; f < batch->nb_files; f++) {
        if (batch->files[f].last_modified > ctx->exe.last_modified) {
            ctx->nb_modified_files += 1;
            batch->is_flagged = 1;
        }
        if (!batch->files[f].is_header) {
            doze___compile_file(ctx, batch->files[f].path, batch->is_flagged);
        }
    }
}

/* --- Public display --- */

void context(struct Doze_Context *ctx, s32 ac, char **av)
{
    static char const *flags[DOZE_NB_OPTIONS] = {
        // single options
        [DOZE_OPT_COMPILER_PATH] = "-C",
        [DOZE_OPT_SOURCE_PATH]   = "-S",
        [DOZE_OPT_OUTPUT_PATH]   = "-B",

        // these options can be manifold and separated with ';'
        [DOZE_OPT_INCLUDE_PATHS] = "-I",
        [DOZE_OPT_LIBRARY_PATHS] = "-L",
        [DOZE_OPT_LIBRARY_NAMES] = "-l",
        [DOZE_OPT_ADDITIONAL]    = "-A",
    };

    memset(ctx, 0, sizeof(*ctx));

    for (s32 i = 0; i < ac; i++) {
        // we check each argument to find doze options
        for (u32 f = 0; f < DOZE_NB_OPTIONS; f++) {
            if (strncmp(av[i], flags[f], 2)) {
                continue;
            }

            // the argument matches an option
            if (strlen(av[i]) > 2) {
                ctx->opts[f].is_offset = 1;
                ctx->opts[f].opt = av[i] + 2;
            } else if (strlen(av[i]) == 2 && i + 1 < ac) {
                ctx->opts[f].is_offset = 0;
                ctx->opts[f].opt = av[i + 1];
            }
        }
    }
}

u32 batch(struct Doze_Context *ctx, char **files, u32 nb_files)
{
    ZEE_ASSERT(ctx->nb_batches < DOZE_BATCH_CAP);
    ZEE_ASSERT(nb_files < DOZE_FILES_PER_BATCH);
    
    struct Doze_Batch *batch = ctx->bucket + ctx->nb_batches;

    memset(batch, 0, sizeof(*batch));
    batch->is_flagged = 0;
    batch->id = ctx->nb_batches;

    for (u32 f = 0; f < nb_files; f++) {
        char *real_path = ZEE_MALLOC(1, 1);
        u32 fidx = batch->nb_files;
        
        if (ctx->opts[DOZE_OPT_SOURCE_PATH].opt) {
            real_path = doze___merge(real_path, ctx->opts[DOZE_OPT_SOURCE_PATH].opt, 0x0);
            DOZE___ENSURE_DIRSLASH(real_path);
        }

        real_path = doze___merge(real_path, files[f], 0x0);
        if (!doze___is_file_usable(real_path) ||
           (!doze___is_file_source(real_path) &&
            !doze___is_file_header(real_path))) {
            // we silently drop unusable files
            ZEE_FREE(real_path);
            continue;
        }
        batch->files[fidx].path = real_path;
        batch->files[fidx].is_header = doze___is_file_header(real_path);
        batch->files[fidx].last_modified = doze___get_timestamp(real_path);
        batch->nb_files += 1;
    }

    ZEE_ASSERT(batch->nb_files);
    ctx->nb_batches += 1;

    return batch->id;
}

void mark(struct Doze_Context *ctx, u32 target_id, u32 dependency_id)
{
    ZEE_ASSERT(target_id < ctx->nb_batches);
    ZEE_ASSERT(dependency_id  < ctx->nb_batches);
    ZEE_ASSERT(ctx->bucket[dependency_id].nb_childs < DOZE_BATCH_CAP);

    struct Doze_Batch *target = ctx->bucket + target_id;

    target->child_ids[target->nb_childs] = dependency_id;
    target->nb_childs += 1;
}

void exe(struct Doze_Context *ctx, char const *name)
{
    ZEE_ASSERT(!ctx->exe.path);

    ctx->object_files = ZEE_MALLOC(1, 1);
    char *real_path = ZEE_MALLOC(1, 1);

    // we construct the executable target name
    if (ctx->opts[DOZE_OPT_OUTPUT_PATH].opt) {
        real_path = doze___merge(real_path, ctx->opts[DOZE_OPT_OUTPUT_PATH].opt, 0x0);
        DOZE___ENSURE_DIRSLASH(real_path);
    }
    real_path = doze___merge(real_path, name, 0x0);
    ctx->exe.path = real_path;
    
    // retrieve executable last modification time if it exists
    if (doze___is_file_usable(real_path)) {
        ctx->exe.last_modified = doze___get_timestamp(real_path);
    } else {
        ctx->exe.last_modified = 0;
    }

    // we generate the dependency order
    for (u32 i = 0; i < ctx->nb_batches; i++) {
        if (ctx->bucket[i].nb_childs > 0) {
            if (doze___resolve_dependencies(ctx, i)) {
                return;
            }
        }
        // we silently ignore batches which are used by no one
    }

    // we traverse the resolved order and do what must be done.
    for (u32 i = 0; i < ctx->nb_resolved; i++) {
        doze___process_batch(ctx, ctx->resolved_cache[i]);
    }

    // if no files changed then no need to recompile
    if (!ctx->nb_modified_files) {
        puts("doze: nothing to do");
        return;
    }

    // now we link the final executable
    char *cmd = ZEE_MALLOC(1, 1);

    cmd = doze___merge(cmd, ctx->opts[DOZE_OPT_COMPILER_PATH].opt, 0x0);
    cmd = doze___merge(cmd, ctx->exe.path, " -o ");
    cmd = doze___merge(cmd, ctx->object_files, 0x0);

    if (ctx->opts[DOZE_OPT_LIBRARY_NAMES].opt) {
        DOZE___SPLIT_REPLACE_MERGE_INTO(
            ctx->opts[DOZE_OPT_LIBRARY_NAMES].opt,
            DOZE_OPTION_SEPARATOR, " -l ",
            cmd);
    }
    if (ctx->opts[DOZE_OPT_LIBRARY_PATHS].opt) {
        DOZE___SPLIT_REPLACE_MERGE_INTO(
            ctx->opts[DOZE_OPT_LIBRARY_PATHS].opt,
            DOZE_OPTION_SEPARATOR, " -L ",
            cmd);
    }
    if (ctx->opts[DOZE_OPT_ADDITIONAL].opt) {
        DOZE___SPLIT_REPLACE_MERGE_INTO(
            ctx->opts[DOZE_OPT_ADDITIONAL].opt,
            DOZE_OPTION_SEPARATOR, " ",
            cmd);
    }

    puts(cmd);
    system(cmd);

    if (ctx->nb_compiled_files) {
        printf("doze: %d file%s compiled\n",
            ctx->nb_compiled_files,
            ctx->nb_compiled_files > 1 ? "s" : "");
    }
    if (ctx->nb_compiled_files != ctx->nb_modified_files) {
        printf("doze: %d file%s modified\n",
            ctx->nb_modified_files,
            ctx->nb_modified_files > 1 ? "s" : "");
    }
    printf("doze: linked executable '%s'\n", ctx->exe.path);

    ZEE_FREE(cmd);
}

#endif /* DOZE_IMPLEMENTATION */
#endif /* ZEE_DOZE_H */
