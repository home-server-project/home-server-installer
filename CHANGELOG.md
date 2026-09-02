# Changelog

All notable changes to Home Server Installer are documented here.

## 0.1-alpha - 2026-09-02

Initial repository alpha assembled from the verified disposable-VM installer
runbook.

### Added

- interactive hardware-disk discovery and exact destructive confirmation;
- `home-server-default-v1` GPT layout;
- embedded OCI archive checksum verification;
- SSH public-key validation and bootc root-key injection;
- target-backed Podman image workspace;
- mandatory non-empty root and boot UUID gate;
- direct `bootc install to-filesystem --skip-finalize` path;
- OSTree deployment-root discovery;
- Installer-owned Home Server update policy;
- persistent `00-home-server.preset` for Zincati/rpm-ostree first-boot behavior;
- temporary workspace cleanup before bootc finalization;
- BLS, origin, SSH-injection, boot-artifact, and update-policy verification;
- installation report;
- dry-run mode;
- non-destructive unit tests and GitHub Actions CI;
- tag-driven release workflow.

### Important verified lessons incorporated

- `bootc` can accept an empty `UUID=` mount specification, so Installer must
  abort before bootc whenever filesystem UUID capture fails.
- `/target` is the OSTree sysroot after install; systemd policy must target the
  actual deployed root under `/target/ostree/deploy/.../deploy/*.0`.
- Fedora CoreOS first-boot preset processing can remove a manually enabled
  `rpm-ostreed-automatic.timer`; a local preset is required for the tested
  Home Server update policy.
