# Contributing

Thanks for helping. Most contributions are new **recipes**: a small YAML file that teaches clean-my-cache how to clean one app or tool. You do not need to know Go for that.

## Add an app in four steps

1. **Find the cache folder.** Use the app a bit, then look in `~/Library/Caches` and `~/Library/Application Support/<App>`. Folders named `Cache`, `Code Cache` or `GPUCache` are caches. Folders with your data (databases, `Local Storage`, `IndexedDB`, `Cookies`, settings) are **not**.

2. **Create the recipe** at `recipes/<category>/<id>.yaml`. The first line turns on autocomplete and validation in VS Code (with the YAML extension):

   Most desktop apps are built with Electron and need only this:

   ```yaml
   # yaml-language-server: $schema=../../schema/recipe.schema.json
   id: linear
   name: Linear
   category: apps
   risk: safe
   app: Linear
   description: "Images and files the app saved to open faster. Your issues live in the cloud."
   preset: electron
   support_folder: Linear
   ```

   For any other folder:

   ```yaml
   # yaml-language-server: $schema=../../schema/recipe.schema.json
   id: my-tool
   name: My Tool
   category: dev
   risk: safe
   description: "What it is and what happens when it is deleted, in plain words."
   paths:
     - ~/Library/Caches/my-tool
   ```

   If the tool has its own clean command, prefer it:

   ```yaml
   # yaml-language-server: $schema=../../schema/recipe.schema.json
   id: my-tool
   name: My Tool
   category: dev
   risk: safe
   description: "What it is and what happens when it is cleaned."
   run: [my-tool, cache, clean]
   measure:
     - ~/Library/Caches/my-tool
   ```

3. **Check it on your machine:**

   ```bash
   make run ARGS="explain my-tool"
   make run ARGS="scan --only my-tool --verbose"
   ```

   Tip: you can also test it without rebuilding by copying the file to `~/.config/clean-my-cache/recipes/`.

4. **Run the checks** and open a pull request:

   ```bash
   make check
   ```

## Recipe fields

| Field | Required | What it does |
| --- | --- | --- |
| `id` | yes | Lowercase kebab-case, same as the file name. Used in `--only` and `explain`. Never rename it after release. |
| `name` | yes | Shown to people. |
| `category` | yes | `dev`, `browsers`, `apps` or `system`. Same as the folder. |
| `risk` | yes | `safe`, `redownload` or `caution`. See below. |
| `description` | yes | One or two short sentences: what it is and what happens after cleaning. Written for someone who is not a developer. |
| `app` | no | The `.app` name in `/Applications`, without `.app`. The recipe is skipped while the app is open. |
| `preset` | no | `electron` adds the standard Chromium cache folders inside `~/Library/Application Support/<support_folder>`. |
| `support_folder` | with `preset` | Folder inside `~/Library/Application Support`. |
| `paths` | no | Folders to delete. Start with `~/`. `*` matches one folder level. |
| `paths_from` | no | Folders found by running a command, for example the current pnpm store. |
| `run` | no | The tool's own clean command, as a list. Never a shell. |
| `measure`, `measure_from` | no | Folders used to show the size before and after `run`. |
| `precheck` | no | Command that must succeed before `run`, like `[docker, info]`. |
| `unavailable_hint` | no | Shown when the precheck fails. |

A recipe has either something to delete (`paths`, `preset` or `paths_from`) or a `run` command, never both.

## Choosing the risk

- `safe`: the app rebuilds it automatically and quickly. Browser and app caches.
- `redownload`: nothing is lost, but getting it back is slow or large. Test browsers, build outputs.
- `caution`: people may want to keep it. Never selected by default.

When in doubt, pick the higher risk.

## Rules every recipe follows

Checked every time a recipe is loaded, so a bad recipe never runs, and covered by tests:

- Paths to delete live inside the allowed roots in [`internal/pathguard/guard.go`](internal/pathguard/guard.go). If you need a new root, open an issue first and explain why it only contains cache.
- Paths never contain `..`.
- Commands name a program, never a path or a shell (`sh`, `bash`, `zsh`, `env`, `sudo`, `rm`...).
- Unknown fields are rejected, so typos never pass silently.
- Every recipe has a real description.

## Changing the engine

```
cmd/clean-my-cache/   entry point
internal/cli/         commands and flags (cobra)
internal/engine/      wires everything together, selection rules
internal/recipe/      YAML loading, presets and validation
internal/pathguard/   deletion safety and glob expansion
internal/scan/        finds and measures caches in parallel
internal/clean/       deletes or moves to Trash, runs tool commands
internal/size/        native disk usage, like du
internal/ui/          colors, report, picker, spinner
recipes/              built-in recipes, embedded in the binary
schema/               JSON Schema for editors
```

- Run `make fmt` before committing.
- Tests use a temporary fake home folder. Never write a test that touches your real home.
- No comments in code. Explain decisions in the pull request description.
