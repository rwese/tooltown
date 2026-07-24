# Adding a new tool to the Tooltown catalog

## TL;DR

```sh
# 1. Create the entry
mkdir -p tools/<slug>/screenshots
$EDITOR tools/<slug>/tooltown.yaml

# 2. Drop any screenshots beside it
cp ~/path/to/usage.svg tools/<slug>/screenshots/

# 3. Regenerate the served catalog
scripts/build-site

# 4. Verify
go test ./...
go vet ./...
git status --short          # static/ + tools/<slug>/ should be the only diff

# 5. Commit + push
git add tools/<slug>/ static/
git commit -m "feat(catalog): add <slug>"
git push origin main
```

The slug becomes both the directory name and the URL path
(`/tools/<slug>/`).

## Pre-flight checks

- **Slug is unique.** `ls tools/ | grep '^<slug>$'` should be empty.
- **Repo is public.** A private repo URL in the public catalog leaks
  its existence, language, license, and default-branch SHA. Make it
  public first.
- **Repo exists and is reachable.** `gh repo view <owner>/<repo>`.

## Schema

`schema: 1` declares the file format version. Keep it at the top.

### Required fields

| Field        | Notes                                                  |
| ------------ | ------------------------------------------------------ |
| `name`       | Matches the slug and the URL path.                     |
| `summary`    | Short tagline shown on the detail page (`<h2>`). Hugo fails to build without it. |
| `source_url` | Public URL to the project's homepage (usually GitHub). |

### Optional fields

| Field         | Notes                                                                  |
| ------------- | ---------------------------------------------------------------------- |
| `tagline`     | One-line hook shown next to the title.                                 |
| `description` | Markdown body. Renders below the summary.                              |
| `category`    | Free-form short label (e.g. `utility`). Shown uppercase in the UI.     |
| `status`      | `active`, `archived`, or similar. Shown uppercase in the UI.           |
| `visual`      | Design theme token: `signal`, `pocket`, or `drift`. Leave blank for default. |
| `language`    | Primary implementation language (free-form string).                    |
| `license`     | Free-form license display string (e.g. `MIT License`, `Not declared`). |
| `platforms`   | List of supported platforms (e.g. `[Linux, macOS, Windows]`).           |
| `install`     | Install command. Pin a SHA for reproducibility (`go install ...@<sha>`). |
| `features`    | List of bullets, each rendered as Markdown.                             |
| `usage`       | Code block (`|-` literal style) with example invocations.              |
| `screenshots` | List of `{src, alt, caption}`; `src` is relative to the tool directory. |

### Screenshots

Put images under `tools/<slug>/screenshots/`. Reference them by path
relative to that directory:

```yaml
screenshots:
  - src: screenshots/usage.svg
    alt: Terminal session showing pwd-copy commands for absolute and relative directory paths.
    caption: The CLI supports current, target, and relative directory paths.
```

Hugo validates that each `src` resolves while generating `static/`.

## Reference example

`tools/pwd-copy/tooltown.yaml` is the canonical seed entry. Copy it as
a starting point, then trim and adapt.

## Generating required fields from a GitHub repo

When the source repo is on GitHub, fetch what you can before writing by
hand:

```sh
gh repo view <owner>/<repo> \
  --json nameWithOwner,description,primaryLanguage,licenseInfo,url,defaultBranchRef
gh api graphql -F owner=<owner> -F name=<repo> -f query='
  query($owner: String!, $name: String!) {
    repository(owner: $owner, name: $name) {
      nameWithOwner url description
      primaryLanguage { name }
      licenseInfo { spdxId name }
      defaultBranchRef { target { oid } }
    }
  }'
gh api repos/<owner>/<repo>/readme --jq '.content' | base64 -d
```

Map into YAML:

- `name` ← `nameWithOwner` (strip `<owner>/`)
- `source_url` ← `url`
- `language` ← `primaryLanguage.name`
- `license` ← `licenseInfo.name`
- `install` SHA ← `defaultBranchRef.target.oid` (only when the install
  form is `go install <module>@<sha>`)
- `description`, `tagline`, `summary`, `features`, `usage` — pull from
  the README; the GitHub `description` is often empty or too short.

Editorial fields that must come from a human, never from GitHub:
`tagline`, `summary`, `category`, `visual`, `status`, `platforms`,
`features`, `usage`, `screenshots`.

## Verification checklist

Before committing:

- [ ] `scripts/build-site` exits 0 (Hugo accepts the YAML and screenshots).
- [ ] `git status --short` shows changes only under `tools/<slug>/` and `static/`.
- [ ] `go test ./...` passes.
- [ ] `go vet ./...` is clean.
- [ ] The new page renders at `/tools/<slug>/index.html` with title and
      summary populated (spot-check the generated HTML).

## Common pitfalls

- **Empty `summary`** — Hugo fails the build silently in some setups.
  Always provide one.
- **Missing screenshot file** — Hugo lists it as a warning and the page
  still builds, but the entry looks broken. Either ship the image or
  remove the `screenshots:` block.
- **Unpinned install command** — `go install ...@latest` is convenient
  but non-reproducible. Prefer `@<sha>` like `pwd-copy` does.
- **Editing `static/` by hand** — overwritten on the next
  `scripts/build-site`. Always edit `tools/<slug>/tooltown.yaml`.
- **Trailing-slash URL mismatch** — the URL is `/tools/<slug>/`, not
  `/tools/<slug>`.

## See also

- `AGENTS.md` — quality gates, commit style, container validation.
- `tools/README.md` — concise schema summary.
- `README.md` — full project documentation including deploy.
- `.env.dist` — every supported environment variable.
