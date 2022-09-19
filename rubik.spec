Name: rubik
Version: 0.1.0
Release: 4
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
strip rubik

%install
# create directory /var/lib/rubik
install -d %{buildroot}%{_sharedstatedir}/%{name}
# install rubik binary
install -Dp ./rubik %{buildroot}%{_sharedstatedir}/%{name}
# install artifacts
install -Dp ./hack/rubik-daemonset.yaml %{buildroot}%{_sharedstatedir}/%{name}/rubik-daemonset.yaml
install -Dp ./Dockerfile %{buildroot}%{_sharedstatedir}/%{name}/Dockerfile

%files
%dir %attr(750,root,root) %{_sharedstatedir}/%{name}
%attr(550,root,root) %{_sharedstatedir}/%{name}/rubik
%attr(640,root,root) %{_sharedstatedir}/%{name}/rubik-daemonset.yaml
%attr(640,root,root) %{_sharedstatedir}/%{name}/Dockerfile

%clean
rm -rf %{buildroot}

%changelog
* Mon Sep 19 2022 yangjiaqi <yangjiaqi16@huawei.com> - 0.1.0-4
- Type:bugfix
- CVE:NA
- SUG:restart
- DESC:strip rubik

* Mon Sep 19 2022 yangjiaqi <yangjiaqi16@huawei.com> - 0.1.0-3
- Type:bugfix
- CVE:NA
- SUG:restart
- DESC:fix compile problem and make rubik real static

* Tue Jan 11 2022 DCCooper <1866858@gmail.com> - 0.1.0-2
- Type:bugfix
- CVE:NA
- SUG:restart
- DESC:fix compile error

* Mon Dec 27 2021 xiadanni <xiadanni1@huawei.com> - 0.1.0-1
- Package init
