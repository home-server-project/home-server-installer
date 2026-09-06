package ignition

import (
	"fmt"

	"github.com/projectbluefin/knuckle/internal/model"
)

const homeServerPublicKeyPath = "/etc/containers/home-server-project.pub"

const homeServerPublicKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEWU3SeANKBm2Dql6FGZYNu2Bd7nZf
wSS/hmdv0B25JOSqi0dbyvW8XAJHJ4UOl/GeOSQM4XuDey9yI9I09r9XWw==
-----END PUBLIC KEY-----`

// GenerateHomeServerFCOSButane creates the intentionally small bootstrap
// configuration used by Home Server Installer V1. Fedora CoreOS is only an
// installation trampoline: user/network identity is provisioned, the Home
// Server signing key/policy is installed, and first boot performs a signed
// rpm-ostree rebase to uCore.
func (g *Generator) GenerateHomeServerFCOSButane(cfg *model.InstallConfig) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config cannot be nil")
	}
	if !isHomeServerImage(cfg.HomeServerImage) {
		return "", fmt.Errorf("unsupported Home Server image %q", cfg.HomeServerImage)
	}

	// Do not let generic temporary-FCOS policy leak into the final uCore
	// deployment. V1 keeps only identity/network/authentication/timezone.
	minimal := *cfg
	minimal.Sysexts = nil
	minimal.Swap = model.SwapConfig{}
	minimal.Tailscale = model.TailscaleConfig{}
	minimal.UpdateStrategy = model.UpdateStrategy{}
	minimal.NvidiaDriverVersion = ""

	b := NewBuilder(&minimal)
	if err := addSharedFragments(b, &minimal); err != nil {
		return "", err
	}

	keyFrag, err := renderTemplate("home-server-key", homeServerKeyTemplate, struct{ Key string }{Key: homeServerPublicKey})
	if err != nil {
		return "", fmt.Errorf("rendering Home Server signing key: %w", err)
	}
	b.AddStorageFile(keyFrag)

	policyFrag, err := renderTemplate("home-server-policy", homeServerPolicyTemplate, struct{ KeyPath string }{KeyPath: homeServerPublicKeyPath})
	if err != nil {
		return "", fmt.Errorf("rendering Home Server signature policy: %w", err)
	}
	b.AddStorageFile(policyFrag)

	autorebaseScriptFrag, err := renderTemplate("home-server-autorebase-script", homeServerAutorebaseScriptTemplate, struct{ Image string }{Image: cfg.HomeServerImage})
	if err != nil {
		return "", fmt.Errorf("rendering Home Server autorebase script: %w", err)
	}
	b.AddStorageFile(autorebaseScriptFrag)

	cleanupScriptFrag, err := renderTemplate("home-server-postrebase-cleanup-script", homeServerPostRebaseCleanupScriptTemplate, struct{}{})
	if err != nil {
		return "", fmt.Errorf("rendering Home Server post-rebase cleanup script: %w", err)
	}
	b.AddStorageFile(cleanupScriptFrag)

	autorebaseServiceFrag, err := renderTemplate("home-server-autorebase", homeServerAutorebaseTemplate, struct{}{})
	if err != nil {
		return "", fmt.Errorf("rendering Home Server autorebase service: %w", err)
	}

	cleanupServiceFrag, err := renderTemplate("home-server-postrebase-cleanup", homeServerPostRebaseCleanupTemplate, struct{}{})
	if err != nil {
		return "", fmt.Errorf("rendering Home Server post-rebase cleanup service: %w", err)
	}

	// Zincati belongs to the temporary FCOS deployment, not the final Home
	// Server update policy. Mask it so it cannot race the one-shot rebase.
	b.AddSystemdUnit("- name: zincati.service\n  mask: true")
	b.AddSystemdUnit(autorebaseServiceFrag)
	b.AddSystemdUnit(cleanupServiceFrag)

	return b.BuildFCOS(), nil
}

func isHomeServerImage(image string) bool {
	switch image {
	case model.HomeServerUCoreImage, model.HomeServerUCoreHCIImage:
		return true
	default:
		return false
	}
}

var homeServerKeyTemplate = `- path: /etc/containers/home-server-project.pub
  mode: 0644
  overwrite: true
  contents:
    inline: |
{{.Key | indentBlock 6}}`

var homeServerPolicyTemplate = `- path: /etc/containers/policy.json
  mode: 0644
  overwrite: true
  contents:
    inline: |
      {
        "default": [{"type": "reject"}],
        "transports": {
          "docker": {
            "ghcr.io/home-server-project/home-server-ucore": [{
              "type": "sigstoreSigned",
              "keyPath": "{{.KeyPath}}",
              "signedIdentity": {"type": "matchRepository"}
            }],
            "ghcr.io/home-server-project/home-server-ucore-hci": [{
              "type": "sigstoreSigned",
              "keyPath": "{{.KeyPath}}",
              "signedIdentity": {"type": "matchRepository"}
            }]
          }
        }
      }`

var homeServerAutorebaseScriptTemplate = `- path: /etc/home-server-installer/autorebase.sh
  mode: 0755
  overwrite: true
  contents:
    inline: |
      #!/usr/bin/bash
      set -u

      IMAGE="{{.Image}}"
      POLICY="/etc/containers/policy.json"
      FCOS_DEFAULT_POLICY="/usr/etc/containers/policy.json"
      BOOTSTRAP_KEY="/etc/containers/home-server-project.pub"
      STATE_DIR="/var/lib/home-server-installer"
      FAILURE_MOTD="/etc/motd.d/99-home-server-installer-rebase-failed"
      MAX_ATTEMPTS=5
      DELAYS=(5 10 20 40)

      log() {
          echo "[home-server-installer] $*"
      }

      write_failure() {
          local reason="$1"
          mkdir -p "${STATE_DIR}" /etc/motd.d
          touch "${STATE_DIR}/rebase-failed"
          cat > "${FAILURE_MOTD}" <<EOF_MOTD
      Home Server Installer: automatic uCore rebase failed.

      ${reason}

      Target: ${IMAGE}
      Check:  systemctl status home-server-autorebase.service
      Logs:   journalctl -u home-server-autorebase.service -b
      Retry:  sudo systemctl restart home-server-autorebase.service
      EOF_MOTD
          log "ERROR: ${reason}"
      }

      mkdir -p "${STATE_DIR}" /etc/motd.d
      rm -f "${STATE_DIR}/rebase-failed" "${FAILURE_MOTD}"

      for attempt in 1 2 3 4 5; do
          log "signed rebase attempt ${attempt}/${MAX_ATTEMPTS}: ${IMAGE}"
          if /usr/bin/rpm-ostree rebase --bypass-driver "ostree-image-signed:docker://${IMAGE}"; then
              log "signed rebase staged successfully"

              # Restore the stock FCOS policy before reboot. This makes /etc match
              # the current deployment default again, so OSTree's three-way merge
              # can take the Home Server image's own policy instead of carrying the
              # temporary bootstrap policy into the new deployment.
              if [[ ! -f "${FCOS_DEFAULT_POLICY}" ]]; then
                  write_failure "Signed rebase staged, but the FCOS default container policy is missing; refusing to reboot."
                  exit 1
              fi
              if ! /usr/bin/cp --remove-destination "${FCOS_DEFAULT_POLICY}" "${POLICY}"; then
                  write_failure "Signed rebase staged, but restoring the FCOS default container policy failed; refusing to reboot."
                  exit 1
              fi
              if command -v restorecon >/dev/null 2>&1; then
                  restorecon "${POLICY}" || true
              fi
              rm -f "${BOOTSTRAP_KEY}" "${FAILURE_MOTD}" "${STATE_DIR}/rebase-failed"
              touch "${STATE_DIR}/rebase-staged"

              /usr/bin/systemctl disable home-server-autorebase.service || true
              sync
              log "rebooting into Home Server uCore"
              /usr/bin/systemctl reboot
              exit 0
          fi

          if (( attempt < MAX_ATTEMPTS )); then
              delay="${DELAYS[$((attempt - 1))]}"
              log "attempt ${attempt} failed; retrying in ${delay}s"
              sleep "${delay}"
          fi
      done

      write_failure "Signed rebase failed after ${MAX_ATTEMPTS} attempts. The machine remains on Fedora CoreOS and can be retried safely."
      exit 1`

var homeServerPostRebaseCleanupScriptTemplate = `- path: /etc/home-server-installer/postrebase-cleanup.sh
  mode: 0755
  overwrite: true
  contents:
    inline: |
      #!/usr/bin/bash
      set -eu

      POLICY="/etc/containers/policy.json"
      IMAGE_DEFAULT_POLICY="/usr/etc/containers/policy.json"
      STATE_DIR="/var/lib/home-server-installer"
      FAILURE_MOTD="/etc/motd.d/99-home-server-installer-rebase-failed"

      echo "[home-server-installer] finalizing Home Server image trust"

      if [[ ! -f "${IMAGE_DEFAULT_POLICY}" ]]; then
          echo "[home-server-installer] ERROR: Home Server image default policy is missing" >&2
          exit 1
      fi

      # Replace any policy carried through the staged deployment with the policy
      # baked into the Home Server image itself.
      /usr/bin/cp --remove-destination "${IMAGE_DEFAULT_POLICY}" "${POLICY}"
      if command -v restorecon >/dev/null 2>&1; then
          restorecon "${POLICY}" || true
      fi

      rm -f /etc/containers/home-server-project.pub "${FAILURE_MOTD}"
      rm -f "${STATE_DIR}/rebase-staged" "${STATE_DIR}/rebase-failed"
      touch "${STATE_DIR}/rebase-complete"

      /usr/bin/systemctl disable home-server-autorebase.service home-server-postrebase-cleanup.service || true
      rm -f \
          /etc/systemd/system/home-server-autorebase.service \
          /etc/systemd/system/home-server-postrebase-cleanup.service \
          /etc/systemd/system/multi-user.target.wants/home-server-autorebase.service \
          /etc/systemd/system/multi-user.target.wants/home-server-postrebase-cleanup.service \
          /etc/home-server-installer/autorebase.sh \
          /etc/home-server-installer/postrebase-cleanup.sh
      /usr/bin/systemctl daemon-reload || true

      echo "[home-server-installer] Home Server image trust finalized"`

var homeServerAutorebaseTemplate = `- name: home-server-autorebase.service
  enabled: true
  contents: |
    [Unit]
    Description=Home Server signed uCore auto-rebase
    ConditionPathExists=!/var/lib/home-server-installer/rebase-staged
    After=network-online.target
    Wants=network-online.target

    [Service]
    Type=oneshot
    StandardOutput=journal+console
    StandardError=journal+console
    ExecStart=/etc/home-server-installer/autorebase.sh

    [Install]
    WantedBy=multi-user.target`

var homeServerPostRebaseCleanupTemplate = `- name: home-server-postrebase-cleanup.service
  enabled: true
  contents: |
    [Unit]
    Description=Finalize Home Server container signature policy after rebase
    ConditionPathExists=/var/lib/home-server-installer/rebase-staged
    ConditionPathExists=/usr/lib/pki/containers/iegorch86.pub
    After=local-fs.target

    [Service]
    Type=oneshot
    StandardOutput=journal+console
    StandardError=journal+console
    ExecStart=/etc/home-server-installer/postrebase-cleanup.sh

    [Install]
    WantedBy=multi-user.target`
