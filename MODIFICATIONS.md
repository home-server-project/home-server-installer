# Modifications to Project Bluefin Knuckle

Home Server Installer is derived from Project Bluefin Knuckle under the Apache License 2.0.

This inventory is based on the Knuckle fork point / merge base at commit `1eef60c3d852390b2ae9334d9ea6681243ab25c5`. Home Server-specific files added after that point are not listed here because they are not modified copies of upstream Knuckle files.

## Inherited modified files with direct notices

The following inherited files were modified by Home Server Project and now carry an in-file modification notice:

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
- README.md
- docs/README.md
- docs/RELEASE.md
- docs/SECURITY.md
- docs/TROUBLESHOOTING.md
- internal/ignition/builder.go
- internal/install/fcos_install.go
- internal/model/model.go
- internal/tui/fcos_tui_test.go
- internal/tui/form_logic.go
- internal/tui/handleenter_test.go
- scripts/tests/build-fcos-iso.bats

## Inherited modified files still needing direct notices

The following inherited files are identified here but were not directly rewritten during the connected-API licensing cleanup because complete file content could not be retrieved safely enough for a whole-file replacement. They should receive the same syntax-appropriate modification comment in a later local or patch-capable edit; no functional change is intended.

- Justfile
- internal/ignition/ignition.go
- internal/tui/forms.go
- internal/tui/tui.go
- internal/tui/update_test.go
- internal/wizard/wizard.go
- scripts/build-fcos-iso.sh

This repository-level inventory is intentionally not presented as a substitute for Apache License 2.0 section 4(b)'s file-level modified-file notice requirement.

The upstream project and history are available at https://github.com/projectbluefin/knuckle.
