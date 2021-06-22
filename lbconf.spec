%global provider      gitlab
%global provider_tld  cern.ch
%global project       lb-experts
%global provider_full %{provider}.%{provider_tld}/%{project}
%global repo          golbconf

%global import_path   %{provider_full}/%{repo}
%global gopath        %{_datadir}/gocode
%global debug_package %{nil}

Name: lbconf
Version: 0.1
Release: 2%{?dist}
Summary: CERN DNS Load Balancer Config Generator
License: ASL 2.0
URL: https://%{import_path}

Source: %{name}-%{version}.tgz
BuildRequires: golang >= 1.13
ExclusiveArch: x86_64

%description
%{summary}

This is a Golang implementation of the CERN LB Configuration Generator.

It queries the Ermis REST service to get the alias parameter information.

It also queries PuppetDB to recover the alias node membership.
It combines this information with the AllowedNodes and ForbiddenNodes
information from the alias parameters.

It generates the lbd config file with the information above.

%prep
%setup -n %{name}-%{version} -q

%build
go build -o lbconf -mod=vendor

%install
# main package binary
install -d -p %{buildroot}%{_bindir}
install -p -m0755 lbconf %{buildroot}%{_bindir}/lbconf

%files
%doc LICENSE COPYING README.md
%attr(755,root,root) %{_bindir}/lbconf

%changelog
* Tue Jun 22 2021 Ignacio Reguero <ignacio.reguero@cern.ch> - 0.1-2
- Fix filenames
* Mon Jun 21 2021 Ignacio Reguero <ignacio.reguero@cern.ch> - 0.1-1
- First version of the rpm
