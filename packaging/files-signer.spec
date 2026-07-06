# RPM spec for COPR / Fedora.
# Build a source tarball named files-signer-%{version}.tar.gz at the repo root
# (ideally with vendored deps: `go mod vendor` committed, so mock builds offline).
Name:           files-signer
Version:        0.1.0
Release:        1%{?dist}
Summary:        Sign, verify and extract files with PKCS#7/CMS digital signatures

License:        MIT
URL:            https://github.com/rquispe/files-signer
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.21
BuildRequires:  gcc
BuildRequires:  pkgconfig(gl)
BuildRequires:  pkgconfig(x11)
BuildRequires:  pkgconfig(xcursor)
BuildRequires:  pkgconfig(xrandr)
BuildRequires:  pkgconfig(xinerama)
BuildRequires:  pkgconfig(xi)
BuildRequires:  pkgconfig(xxf86vm)
BuildRequires:  desktop-file-utils
BuildRequires:  libappstream-glib

Requires:       zenity

%description
files-signer signs and verifies files of any type using standard PKCS#7/CMS
digital signatures and a PEM certificate. Cross-platform alternative to
XolidoSign, with a desktop app and a command-line interface. Supports attached
(.p7m) and detached (.p7s) signatures, optional trust-chain validation, and
recovering the original file from an attached signature.

%prep
%autosetup

%build
export CGO_ENABLED=1
# Use vendored modules when present (offline mock builds).
%if 0%{?_with_vendor:1}
export GOFLAGS=-mod=vendor
%endif
go build -o files-signer ./cmd/files-signer
go build -o files-signer-gui ./cmd/files-signer-gui

%install
install -Dm0755 files-signer        %{buildroot}%{_bindir}/files-signer
install -Dm0755 files-signer-gui     %{buildroot}%{_bindir}/files-signer-gui
install -Dm0644 packaging/com.rquispe.filessigner.desktop \
        %{buildroot}%{_datadir}/applications/com.rquispe.filessigner.desktop
install -Dm0644 packaging/com.rquispe.filessigner.metainfo.xml \
        %{buildroot}%{_metainfodir}/com.rquispe.filessigner.metainfo.xml
install -Dm0644 assets/icon-512.png %{buildroot}%{_datadir}/icons/hicolor/512x512/apps/com.rquispe.filessigner.png
install -Dm0644 assets/icon-256.png %{buildroot}%{_datadir}/icons/hicolor/256x256/apps/com.rquispe.filessigner.png
install -Dm0644 assets/icon-128.png %{buildroot}%{_datadir}/icons/hicolor/128x128/apps/com.rquispe.filessigner.png
install -Dm0644 assets/icon.svg     %{buildroot}%{_datadir}/icons/hicolor/scalable/apps/com.rquispe.filessigner.svg

%check
desktop-file-validate %{buildroot}%{_datadir}/applications/com.rquispe.filessigner.desktop
appstream-util validate-relax --nonet %{buildroot}%{_metainfodir}/com.rquispe.filessigner.metainfo.xml

%files
%license LICENSE
%doc README.md MANUAL.md
%{_bindir}/files-signer
%{_bindir}/files-signer-gui
%{_datadir}/applications/com.rquispe.filessigner.desktop
%{_metainfodir}/com.rquispe.filessigner.metainfo.xml
%{_datadir}/icons/hicolor/*/apps/com.rquispe.filessigner.*

%changelog
* Fri Jul 03 2026 Ruddy Quispe <development.rquispe@gmail.com> - 0.1.0-1
- Initial package: CLI + desktop app to sign, verify and extract files.
