# Troubleshooting

This guide covers the current Home Server Installer V1 path.

## Installer does not appear

V1 is UEFI-only. Make sure the VM or test machine is booting the installer ISO in UEFI mode.

If the Fedora CoreOS live environment boots but the TUI does not appear immediately, wait for the live system to finish starting. The Home Server Installer service runs on `tty1`.

From another shell, useful checks are:

```bash
sudo systemctl status home-server-installer.service --no-pager
sudo journalctl -u home-server-installer.service -b --no-pager
```

## Target disk is missing

Check what the live system sees:

```bash
lsblk -o NAME,SIZE,TYPE,MODEL,SERIAL,MOUNTPOINTS
ls -l /dev/disk/by-id/ 2>/dev/null
```

The installer prefers stable disk identity when available. Some VM disks do not expose useful `/dev/disk/by-id` entries; in that case the installer may fall back to the raw virtio device path.

## Installation appears to pause around image deployment

The signed container image must be downloaded and deployed before the installation can finish. Progress may remain at roughly the same percentage for several minutes while this happens.

Do not power off or reboot simply because the percentage has not moved for a while.

## Installation fails

Before rebooting the live environment, collect:

```bash
cat /tmp/knuckle.log 2>/dev/null || true
sudo journalctl -u home-server-installer.service -b --no-pager
lsblk -o NAME,SIZE,FSTYPE,TYPE,MODEL,SERIAL,MOUNTPOINTS
```

If reporting a problem, include the selected image, selected disk, selected `/boot` layout, and the relevant installer log output.

## First boot: SSH host key changed

If the same IP address was previously used by the live installer or an older installation, SSH may report:

```text
REMOTE HOST IDENTIFICATION HAS CHANGED!
```

A fresh installation generates a fresh SSH host identity. If you know this is the machine you just reinstalled, remove the stale client entry and reconnect:

```bash
ssh-keygen -R <IP>
ssh <user>@<IP>
```

Verify and accept the new host fingerprint when prompted.

## First boot: public-key login fails

The installer writes the configured public SSH key into the installed user's authorized keys.

If the matching private key is loaded in `ssh-agent`, is in a normal default SSH identity location, or is configured in the SSH client, plain SSH should work:

```bash
ssh <user>@<IP>
```

If the private key uses a custom filename and is not loaded in an agent, specify it explicitly:

```bash
ssh -i ~/.ssh/<private-key> <user>@<IP>
```

The filename of the public `.pub` file used during ISO creation does not affect server-side authorization; the public-key contents are what matter.

## First boot: check the installed system

Useful basic checks:

```bash
sudo bootc status
sudo systemctl --failed --no-pager
lsblk -f
```

For SSH-key troubleshooting:

```bash
ls -ld ~/.ssh
ls -l ~/.ssh/authorized_keys
```

## Secure Boot

Secure Boot must be disabled during the current V1 installation path. After installation, follow the current [uCore documentation](https://github.com/ublue-os/ucore) if you want to configure or enable Secure Boot.

## Testing safety

VM testing is recommended first. For bare-metal testing, use a dedicated test drive or hardware where the selected installation disk can be safely erased, and keep backups of anything important.
