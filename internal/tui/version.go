package tui

// installerVersion is supplied by cmd/knuckle at runtime from the version
// embedded into the binary at build time. Release and Builder workflows
// already set main.version from the Git tag, so the TUI can show that same
// Home Server Installer version instead of unrelated Flatcar metadata.
var installerVersion = "dev"

// SetInstallerVersion sets the version shown in the persistent TUI header.
func SetInstallerVersion(version string) {
	if version == "" {
		version = "dev"
	}
	installerVersion = version
}
