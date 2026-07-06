# RPM spec for COPR / Fedora.
# Build a source tarball named files-sign-%{version}.tar.gz at the repo root
# (ideally with vendored deps: `go mod vendor` committed, so mock builds offline).
Name:           files-sign
Version:        0.1.0
Release:        1%{?dist}
Summary:        Sign, verify and extract files with PKCS#7/CMS digital signatures

License:        MIT
URL:            https://github.com/rquispe/files-sign
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
files-sign signs and verifies files of any type using standard PKCS#7/CMS
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
go build -o files-sign ./cmd/files-sign
go build -o files-sign-gui ./cmd/files-sign-gui

%install
install -Dm0755 files-sign        %{buildroot}%{_bindir}/files-sign
install -Dm0755 files-sign-gui     %{buildroot}%{_bindir}/files-sign-gui
install -Dm0644 packaging/com.rquispe.filessign.desktop \
        %{buildroot}%{_datadir}/applications/com.rquispe.filessign.desktop
install -Dm0644 packaging/com.rquispe.filessign.metainfo.xml \
        %{buildroot}%{_metainfodir}/com.rquispe.filessign.metainfo.xml
install -Dm0644 assets/icon-512.png %{buildroot}%{_datadir}/icons/hicolor/512x512/apps/com.rquispe.filessign.png
install -Dm0644 assets/icon-256.png %{buildroot}%{_datadir}/icons/hicolor/256x256/apps/com.rquispe.filessign.png
install -Dm0644 assets/icon-128.png %{buildroot}%{_datadir}/icons/hicolor/128x128/apps/com.rquispe.filessign.png
install -Dm0644 assets/icon.svg     %{buildroot}%{_datadir}/icons/hicolor/scalable/apps/com.rquispe.filessign.svg

%check
desktop-file-validate %{buildroot}%{_datadir}/applications/com.rquispe.filessign.desktop
appstream-util validate-relax --nonet %{buildroot}%{_metainfodir}/com.rquispe.filessign.metainfo.xml

%files
%license LICENSE
%doc README.md MANUAL.md
%{_bindir}/files-sign
%{_bindir}/files-sign-gui
%{_datadir}/applications/com.rquispe.filessign.desktop
%{_metainfodir}/com.rquispe.filessign.metainfo.xml
%{_datadir}/icons/hicolor/*/apps/com.rquispe.filessign.*

%changelog
* Fri Jul 03 2026 Ruddy Quispe <development.rquispe@gmail.com> - 0.1.0-1
- Initial package: CLI + desktop app to sign, verify and extract files.
