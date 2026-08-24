# aps

AWS profile switcher, kubectx-style. Run `aps` with no arguments to
fuzzy-pick a profile from `~/.aws/config` / `~/.aws/credentials`, or
`aps <profile>` to jump straight to one. Like the AWS CLI itself, aps
honors `AWS_CONFIG_FILE` / `AWS_SHARED_CREDENTIALS_FILE` if you keep
these files elsewhere (e.g. `~/.config/aws`).

A plain binary can't change its parent shell's environment, so `aps`
follows the [zoxide](https://github.com/ajeetdsouza/zoxide) pattern: the
binary only prints a profile name, and a shell function (installed via
`aps init <shell>`) captures that and exports it.

## Install

```sh
go install github.com/obvionaoe/aps@latest
```

Then add the shell integration to your rc file:

```sh
eval "$(aps init zsh)"    # ~/.zshrc
eval "$(aps init bash)"   # ~/.bashrc
aps init fish | source    # ~/.config/fish/config.fish
```

Or let `aps init <shell> --install` append that line for you.

Restart your shell (or re-source the rc file).

## Usage

```sh
aps              # fuzzy-search profiles, export the one you pick
aps prod         # export AWS_PROFILE=prod directly
aps --unset      # unset AWS_PROFILE
```
