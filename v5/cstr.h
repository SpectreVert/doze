/*
 * Created by Costa Bushnaq
 *
 * 05-09-2023 @ 13:21:56
 *
 * see LICENSE
*/

#ifndef CSTR_H
#define CSTR_H

#include <string.h>
#include <stdint.h>

#ifndef CSTR_DEF
#ifdef CSTR_STATIC
#define CSTR_DEF static
#else
#define CSTR_DEF extern
#endif
#endif

#ifndef CSTR_MAX_PREALLOCATE
#define CSTR_MAX_PREALLOCATE (1024 * 1024)
#endif

typedef char* cstr;

CSTR_DEF uint64_t cstr_len(const cstr);
CSTR_DEF uint64_t cstr_cap(const cstr);
CSTR_DEF uint64_t cstr_avail(const cstr);

CSTR_DEF cstr cstr_new(void);
CSTR_DEF cstr cstr_dup(const cstr);
CSTR_DEF cstr cstr_from(char const*);
CSTR_DEF cstr cstr_from_s(char const*,uint64_t);

CSTR_DEF cstr cstr_concat(cstr,char const*);
CSTR_DEF cstr cstr_concat_s(cstr,char const*,uint64_t);

CSTR_DEF cstr cstr_take_in(cstr,uint64_t);

CSTR_DEF void cstr_free(cstr);

#ifdef CSTR_IMPLEMENTATION

#include <assert.h>
#include <stdlib.h>

#ifndef CSTR_MALLOC
#define CSTR_MALLOC(s) malloc(s)
#define CSTR_FREE(p) free(p)
#define CSTR_REALLOC(p, s) realloc(p, s)
#endif

struct __attribute__((__packed__)) cstr_header_8 {
    /* max cap: 255 */
    uint8_t len;
    uint8_t cap;
    unsigned char flags;
    char data[];
};

struct __attribute__((__packed__)) cstr_header_16 {
    /* max cap: 65'535 */
    uint16_t len;
    uint16_t cap;
    unsigned char flags;
    char data[];
};

struct __attribute__((__packed__)) cstr_header_32 {
    /* max cap: 4'294'967'295 */
    uint32_t len;
    uint32_t cap;
    unsigned char flags;
    char data[];
};

/* We are using the 2 least significant bits for the type. Remains 6 unused. */
#define CSTR_TYPE_8  1
#define CSTR_TYPE_16 2
#define CSTR_TYPE_32 3
#define CSTR_TYPE_MASK 3

#define CSTR_HEADER(s, t) \
    ((struct cstr_header_##t*)((s)-(sizeof(struct cstr_header_##t)))) \

/* PRIVATE */

static inline char
cstr___define_type(
    uint64_t size
) {
    if (size < 1<<8) { return CSTR_TYPE_8;  }
    if (size < 1<<16) { return CSTR_TYPE_16; }
    if (size < 1ll<<32) { return CSTR_TYPE_32; }
    assert(0); /* @Implement CSTR_TYPE_64 */
}

static inline int
cstr___get_header_size(
    char type
) {
    switch (type) {
    case CSTR_TYPE_8: return sizeof(struct cstr_header_8);
    case CSTR_TYPE_16: return sizeof(struct cstr_header_16);
    case CSTR_TYPE_32: return sizeof(struct cstr_header_32);
    } return 0;
}

static inline void
cstr___set_len(
    cstr s,
    uint64_t new_len
) {
    unsigned char flags = s[-1];

    switch (flags & CSTR_TYPE_MASK) {
    case CSTR_TYPE_8:  CSTR_HEADER(s, 8)->len = new_len; break;
    case CSTR_TYPE_16: CSTR_HEADER(s, 16)->len = new_len; break;
    case CSTR_TYPE_32: CSTR_HEADER(s, 32)->len = new_len; break;
    }
}

static inline void
cstr___set_cap(
    cstr s,
    uint64_t new_cap
) {
    unsigned char flags = s[-1];

    switch (flags & CSTR_TYPE_MASK) {
    case CSTR_TYPE_8:  CSTR_HEADER(s, 8)->cap = new_cap; break;
    case CSTR_TYPE_16: CSTR_HEADER(s, 16)->cap = new_cap; break;
    case CSTR_TYPE_32: CSTR_HEADER(s, 32)->cap = new_cap; break;
    }
}

/* PUBLIC */

CSTR_DEF uint64_t
cstr_len(
    const cstr s
) {
    unsigned char flags = s[-1];

    switch (flags & CSTR_TYPE_MASK) {
    case CSTR_TYPE_8:  return CSTR_HEADER(s, 8)->len;
    case CSTR_TYPE_16: return CSTR_HEADER(s, 16)->len;
    case CSTR_TYPE_32: return CSTR_HEADER(s, 32)->len;
    } return 0;
}

CSTR_DEF uint64_t
cstr_cap(
    const cstr s
) {
    unsigned char flags = s[-1];

    switch (flags & CSTR_TYPE_MASK) {
    case CSTR_TYPE_8:  return CSTR_HEADER(s, 8)->cap;
    case CSTR_TYPE_16: return CSTR_HEADER(s, 16)->cap;
    case CSTR_TYPE_32: return CSTR_HEADER(s, 32)->cap;
    } return 0;
}

CSTR_DEF uint64_t
cstr_avail(
    const cstr s
) {
    unsigned char flags = s[-1];

    switch (flags & CSTR_TYPE_MASK) {
    case CSTR_TYPE_8:
        return CSTR_HEADER(s, 8)->cap - CSTR_HEADER(s, 8)->len;
    case CSTR_TYPE_16:
        return CSTR_HEADER(s, 16)->cap - CSTR_HEADER(s, 16)->len;
    case CSTR_TYPE_32:
        return CSTR_HEADER(s, 32)->cap - CSTR_HEADER(s, 32)->len;
    } return 0;
}

CSTR_DEF cstr
cstr_new(
    void
) {
    return cstr_from(0x0);
}

CSTR_DEF cstr
cstr_dup(
    const cstr s
) {
    return cstr_from_s(s, cstr_len(s));
}

CSTR_DEF cstr
cstr_from(
    char const *o
) {
    uint64_t o_len = 0;

    if (o) { o_len = strlen(o); }
    return cstr_from_s(o, o_len);
}

CSTR_DEF cstr
cstr_from_s(
    char const *o,
    uint64_t o_len
) {
    char type;
    int header_size;
    unsigned char *flags;
    void *base;
    cstr s;

    type = cstr___define_type(o_len);
    header_size = cstr___get_header_size(type);

    base = CSTR_MALLOC(header_size + o_len + 1); /* null byte */
    if (!base) { return 0x0; }
    if (!o && o_len) { memset(base, 0, header_size + o_len + 1); }

    s = (cstr)base + header_size;
    flags = ((unsigned char*)s) - 1;

    switch (type) {
    case CSTR_TYPE_8: {
        struct cstr_header_8 *header = base;
        header->len = header->cap = o_len;
        *flags = type;
        break;
    }
    case CSTR_TYPE_16: {
        struct cstr_header_16 *header = base;
        header->len = header->cap = o_len;
        *flags = type;
        break;
    }
    case CSTR_TYPE_32: {
        struct cstr_header_32 *header = base;
        header->len = header->cap = o_len;
        *flags = type;
        break;
    }
    }

    if (o && o_len) { memcpy(s, o, o_len); }
    s[o_len] = 0;
    return s;
}

CSTR_DEF cstr
cstr_concat(
    cstr c,
    char const *o
) {
    return cstr_concat_s(c, o, strlen(o));
}

CSTR_DEF cstr
cstr_concat_s(
    cstr c,
    char const *o,
    uint64_t o_len
) {
    uint64_t len = cstr_len(c);

    c = cstr_take_in(c, o_len);
    if (!c) { return 0x0; }
    memcpy(c + len, o, o_len);
    cstr___set_len(c, len + o_len);
    c[len + o_len] = 0;
    return c;
}

CSTR_DEF cstr
cstr_take_in(
    cstr c,
    uint64_t r
) {
    /* The user notifies that he plans taking in `r` bytes more than the
     * cstr already contains. Our job is to make sure we have enough room
     * to prevent futher allocations.
    */
    char type, new_type;
    uint64_t len, new_cap;
    int new_header_size;
    void *base, *new_base;
    cstr s;

    /* return immediately if we have enough space available */
    if (cstr_avail(c) >= r) { return c; }

    /* We will allocate more bytes than requested to reduce the number of
     * malloc calls, but not more than CSTR_MAX_PREALLOCATE.
    */
    len = cstr_len(c);
    new_cap = len + r;
    if (new_cap < CSTR_MAX_PREALLOCATE) {
        new_cap = new_cap * 2;
    } else {
        new_cap += CSTR_MAX_PREALLOCATE;
    }

    type = (((unsigned char)c[-1]) & CSTR_TYPE_MASK);
    base = c - cstr___get_header_size(type);

    new_type = cstr___define_type(new_cap);
    new_header_size = cstr___get_header_size(new_type);

    if (new_type == type) {
        /* the allocation fits in our current cstr category */
        new_base = CSTR_REALLOC(base, new_header_size + new_cap + 1);
        if (!new_base) { return 0x0; }
        s = (cstr)new_base + new_header_size;
    } else {
        /* the allocation will not fit in our current cstr category, so we
         * remake one */
        new_base = CSTR_MALLOC(new_header_size + new_cap + 1);
        if (!new_base) { return 0x0; }
        s = (cstr)new_base + new_header_size;
        memcpy(s, c, len + 1); /* copy the old string into the new one */
        CSTR_FREE(base);
        s[-1] = new_type;
        cstr___set_len(s, len);
    }
    cstr___set_cap(s, new_cap);
    return s;
}

CSTR_DEF void
cstr_free(
    cstr c
) {
    if (!c) { return; }
    CSTR_FREE(c - cstr___get_header_size(c[-1]));
}

#endif /* CSTR_IMPLEMENTATION */
#endif /* CSTR_H */
