<!-- Modified by Home Server Project from Project Bluefin Knuckle. -->
# Security

Home Server Installer is a privileged installer. V1 is designed around a small, explicit trust and disk-safety model.

## Image trust

The Home Server install path accepts only the supported signed image repositories exposed by the installer.

- Home Server Project images are verified with the Home Server Project Cosign public key.
- Upstream `ublue-os/ucore*` images are verified with Universal Blue's uCore Cosign public key.
- Unknown image repositories are rejected by the Home Server path.
- Temporary signature-discovery configuration used during installation is removed and does not persist into the installed system.

## Target-disk safety

The user selects the installation target disk in the TUI. All installer-created partitions are placed on that selected drive.

Immediately before partitioning, the installer revalidates the selected target, including block-device identity, expected size, serial when available, and mounted-filesystem checks.

VM testing has verified that a separate attached non-target sentinel disk remains unchanged during installation.

The selected target disk is intentionally erased and repartitioned as part of installation. VM testing is recommended first. Bare-metal testing should use a dedicated test drive or hardware where the selected installation disk can be safely erased, with backups of anything important.

## SSH keys and passwords

The installer supports a local password and SSH authorized keys for the primary user.

Only SSH **public keys** belong in the installer ISO or future Builder workflow. Private SSH keys must remain on the user's own device and must never be embedded in an ISO or committed to this repository.

The planned [Home Server uCore Builder](https://github.com/home-server-project/home-server-ucore-builder) will use a public SSH key supplied through GitHub Secrets when creating a personalized ISO. The corresponding private key remains local to the user.

## Installed-system cleanup

The direct install path is designed so installer-only state does not remain in the finished system. Current VM validation confirms that temporary installer signature configuration, the live installer service, the live SSH handoff file, and `/opt/knuckle` are absent after installation.

## Secure Boot

Secure Boot must be disabled during the current V1 installation path. After installation, follow the current [uCore documentation](https://github.com/ublue-os/ucore) if you want to configure or enable Secure Boot.

## Reporting security issues

Please use the GitHub security reporting features for the [Home Server Installer repository](https://github.com/home-server-project/home-server-installer) rather than opening a public issue for a vulnerability.