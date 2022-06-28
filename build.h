/*
 * Created by Costa Bushnaq
 *
 * 12-03-2022 @ 00:18:00
 *
 * see LICENSE
*/

#ifndef OG_BUILD_H_
#define OG_BUILD_H_

// CONFIGURATION

#define MAX_BUILD_NODES   2048
#define MAX_BATCH_SIZE    2048
#define READ_BUFFER_SIZE  2048

//----------------------------

#define OG_BUILD_VERSION "0.1.0"

#include <assert.h>
#include <stdarg.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>

#include <openssl/md5.h>

typedef int8_t   s8;  // char
typedef int16_t  s16;
typedef int32_t  s32; // int
typedef int64_t  s64; // ssize_t
typedef uint8_t  u8;
typedef uint16_t u16;
typedef uint32_t u32; // size_t
typedef uint64_t u64;

typedef struct {
    u32 nb_files;
    struct {
        char *path;
        u8 header;
        u8 digest[MD5_DIGEST_LENGTH];
    } files[MAX_BATCH_SIZE];
} Batch;

typedef struct {
    char *path;
} Binary;

enum { ND_BATCH, ND_BINARY };
typedef struct {
    u32 id;
    u16 type;
    void *data;
    u64 last_modification_stamp;
    u32 nb_dep_ids;
    u32 dep_ids[MAX_BUILD_NODES];
} Build_Node;

enum {
    // single-values
    OPT_COMPILER_BIN,
    // OPT_LINKER_BIN, @Unused
    OPT_SRC_PATH,
    OPT_BUILD_PATH,

    // multi-values
    OPT_INCLUDE_PATHS,
    OPT_LIBRARY_PATHS,
    OPT_LIBRARIES,
    OPT_ADDITIONAL,

    // internal
    OPT_END
};
typedef struct {
    char **options[OPT_END];
} Compiler_Options;

typedef struct {
    u32 nb_nodes;
    Build_Node *nodes[MAX_BUILD_NODES];
    Compiler_Options *opts;
} Build_Context;

#define BU_ASSERT_PERR(s, err) if (!(s)) bu_perror_abort(err);
#define BU_ASSERT_MSG(s, err) if (!(s)) bu_printf_abort(err);

// ---------------------------------------------------------------
// Prototypes
static u8 batch_compile(Build_Context *ctx, Build_Node *node, u8 force);
static u8 binary_link(Build_Context *ctx, Build_Node *node, u8 updated);

// ---------------------------------------------------------------
// Utils
static inline void bu_perror_abort(char const *err_msg)
{
    // @Robustnes Maybe also provide line number and/or file name.
    fprintf(stderr, "assert failed: ");
    perror(err_msg);
    abort();
}

static inline void bu_printf_abort(char const *err_msg)
{
    fprintf(stderr, "assert failed: %s\n", err_msg);
    abort();
}

static inline void *bu_zalloc(u64 nb_elems, u64 size_elems)
{
    void *data = calloc(nb_elems, size_elems);
    BU_ASSERT_PERR(data, "calloc")
    return data;
}

static inline void *bu_realloc(void *ptr, u64 new_size)
{
    void *data = realloc(ptr, new_size);
    BU_ASSERT_PERR(data, "realloc");
    return data;
}

static char **process_opt_string(char *str, char const *prefix)
{
    char **res = bu_zalloc(1, sizeof(char**));
    u64 prefix_length = prefix ? strlen(prefix) : 0;
    u32 nb_elems = 0;

    res[0] = 0x0;
    for (char *tok = strtok(str, ";"); tok; tok = strtok(0x0, ";")) {
        if (prefix_length) {
            res[nb_elems] = bu_zalloc(strlen(tok) + prefix_length + 1, 1);
            strcat(res[nb_elems], prefix);
        } else
            res[nb_elems] = bu_zalloc(strlen(tok) + 1, 1);
        strcat(res[nb_elems], tok);
        nb_elems += 1;
        res = bu_realloc(res, sizeof(char**) * (nb_elems + 1));
        res[nb_elems] = 0x0;
    }
    return res;
}

static void get_file_hash(const char *path, u8 *digest_buf)
{
    FILE *fp = fopen(path, "r"); BU_ASSERT_PERR(fp, "fopen");
    MD5_CTX ctx;
    u8  bytes_buf[READ_BUFFER_SIZE];
    u64 bytes_read;

    MD5_Init(&ctx);
    while ((bytes_read = fread(bytes_buf, 1, READ_BUFFER_SIZE, fp)) != 0)
        MD5_Update(&ctx, bytes_buf, bytes_read);
    MD5_Final(digest_buf, &ctx);
    fclose(fp);
}

static inline u8 is_file_header(const char *path)
{
    // We assume that `path` has been verified to end with an extension
    // and that it is a valid pointer.
    return path[strlen(path) - 1] == 'h';
}

static inline u8 is_file_correct(const char *path)
{
    if (!path) return 0;
    if (strlen(path) <= 2) return 0;
    if (access(path, R_OK)) return 0;
    if (path[strlen(path) - 2] != '.') return 0;
    if (path[strlen(path) - 1] != 'c')
        if (!is_file_header(path))
            return 0;
    return 1;
}

static inline char *merge(char *dest, char const *src, char const *sep)
{
    u64 destlen = strlen(dest);
    u64 srclen, seplen = 0;

    // We merge `src` into `dest` using optional separator `sep`;
    // reallocating if we need to.
    BU_ASSERT_MSG(src, "provide a valid (char*): src"); 
    srclen = strlen(src);
    if (sep)
        seplen = strlen(sep);
    dest = bu_realloc(dest, destlen + srclen + seplen + 1); // @Bug
    if (sep)
        dest = strcat(dest, sep);
    dest = strcat(dest, src);
    return dest;
}

static void source_compile(char const *path, Compiler_Options *opts)
{
    char *cmd = bu_zalloc(1, 1);

    // We assume that `opts` has been properly prepared using bake_options().
    // And that `path` is the path for a c source code file.
    cmd = merge(cmd, opts->options[OPT_COMPILER_BIN][0], 0x0);
    if (opts->options[OPT_SRC_PATH])
        cmd = merge(cmd, opts->options[OPT_SRC_PATH][0], " ");
    else // we need a space if we don't have a source path.
        cmd = merge(cmd, " ", 0x0);
    cmd = merge(cmd, path, 0x0);
    cmd = merge(cmd, " -c -o", 0x0);
    if (opts->options[OPT_BUILD_PATH])
        cmd = merge(cmd, opts->options[OPT_BUILD_PATH][0], " ");
    else // we need a space if we don't have a build path.
        cmd = merge(cmd, " ", 0x0);
    cmd = merge(cmd, path, 0x0);
    cmd[strlen(cmd) - 1] = 'o';
    if (opts->options[OPT_INCLUDE_PATHS])
        for (u32 i = 0; opts->options[OPT_INCLUDE_PATHS][i]; i++)
            cmd = merge(cmd, opts->options[OPT_INCLUDE_PATHS][i], " ");
    if (opts->options[OPT_ADDITIONAL])
        for (u32 i = 0; opts->options[OPT_ADDITIONAL][i]; i++)
            cmd = merge(cmd, opts->options[OPT_ADDITIONAL][i], " ");
    puts(cmd);
    system(cmd);
    free(cmd);
}

static char *batch_objects_merge(Batch *batch, Compiler_Options *opts, u32 *nb_ob)
{
    char *result = bu_zalloc(1, 1);
    *nb_ob = 0;

    for (u32 i = 0; i < batch->nb_files; i++) {
        if (is_file_header(batch->files[i].path))
            continue;
        if (opts->options[OPT_BUILD_PATH]) {
            result = merge(result, opts->options[OPT_BUILD_PATH][0], " ");
            result = merge(result, batch->files[i].path, 0x0);
        } else
            result = merge(result, batch->files[i].path, " ");
        result[strlen(result) - 1] = 'o';
        *nb_ob += 1;
    }
    return result;
}

// ---------------------------------------------------------------
// Build_Context
static Build_Context *make_context(Compiler_Options *opts)
{
    Build_Context *ctx = bu_zalloc(1, sizeof(Build_Context));

    if (opts)
        ctx->opts = opts;
    return ctx;
}

static Build_Context *set_options(Build_Context *ctx, Compiler_Options *opts)
{
    BU_ASSERT_MSG(opts, "provide a non-null (Compiler_Options*): opts");
    ctx->opts = opts;
    return ctx;
}

// ---------------------------------------------------------------
// Compiler_Options
// @Incomplete We would need to perform some checks in here to make it better;
// i.e, if OPT_BUILD_PATH && OPT_SRC_PATH end with '/'
static void bake_options(Compiler_Options *opts, u32 nb_options, ...)
{
    BU_ASSERT_MSG(opts, "provide a non-null (Compiler_Options*): opts");
    BU_ASSERT_MSG(nb_options > 0, "provide a non-zero (u32): nb_options");
    memset(opts, 0, sizeof(Compiler_Options));

    u32  type;
    char *str;
    va_list list_opts;
    static char const *prefixes[OPT_END] = {
        [OPT_INCLUDE_PATHS] = "-I",
        [OPT_LIBRARY_PATHS] = "-L",
        [OPT_LIBRARIES]     = "-l",
    };

    va_start(list_opts, nb_options);
    for (; nb_options > 0; nb_options--) {
        type = va_arg(list_opts, u32);
        str  = va_arg(list_opts, char*);
        switch (type) {
        case OPT_COMPILER_BIN:
        case OPT_SRC_PATH:
        case OPT_BUILD_PATH:
            // @Cleanup It's kinda dirty to make an array just for a single value
            opts->options[type] = process_opt_string(str, 0x0);
            break;
        case OPT_INCLUDE_PATHS:
        case OPT_LIBRARY_PATHS:
        case OPT_LIBRARIES:
            opts->options[type] = process_opt_string(str, prefixes[type]);
            break;
        case OPT_ADDITIONAL:
            opts->options[type] = process_opt_string(str, 0x0);
            break;
        default:
            BU_ASSERT_MSG(type <= OPT_END, "provide suitable values to bake_options()");
        }
    }
    va_end(list_opts);
}

// ---------------------------------------------------------------
// Build_Node
static Build_Node *make_node(Build_Context *ctx, u16 type, void *data)
{
    BU_ASSERT_MSG(ctx->nb_nodes < MAX_BUILD_NODES,
                  "MAX_BUILD_NODES limit reached: too many nodes");
    Build_Node *node = bu_zalloc(1, sizeof(Build_Node));

    node->id = ctx->nb_nodes;
    node->type = type;
    node->data = data;
    node->last_modification_stamp = 0; // just a reminder

    ctx->nodes[ctx->nb_nodes] = node;
    ctx->nb_nodes += 1;
    return node;
}

static void node_add_dependency(Build_Context *ctx, u32 self_id, u32 dep_id)
{
    // @NoCheckin If adding same dependency twice
    // @NoCheckin Also not checking for cirular dependency
    Build_Node *node = ctx->nodes[self_id];
    BU_ASSERT_MSG(node, "provide a valid node id (u32): self_id");
    BU_ASSERT_MSG(node->nb_dep_ids < MAX_BUILD_NODES,
                  "MAX_BUILD_NODES limit reached: too many dependencies");
    
    node->dep_ids[node->nb_dep_ids] = dep_id;
    node->nb_dep_ids += 1;
}

static u8 node_compile(Build_Context *ctx, u32 self_id, u64 const current_stamp)
{
    Build_Node *node = ctx->nodes[self_id];
    u8 updated = 0;

    // First, we run node_compile() for each dependency. Recompilation will 
    // be executed if needed and if such, updated is changed to 1.
    for (u32 i = 0; i < node->nb_dep_ids; i++) {
        if (node_compile(ctx, node->dep_ids[i], current_stamp))
            updated = 1;
    }

    // Second, we return if these conditions are met.
    // 1. no dependencies were updated,
    // 2. there is a saved last_modification_stamp
    // 3. the current timestamp is inferior or equal to the last saved timestamp
    if (!updated && node->last_modification_stamp &&
        node->last_modification_stamp >= current_stamp)
        return 0;

    // Finally, we can just go ahead with compiling the node.
    switch (node->type) {
        case ND_BATCH:
            updated = batch_compile(ctx, node, updated);
            break;
        case ND_BINARY:
            updated = binary_link(ctx, node, updated);
            break;
        default:
            BU_ASSERT_MSG(0, "invalid node->type found");
    }
    node->last_modification_stamp = current_stamp;
    return updated;
}

// ---------------------------------------------------------------
// Batch
static Build_Node *make_batch(Build_Context *ctx, char *files[])
{
    Batch *batch = bu_zalloc(1, sizeof(Batch));
    Compiler_Options *opts = ctx->opts;

    batch->nb_files = 0;
    for (u32 i = 0; i < MAX_BATCH_SIZE && files[i]; i++) {
        char *path = bu_zalloc(1, 1);
        BU_ASSERT_MSG(is_file_correct(files[i]), "provide accessible c source code files or headers");
        if (opts->options[OPT_SRC_PATH])
            path = merge(path, opts->options[OPT_SRC_PATH][0], 0x0);
        path = merge(path, files[i], 0x0);
        batch->files[i].path   = path;
        batch->files[i].header = is_file_header(path);
        batch->nb_files += 1;
    }
    return make_node(ctx, ND_BATCH, (void*)batch);
}

static u8 batch_compile(Build_Context *ctx, Build_Node *node, u8 force)
{
    Batch *batch = (Batch*)node->data;
    u8 buf[MD5_DIGEST_LENGTH];
    u32 nb_updates = 0;
    u32 nb_compiles = 0;
    s32 cmp_res;

    // For each file in the batch, we check if its current hash is the same
    // as the one we have stored. 
    for (u32 i = 0; i < batch->nb_files; i++) {
        get_file_hash(batch->files[i].path, buf);
        cmp_res = memcmp(batch->files[i].digest, buf, MD5_DIGEST_LENGTH);
        if (force || cmp_res) {
            // If dependencies were updated or the digests don't match, we
            // (re)compile. We update the digest if it is needed.
            if (cmp_res) {
                memcpy(batch->files[i].digest, buf, MD5_DIGEST_LENGTH);
                nb_updates += 1;
            }
            if (!batch->files[i].header) {
                source_compile(batch->files[i].path, ctx->opts);
                nb_compiles += 1;
            }
        }
    }
    if (nb_updates == 0 && nb_compiles == 0)
        return 0;
    printf("batch[%d]: %d file%s modified, %d file%s compiled.\n",
        node->id,
        nb_updates, nb_updates == 1 ? "" : "s",
        nb_compiles, nb_compiles == 1 ? "" : "s"
    );
    return nb_updates + nb_compiles > 0 ? 1 : 0;
}

// ---------------------------------------------------------------
// Binary
static Build_Node *make_binary(Build_Context *ctx, char const *name)
{
    Binary *binary = bu_zalloc(1, sizeof(Binary));
    Compiler_Options *opts = ctx->opts;
    char *path = bu_zalloc(1, 1);

    if (opts->options[OPT_BUILD_PATH])
        path = merge(path, opts->options[OPT_BUILD_PATH][0], 0x0);
    path = merge(path, name, 0x0);
    binary->path = path;
    return make_node(ctx, ND_BINARY, (void*)binary);
}

static u8 binary_link(Build_Context *ctx, Build_Node *node, u8 updated)
{
    Binary *binary = (Binary*)node->data;
    Compiler_Options *opts = ctx->opts;
    char *cmd = bu_zalloc(1, 1);
    char *buf;
    u32 nb_object_files = 0;

    // If no dependencies were updated then return.
    if (!updated) {
        printf("bin  [%d]: Nothing to do.\n", node->id);
        free(cmd);
        return 0;
    }

    cmd = merge(cmd, opts->options[OPT_COMPILER_BIN][0], 0x0);
    cmd = merge(cmd, binary->path, " -o");
    for (u32 i = 0; i < node->nb_dep_ids; i++) {
        u32 dep_id = node->dep_ids[i];
        u32 nb_ob;
        if (ctx->nodes[dep_id]->type == ND_BATCH) {
            buf = batch_objects_merge((Batch*)ctx->nodes[dep_id]->data,
                                      opts, &nb_ob);
            if (nb_ob) {
                cmd = merge(cmd, buf, 0x0);
                nb_object_files += nb_ob;
            }
            free(buf);
        }
    }
    if (opts->options[OPT_INCLUDE_PATHS])
        for (u32 i = 0; opts->options[OPT_INCLUDE_PATHS][i]; i++)
            cmd = merge(cmd, opts->options[OPT_INCLUDE_PATHS][i], " ");
    if (opts->options[OPT_LIBRARY_PATHS])
        for (u32 i = 0; opts->options[OPT_LIBRARY_PATHS][i]; i++)
            cmd = merge(cmd, opts->options[OPT_LIBRARY_PATHS][i], " ");
    if (opts->options[OPT_LIBRARIES])
        for (u32 i = 0; opts->options[OPT_LIBRARIES][i]; i++)
            cmd = merge(cmd, opts->options[OPT_LIBRARIES][i], " ");
    if (opts->options[OPT_ADDITIONAL])
        for (u32 i = 0; opts->options[OPT_ADDITIONAL][i]; i++)
            cmd = merge(cmd, opts->options[OPT_ADDITIONAL][i], " ");
    puts(cmd);
    system(cmd);
    free(cmd);
    printf("bin  [%d]: %d object file%s linked.\n",
        node->id,
        nb_object_files, nb_object_files == 1 ? "" : "s"
    );
    return 1;
}

#endif /* OG_BUILD_H_ */
