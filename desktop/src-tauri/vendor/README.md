# GLib 0.18 security backport

`glib/` contains the complete crates.io `glib` 0.18.5 source archive, with
one change: the mutable output-pointer fix from
[gtk-rs/gtk-rs-core#1343](https://github.com/gtk-rs/gtk-rs-core/pull/1343).
It addresses [RUSTSEC-2024-0429](https://rustsec.org/advisories/RUSTSEC-2024-0429)
in `VariantStrIter::impl_get`. The pointer passed to `g_variant_get_child`
must be `&mut p`; writing through `&p` is undefined behavior.

The original archive is
<https://static.crates.io/crates/glib/glib-0.18.5.crate>, SHA-256:
`233daaf6e83ae6a12a52055f568f9d7cf4671dabb78ff9560ab6da230ce00ee5`.
All other files are unchanged, including upstream copyright and MIT license.
The package keeps its upstream version; Cargo's path patch selects this source.

Tauri's GTK 3 backend requires GLib 0.18; GLib 0.20 is not a compatible
replacement. The upstream 0.18 backport request
[#2009](https://github.com/gtk-rs/gtk-rs-core/pull/2009) was closed without
merging. Remove this copy when a compatible fixed dependency is available.

`scripts/verify-glib-backport.py` checks every vendored file against the
checksum-verified original archive, allowing only the exact upstream fix.
The Linux desktop workflow also runs an optimized regression test through
the app's resolved GLib dependency; debug-only testing can miss this bug.

Version-only scanners may still report the upstream advisory. Do not rename
the package or change its version to hide the finding. Review the source
verification and regression test alongside any advisory report.
