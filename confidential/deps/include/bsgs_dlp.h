#ifndef BSGS_DLP_H
#define BSGS_DLP_H

#include <secp256k1.h>
#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @file bsgs_dlp.h
 * @brief Baby-Step Giant-Step discrete logarithm solver for EC-ElGamal decryption.
 *
 * Provides a BSGS solver over secp256k1 for recovering a 64-bit plaintext m
 * from the unmasked point M = m*G, as required by secp256k1_elgamal_decrypt_bsgs().
 *
 * @details
 * **Algorithm:**
 * Standard BSGS with a cuckoo hash (k=3) baby table and windowed TreeMon
 * batch inversion for the giant step loop, giving O(sqrt(N)) time and
 * O(sqrt(N)) space for a search range of [0, 2^bits_total).
 *
 * **Lifecycle:**
 * Building the baby table is expensive (seconds to minutes depending on l1).
 * Callers must initialize a secp256k1_elgamal_bsgs_ctx once — at process
 * startup or wallet initialization — and reuse it across all decrypt calls.
 * The context is read-only after initialization and safe to use from a
 * single thread. Multi-threaded use is not supported in this version.
 *
 * **Baby table cache:**
 * If cache_path is non-NULL, the baby table is saved to / loaded from a
 * binary file on disk. On a cache hit the init call returns in milliseconds.
 * Pass NULL to disable caching (table rebuilt on every process start).
 *
 * The cache format is host-specific: the header and entries are written in
 * native byte order and depend on struct layout, so a cache file is not
 * portable across architectures or compilers. On load it is validated (magic,
 * version, and the l1-derived geometry) and any mismatch fails closed — the
 * table is simply rebuilt — so a stale or foreign cache is safe, never wrong.
 * Treat cache files as a per-host build artifact; do not share them between
 * machines.
 *
 * @warning Always pass a non-NULL cache_path in any long-running or
 * repeatedly-started application. At default parameters (l1=22) the baby
 * table takes several seconds to build and ~22 MB of memory. Without a
 * cache file it is rebuilt from scratch on every process start, which is
 * unacceptable for interactive applications. The caller is responsible for
 * choosing an appropriate path (e.g. ~/.cache/myapp/bsgs_baby.bin) and
 * ensuring the directory exists.
 *
 * **Range parameters:**
 * bits_total controls the maximum recoverable plaintext: m < 2^bits_total.
 * l1 controls the baby-step count (baby table stores 2^(l1-1)+1 entries).
 * Use the defaults MPT_BSGS_DEFAULT_BITS_TOTAL / MPT_BSGS_DEFAULT_L1 unless
 * there is a specific reason to deviate.
 *
 * Memory at default parameters (bits_total=40, l1=22):
 *   Baby table: 2^21 entries × 8 bytes ≈ 16 MB (stored entries);
 *   ~22 MB allocated due to cuckoo hash overhead (~1.3× load factor).
 *
 * **Timing:**
 * This solver is NOT constant-time. The giant-step loop exits as soon as
 * a match is found, so decryption time leaks information about the
 * plaintext m (smaller values decrypt faster). It must only be used in
 * an environment where the caller has full control over the execution
 * context — such as a local tool or wallet running on the key owner's
 * own machine. It must not be used in server-side or shared
 * infrastructure where a third party could measure response times and
 * infer the plaintext.
 *
 * @see secp256k1_elgamal_decrypt for the fixed-range, timing-resistant solver.
 */

/**
 * Default range parameters.
 *
 * bits_total=40, l1=22: baby table ≈ 22 MB allocated (16 MB entries + cuckoo overhead), giant steps
 * = 2^18. Covers amounts up to ~1 trillion drops (2^40 ≈ 1.1e12). Suitable for most use cases.
 * Raise both values if the application needs to cover a larger supply, e.g., bits_total=54, l1=27 →
 * baby table ≈ 512 MB, covers 2^54 values.
 */
#define MPT_BSGS_DEFAULT_BITS_TOTAL 40
#define MPT_BSGS_DEFAULT_L1 22

/**
 * Default window size for the windowed TreeMon batch inversion in the
 * giant-step loop. Must be a power of 2. Larger values amortize field
 * inversions more aggressively at the cost of more stack allocation.
 */
#define MPT_BSGS_DEFAULT_WINDOW 128

/**
 * Maximum accepted window. secp256k1_elgamal_decrypt_bsgs() rejects any window
 * outside [1, MPT_BSGS_MAX_WINDOW]: the solver allocates O(window) points and
 * field elements, so an unbounded value risks overflow in the allocation-size
 * computation. This cap is far above any useful window (default is 128).
 */
#define MPT_BSGS_MAX_WINDOW 65536

/** Opaque BSGS context. Allocated by secp256k1_elgamal_bsgs_ctx_create(). */
typedef struct secp256k1_elgamal_bsgs_ctx secp256k1_elgamal_bsgs_ctx;

/**
 * @brief Allocates and initializes a BSGS context.
 *
 * Builds (or loads from cache) the baby-step hash table. This is the
 * expensive one-time setup call; all subsequent decrypt calls are fast.
 *
 * @param[in] ctx         A secp256k1 context with at least VERIFY capability.
 *                        The caller retains ownership; the context must remain
 *                        valid for the lifetime of the returned bsgs_ctx.
 * @param[in] bits_total  Search range: m in [0, 2^bits_total). Use
 *                        MPT_BSGS_DEFAULT_BITS_TOTAL.
 * @param[in] l1          Baby-step bit width. Baby table stores 2^(l1-1)+1
 *                        entries. Use MPT_BSGS_DEFAULT_L1.
 * @param[in] cache_path  Path for the binary baby table cache file, or NULL
 *                        to skip disk caching. If the file exists and is
 *                        valid it is loaded; otherwise the table is built
 *                        and saved to this path.
 *
 * @return Pointer to an initialized context, or NULL on failure (OOM,
 *         invalid parameters, or cache I/O error with no fallback).
 */
secp256k1_elgamal_bsgs_ctx*
secp256k1_elgamal_bsgs_ctx_create(
    secp256k1_context const* ctx,
    int bits_total,
    int l1,
    char const* cache_path);

/**
 * @brief Frees a BSGS context.
 *
 * Safe to call with NULL. The secp256k1_context passed at creation is
 * not freed; the caller remains responsible for it.
 *
 * @param[in] bsgs_ctx  Context to free.
 */
void
secp256k1_elgamal_bsgs_ctx_destroy(secp256k1_elgamal_bsgs_ctx* bsgs_ctx);

/**
 * @brief Decrypts an ElGamal ciphertext using BSGS for large-range decryption.
 *
 * Recovers plaintext amounts in [0, 2^bits_total) where bits_total was set
 * at context creation. This is the large-range replacement for
 * secp256k1_elgamal_decrypt, which is capped at 1,000,000.
 *
 * The m=0 case (C2 == sk*C1) is detected with a constant-time comparison
 * before the BSGS solver is invoked; the solver itself is not constant-time.
 *
 * @param[in]  ctx       A secp256k1 context.
 * @param[in]  bsgs_ctx  An initialized BSGS context.
 * @param[out] amount    Receives the recovered plaintext on success.
 * @param[in]  c1        First ciphertext component.
 * @param[in]  c2        Second ciphertext component.
 * @param[in]  privkey   32-byte private key scalar.
 * @param[in]  window    Window size for batch inversion. Must be in
 *                       [1, MPT_BSGS_MAX_WINDOW]; values outside that range
 *                       return 0. Internally rounded up to a power of 2.
 *                       Use MPT_BSGS_DEFAULT_WINDOW.
 *
 * @return 1 if the amount was successfully recovered and *amount is set;
 *         0 if the window is out of range, the plaintext is out of range, or
 *         an internal error occurred.
 */
SECP256K1_API int
secp256k1_elgamal_decrypt_bsgs(
    secp256k1_context const* ctx,
    secp256k1_elgamal_bsgs_ctx const* bsgs_ctx,
    uint64_t* amount,
    secp256k1_pubkey const* c1,
    secp256k1_pubkey const* c2,
    unsigned char const* privkey,
    int window);

#ifdef __cplusplus
}
#endif

#endif /* BSGS_DLP_H */
