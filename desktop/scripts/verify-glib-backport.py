"""Verify the GLib source copy against crates.io plus the exact upstream fix."""

import argparse
import hashlib
import io
import pathlib
import tarfile
import tomllib
import urllib.request

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--archive", type=pathlib.Path, help="Use an already downloaded original crate")
args = parser.parse_args()
if args.archive:
    data = args.archive.read_bytes()
else:
    with urllib.request.urlopen(
        "https://static.crates.io/crates/glib/glib-0.18.5.crate", timeout=30
    ) as response:
        data = response.read(4 * 1024 * 1024)
if hashlib.sha256(data).hexdigest() != "233daaf6e83ae6a12a52055f568f9d7cf4671dabb78ff9560ab6da230ce00ee5":
    raise SystemExit("Original GLib archive checksum mismatch")

original = b"""            let p: *mut libc::c_char = std::ptr::null_mut();
            let s = b"&s\\0";
            ffi::g_variant_get_child(
                self.variant.to_glib_none().0,
                i,
                s as *const u8 as *const _,
                &p,"""
patched = original.replace(b"let p:", b"let mut p:").replace(b"                &p,", b"                &mut p,")
root = pathlib.Path(__file__).resolve().parents[1] / "src-tauri/vendor/glib"
manifest = tomllib.loads((root.parents[1] / "Cargo.toml").read_text())
lock = tomllib.loads((root.parents[1] / "Cargo.lock").read_text())
if manifest.get("patch", {}).get("crates-io", {}).get("glib") != {"path": "vendor/glib"}:
    raise SystemExit("Desktop does not select the verified GLib source")
resolved = [package for package in lock["package"] if package["name"] == "glib"]
if len(resolved) != 1 or resolved[0]["version"] != "0.18.5" or "source" in resolved[0]:
    raise SystemExit("Desktop lock does not resolve the local GLib backport")
expected = set()
with tarfile.open(fileobj=io.BytesIO(data)) as archive:
    for member in archive.getmembers():
        if not member.isfile():
            raise SystemExit(f"Unexpected archive entry: {member.name}")
        relative = pathlib.PurePosixPath(member.name).relative_to("glib-0.18.5")
        if ".." in relative.parts:
            raise SystemExit("Unsafe archive path")
        name = relative.as_posix()
        expected.add(name)
        content = archive.extractfile(member).read()
        if name == "src/variant_iter.rs":
            if content.count(original) != 1:
                raise SystemExit("Upstream patch context mismatch")
            content = content.replace(original, patched)
        target = root / name
        if target.is_symlink() or not target.is_file() or target.read_bytes() != content:
            raise SystemExit(f"Vendored source differs: {name}")
actual = {p.relative_to(root).as_posix() for p in root.rglob("*") if p.is_file()}
if actual != expected:
    raise SystemExit(f"Unexpected source files: {sorted(actual ^ expected)}")
print(f"Verified {len(expected)} GLib files; only the upstream VariantStrIter fix differs.")
