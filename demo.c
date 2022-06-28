#include "build.h"

char *src_files[] = {
    "main.c",
    0x0
};

char *header_files[] = {
    "build.h",
    "include/header.h",
    0x0
};

int main(int ac, char *av[]) {
    Build_Context *ctx = make_context(0x0);
    Compiler_Options opts;
    
    bake_options(&opts, 3,
        OPT_COMPILER_BIN,  "tcc",
        OPT_BUILD_PATH,    "build/",
        OPT_LIBRARIES,     "ssl;ncurses"
    );
    set_options(ctx, &opts);
    
    Build_Node *hds = make_batch(ctx, header_files);
    Build_Node *src = make_batch(ctx, src_files);
    Build_Node *bin = make_binary(ctx, "b.out");

    node_add_dependency(ctx, src->id, hds->id);
    node_add_dependency(ctx, bin->id, src->id);

    char buf[1];
    u32 stamp = 1;
    while (1) {
        node_compile(ctx, bin->id, stamp++);
        read(0, buf, 1);
    }

    return 0;
}
