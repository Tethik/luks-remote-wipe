# luks-remote-wipe

This program attempts to remove the keyslots used by Linux's CryptSetup disk encryption utility, effectively
leaving a disk encrypted with an unknown master key.

Use cases:

1. Stolen computers that somehow had disk encryption bypassed via e.g. weak passphrase or having been left unlocked.
2. Admins needing to remotely shut down online compromised computers e.g. from rogue employees or hacked.

🚨⚠️ Use with caution! This program will cause you to lose your data. Don't run this with sudo.

# Requirements

This program is currently only intended to run on LUKS-encrypted Linux setups.

# Dependencies

Requires libcryptsetup-dev to compile
