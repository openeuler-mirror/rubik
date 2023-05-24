Name: rubik
Version: 2.0.0
Release: 1
Summary: Hybrid Deployment for Cloud Native
License: Mulan PSL V2
URL: https://gitee.com/openeuler/rubik
Source0: https://gitee.com/openeuler/rubik/repository/archive/v%{version}.tar.gz
Source1: git-commit
Source2: VERSION-vendor
Source3: apply-patch
Source4: gen-version.sh
Source5: series.conf
Source6: patch.tar.gz
Source7: build_rubik_image.sh
BuildRoot: %{_tmppath}/%{name}-%{version}-build
BuildRequires: golang >= 1.17

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
cp %{SOURCE7} .

%build
sh ./apply-patch
make release
strip ./build/rubik

%install
# create directory /var/lib/rubik
install -d %{buildroot}%{_sharedstatedir}/%{name}
# create directory /var/lib/rubik/build for image build
install -d %{buildroot}%{_sharedstatedir}/%{name}/build
# install rubik binary to build folder
install -Dp ./build/rubik %{buildroot}%{_sharedstatedir}/%{name}/build
# install artifacts
install -Dp ./hack/rubik-daemonset.yaml %{buildroot}%{_sharedstatedir}/%{name}/rubik-daemonset.yaml
install -Dp ./Dockerfile %{buildroot}%{_sharedstatedir}/%{name}/Dockerfile
install -Dp ./build_rubik_image.sh %{buildroot}%{_sharedstatedir}/%{name}/build_rubik_image.sh

%files
%dir %attr(750,root,root) %{_sharedstatedir}/%{name}
%attr(550,root,root) %{_sharedstatedir}/%{name}/build/rubik
%attr(640,root,root) %{_sharedstatedir}/%{name}/rubik-daemonset.yaml
%attr(640,root,root) %{_sharedstatedir}/%{name}/Dockerfile
%attr(550,root,root) %{_sharedstatedir}/%{name}/build_rubik_image.sh

%clean
rm -rf %{buildroot}

%changelog
* Wed May 24 2023 vegbir <yangjiaqi16@huawei.com> - 2.0.0-1
- Type:bugfix
- CVE:NA
- SUG:restart
- DESC:upgrade rubik version to v2.0.0

* Tue Nov 29 2022 CooperLi <a710905118@163.com> - 1.0.0-5
- Type:bugfix
- CVE:NA
- SUG:restart
- DESC:add build folder for container image build

* Mon Nov 21 2022 yangjiaqi <yangjiaqi16@huawei.com> - 1.0.0-4
- Type:bugfix
- CVE:NA
- SUG:restart
- DESC:add yaml

* Wed Nov 16 2022 yangjiaqi <yangjiaqi16@huawei.com> - 1.0.0-3
- Type:bugfix
- CVE:NA
- SUG:restart
- DESC:bump version

* Wed Nov 16 2022 yangjiaqi <yangjiaqi16@huawei.com> - 1.0.0-2
- Type:bugfix
- CVE:NA
- SUG:restart
- DESC:set the burst value for the pod to enable the container burst

* Mon Nov 14 2022 hanchao <hanchao47@huawei.com> - 1.0.0-1
- Type:bugfix
- CVE:NA
- SUG:restart
- DESC:upgrade rubik version to v1.0.0

* Mon Sep 19 2022 yangjiaqi <yangjiaqi16@huawei.com> - 0.1.0-3
- Type:bugfix
- CVENA
- SUG:restart
- DESC:fix compile problem and make rubik real static

* Tue Jan 11 2022 DCCooper <1866858@gmail.com> - 0.1.0-2
- Type:bugfix
- CVE:NA
- SUG:restart
- DESC:fix compile error

* Mon Dec 27 2021 xiadanni <xiadanni1@huawei.com> - 0.1.0-1
- Package init
