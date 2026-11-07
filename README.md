# clean my cache

[Leia em português](README.pt-BR.md)

Free up gigabytes on your Mac by cleaning the caches that developer tools and apps leave behind, and **see what each one is before you delete it**.

```
Developer tools
  + npm                                  1.4 GB  safe
  + Playwright browsers                  1.0 GB  re-download
  + Homebrew                             812 MB  safe

Apps
  + Slack                                1.3 GB  safe
  ! VS Code                              929 MB  skipped, Visual Studio Code is open

┃ What should be cleaned?
┃ > ✓ npm                     1.4 GB  safe     Every package npm has ever downloaded...
┃   ✓ Slack                   1.3 GB  safe     Images and files from conversations...
┃   • App logs                 76 MB  caution  Log files written by apps...
```

Pick what you want, confirm, done. No Node, Python or anything else needed: it is a single small binary.

## Install

**Homebrew**

```bash
brew install jefferson-calmon/tap/clean-my-cache
```

**Without Homebrew**

```bash
curl -fsSL https://raw.githubusercontent.com/jefferson-calmon/clean-my-cache/main/install.sh | sh
```

The script downloads the right binary for your Mac (Apple Silicon or Intel), checks its checksum and installs it in `~/.local/bin`.

**With Go**

```bash
go install github.com/jefferson-calmon/clean-my-cache/cmd/clean-my-cache@latest
```

## Usage

| What you want | Command |
| --- | --- |
| Pick what to clean, then confirm | `clean-my-cache` |
| Only see how much space you can free | `clean-my-cache scan` |
| Clean only some items | `clean-my-cache --only npm,slack` |
| Clean everything except some items | `clean-my-cache --exclude 'docker-*'` |
| Move to the Trash instead of deleting | `clean-my-cache --trash` |
| Clean without questions (scripts, cron) | `clean-my-cache --yes` |
| See everything it knows how to clean | `clean-my-cache list` |
| Understand one item | `clean-my-cache explain <id>` |

`--only` and `--exclude` accept wildcards. Add `--json` for machine-readable output and `--verbose` to also see items with nothing to clean.

Shell completion for zsh, bash and fish comes with Homebrew, or run `clean-my-cache completion --help`.

## Why

- **It explains itself.** Every item says what it is and what happens if you delete it.
- **Safe by default.** Only cache folders are touched. Passwords, logins, history, settings, projects and messages are never deleted.
- **It waits for you.** Apps that are open are skipped, so nothing gets corrupted. It tells you which ones to close.
- **It uses each tool's own cleaner** when there is one (`brew cleanup`, `npm cache clean`, `pnpm store prune`, `docker builder prune`).
- **Easy to extend.** Every app is a small YAML file. Add your own without touching code.

## What the labels mean

| Label | Meaning | Selected by default |
| --- | --- | --- |
| `safe` | Nothing is lost. The app rebuilds it on its own. | Yes |
| `re-download` | Nothing is lost, but it is downloaded or rebuilt next time, which can take a while. | Yes |
| `caution` | Usually fine, but you might want to keep it (app logs, Docker images). | No, use `--all` or `--only` |

## What it cleans

- **Developer tools:** Homebrew, npm, pnpm, Yarn, Bun, Deno, pip, uv, Go, CocoaPods, Xcode DerivedData, iOS Simulator, Docker, Playwright, Puppeteer, Prisma
- **Browsers:** Chrome, Edge, Brave, Arc, Firefox
- **Apps:** VS Code, Cursor, Slack, Discord, Figma, Claude, Notion, Obsidian, Postman, Spotify, Steam
- **System:** app updates that were downloaded and never installed, app logs

Run `clean-my-cache list` for the full list with ids.

## Your own recipes

Drop a YAML file in `~/.config/clean-my-cache/recipes/` and it shows up next to the built-in ones:

```yaml
id: my-tool
name: My Tool
category: dev
risk: safe
description: Files my internal tool downloads. It downloads them again when needed.
paths:
  - ~/Library/Caches/my-tool
```

Your recipes follow the same safety rules as the built-in ones. If it is useful for others, [open a pull request](CONTRIBUTING.md).

## What it never touches

- Anything outside `~/Library/Caches`, `~/Library/Application Support`, `~/Library/Logs`, `~/.cache`, `~/.npm`, the pnpm store and Xcode build folders
- Symbolic links, or folders reached through one
- Files that need `sudo`
- WhatsApp and Telegram media: those are your photos and videos, not cache. Clean them inside each app (Settings → Storage).

These rules live in one place, [`internal/pathguard`](internal/pathguard/guard.go), are checked again right before each deletion, and are covered by tests.

## FAQ

**Will the space come back?** Partly. Caches grow again as you use your apps, that is their job. Running it once a month keeps things tidy.

**Why is the freed space different from what Finder says?** macOS counts some space as "purgeable" and keeps local snapshots, so Finder can lag behind.

**Can I undo it?** Use `--trash` and folders go to the Trash first. Tools with their own cleaner (Homebrew, npm, pnpm, Docker) still clean themselves, because only they know what is safe to remove.

## Development

You need Go 1.26 or newer.

```bash
make run ARGS="scan -v"   # run from source
make test                 # tests run against a temporary fake home folder
make lint
make build                # creates bin/clean-my-cache
```

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
