#!/usr/bin/env bash
#
# Build (or refresh) a signed static APT repository from one or more .deb files.
#
# Layout produced under REPO_DIR:
#   KEY.gpg                                         public signing key (armored)
#   pool/main/f/files-signer/*.deb                  package pool (accumulates versions)
#   dists/stable/main/binary-amd64/Packages[.gz]    package index
#   dists/stable/Release, Release.gpg, InRelease     signed release metadata
#
# Usage:
#   build-repo.sh REPO_DIR GPG_KEY_ID DEB [DEB...]
#
# Requires: apt-utils (apt-ftparchive), gpg, gzip. The secret key for GPG_KEY_ID
# must already be imported in the current gpg keyring.
#
set -euo pipefail

REPO_DIR=${1:?"REPO_DIR required"}
GPG_KEY_ID=${2:?"GPG_KEY_ID required"}
shift 2
DEBS=("$@")
[ "${#DEBS[@]}" -gt 0 ] || { echo "error: no .deb files given" >&2; exit 1; }

SUITE=stable
COMPONENT=main
ARCH=amd64
ORIGIN="files-signer"
LABEL="files-signer"

POOL="pool/${COMPONENT}/f/files-signer"
BINDIR="dists/${SUITE}/${COMPONENT}/binary-${ARCH}"

mkdir -p "${REPO_DIR}/${POOL}" "${REPO_DIR}/${BINDIR}"

# 1. Copy the new .deb(s) into the pool. Old versions already there are kept, so
#    clients can still resolve/downgrade and updates work across releases.
for deb in "${DEBS[@]}"; do
  cp -f "${deb}" "${REPO_DIR}/${POOL}/"
done

# From here on, paths must be relative to REPO_DIR so the "Filename:" field in the
# index points at "pool/..." as the client expects.
cd "${REPO_DIR}"

# 2. Package index: scan the whole pool (keeps every version listed).
apt-ftparchive packages "pool" > "${BINDIR}/Packages"
gzip -9kf "${BINDIR}/Packages"

# 3. Release file with the repo metadata + checksums of the index.
apt-ftparchive \
  -o "APT::FTPArchive::Release::Origin=${ORIGIN}" \
  -o "APT::FTPArchive::Release::Label=${LABEL}" \
  -o "APT::FTPArchive::Release::Suite=${SUITE}" \
  -o "APT::FTPArchive::Release::Codename=${SUITE}" \
  -o "APT::FTPArchive::Release::Architectures=${ARCH}" \
  -o "APT::FTPArchive::Release::Components=${COMPONENT}" \
  release "dists/${SUITE}" > "dists/${SUITE}/Release"

# 4. Sign the Release: InRelease (inline/clearsigned) + Release.gpg (detached).
gpg --batch --yes --default-key "${GPG_KEY_ID}" \
  --clearsign -o "dists/${SUITE}/InRelease" "dists/${SUITE}/Release"
gpg --batch --yes --default-key "${GPG_KEY_ID}" \
  -abs -o "dists/${SUITE}/Release.gpg" "dists/${SUITE}/Release"

# 5. Publish the public key so users can verify the repo.
gpg --armor --export "${GPG_KEY_ID}" > "KEY.gpg"

echo "OK: APT repo built at ${REPO_DIR} (suite=${SUITE}, arch=${ARCH})"
