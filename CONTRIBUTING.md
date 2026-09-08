<!-- Modified by Home Server Project from Project Bluefin Knuckle. -->
# Contributing to Home Server Installer

Thanks for contributing to Home Server Installer.

This repository is a Home Server Project fork of Project Bluefin Knuckle, adapted for a signed, direct-install uCore workflow.

## Before making changes

Read:

- [README.md](README.md) for current V1 behavior.
- [docs/TESTING.md](docs/TESTING.md) for the validation workflow.
- [docs/SECURITY.md](docs/SECURITY.md) for trust and disk-safety rules.
- [AGENTS.md](AGENTS.md) if you are using an AI coding agent against the repository.

For larger behavioral changes, open an issue first so the scope can be discussed before implementation.

## Clone

```bash
git clone https://github.com/home-server-project/home-server-installer.git
cd home-server-installer
```

## Go development

The project uses Go and keeps CGO disabled for the installer binary.

For the current Home Server-specific path, a useful minimum test set is:

```bash
GOTOOLCHAIN=auto go test ./internal/install ./internal/model ./internal/tui
```

Before submitting Go changes:

```bash
gofmt -w <changed-go-files>
```

Review the resulting diff before committing formatting changes.

## Installer and disk changes

Changes involving any of the following require VM validation, not only unit tests:

- target-disk discovery or validation;
- partition layout;
- `bootc install` behavior;
- image selection or signature verification;
- primary-user or password provisioning;
- SSH key provisioning;
- ISO boot/startup behavior.

Use a disposable VM first. When disk behavior changes, attach a separate sentinel disk and verify that only the selected target disk is modified.

See [docs/TESTING.md](docs/TESTING.md) for the current test procedure.

## SSH material

Never commit private SSH keys or embed private keys in an installer ISO.

The installer and future Builder flow only require SSH **public keys**. Private keys stay on the user's own device.

## Documentation changes

If behavior changes, update the relevant documentation in the same change:

- `README.md` for product behavior;
- `CHANGELOG.md` for Home Server-specific milestones;
- `docs/TESTING.md` for validation changes;
- `docs/TROUBLESHOOTING.md` for operational fixes;
- `docs/SECURITY.md` for trust or disk-safety changes;
- `docs/RELEASE.md` for release requirements.

## Pull requests

Keep pull requests focused and describe:

- what changed;
- why it changed;
- what tests were run;
- whether VM installation was required and completed;
- any remaining limitations.

Do not claim bare-metal behavior from VM testing alone. If dedicated hardware was tested, say exactly what was tested.

## Upstream

The original project is [Project Bluefin Knuckle](https://github.com/projectbluefin/knuckle). Useful upstream fixes can still be reviewed and adapted when they fit the Home Server Installer architecture.