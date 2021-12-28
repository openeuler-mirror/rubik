Name: rubik
Version: 0.1.0
Release: 1
Summary: Hybrid Deployment for Cloud Native
License: Mulan PSL V2
URL: https://gitee.com/openeuler/rubik
Source0: https://gitee.com/openeuler/rubik/repository/archive/v%{version}.tar.gz
Source1: git-commit
Source2: VERSION-openeuler
Source3: apply-patch
Source4: gen-version.sh
Source5: series.conf
Source6: patch.tar.gz
BuildRoot: %{_tmppath}/%{name}-%{version}-build
BuildRequires: golang >= 1.13

%description
This is hybrid deployment component for cloud native, it should be running in kubernetes environment.

%prep
cp %{SOURCE0} .
cp %{SOURCE1} .
cp %{SOURCE2} .
cp %{SOURCE3} .
cp %{SOURCE4} .
cp %{SOURCE5} .
cp %{SOURCE6} .

%build
sh ./apply-patch
make release

%install
# create directory /var/lib/rubik
install -d -m 0750 %{buildroot}%{_sharedstatedir}/%{name}
# install rubik binary
install -Dp -m 0550 ./rubik %{buildroot}%{_sharedstatedir}/%{name}
# install artifacts
install -Dp -m 0640 ./hack/rubik-daemonset.yaml %{buildroot}%{_sharedstatedir}/%{name}/rubik-daemonset.yaml
install -Dp -m 0640 ./Dockerfile %{buildroot}%{_sharedstatedir}/%{name}/Dockerfile

%files
%dir %{_sharedstatedir}/%{name}
%{_sharedstatedir}/%{name}/rubik
%{_sharedstatedir}/%{name}/rubik-daemonset.yaml
%{_sharedstatedir}/%{name}/Dockerfile

%clean
rm -rf %{buildroot}

%changelog
* Mon Dec 27 2021 xiadanni <xiadanni1@huawei.com> - 0.1.0-1
- Package init
