# Modifications to Project Bluefin Knuckle

Home Server Installer is derived from Project Bluefin Knuckle under the Apache License 2.0.

Home Server Project has modified the following files that were inherited from the Knuckle fork point. These files also carry an in-file modification notice.

- .coderabbit.yaml
- .github/ISSUE_TEMPLATE/bug-report.yml
- .github/ISSUE_TEMPLATE/feature-request.yml
- .github/PULL_REQUEST_TEMPLATE.md
- .github/workflows/ci.yml
- .github/workflows/post-merge-smoke.yml
- .github/workflows/release.yml
- .github/workflows/security.yml
- .github/workflows/vm-e2e.yml
- AGENTS.md
- CHANGELOG.md
- CONTRIBUTING.md
- Justfile
- README.md
- docs/README.md
- docs/RELEASE.md
- docs/SECURITY.md
- docs/TROUBLESHOOTING.md
- internal/ignition/builder.go
- internal/ignition/ignition.go
- internal/install/fcos_install.go
- internal/model/model.go
- internal/tui/fcos_tui_test.go
- internal/tui/form_logic.go
- internal/tui/forms.go
- internal/tui/handleenter_test.go
- internal/tui/tui.go
- internal/tui/update_test.go
- internal/wizard/wizard.go
- scripts/build-fcos-iso.sh
- scripts/tests/build-fcos-iso.bats

Home Server-specific files added after the fork point are not listed here because they are not modified copies of upstream Knuckle files.

The upstream project and history are available at https://github.com/projectbluefin/knuckle.
