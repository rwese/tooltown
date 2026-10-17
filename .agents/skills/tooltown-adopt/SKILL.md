---
name: tooltown-adopt
description: "Adopt a project into the Tooltown catalog: judge whether its repository is ready, then prepare it upstream — public, licensed, installable from a pinned ref, runnable, and describable — before the catalog entry is written here. Use when: (1) a user asks to adopt a tool or repo into Tooltown, (2) judging whether a project is ready for Tooltown, (3) preparing an upstream project so its Tooltown entry can be written."
license: MIT
---

# Adopting a project into Tooltown

Tooltown catalogs small CLI tools. Adoption touches two repositories: this
skill prepares the project **in its own repository**, and the `tooltown` skill
writes the catalog entry **here**. Preparation comes first — an entry written
against an unprepared project ships an install command nobody can run.

A project is **ready** when every prerequisite below passes. Report the whole
list, met or unmet, and name what is missing instead of writing the entry
around a gap.

## Readiness prerequisites

Check the project's own working copy (for example
`~/Repositories/<host>/<owner>/<repo>`). Every row is checkable. Projects live
on GitHub (`gh`) or on Forgejo (API calls with
`gopass show forgejo/ai-factory-token`; see `~/.agents/references/forgejo.md`).

| Prerequisite | How to check | Unmet means |
| --- | --- | --- |
| Public repository | GitHub: `gh repo view <owner>/<repo> --json visibility` reports `PUBLIC`. Forgejo: `private` is `false` in `/api/v1/repos/<owner>/<repo>`, and an anonymous request to the repo URL returns `200` | the entry would expose a private project's existence, language, license, and branch |
| HTTPS source URL | the repo URL starts with `https://` | Hugo's content adapter rejects a non-HTTPS `source_url` and the build fails |
| Pinnable install ref | a commit SHA, release tag, or registry version exists for the install | `install:` ships unpinned and the page reproduces nothing later |
| Install verified by running it | the pinned command installs, then `<tool> --help` and the README's first invocation both run | the showcased command is a guess |
| README substance | the README says what the tool does, how to install it, and shows at least one invocation | `tagline`, `summary`, `features`, and `usage` have nothing to draw on |
| Declared license | a recognised license file is committed and the host detects it (GitHub `licenseInfo.spdxId`, Forgejo `license`). A manifest field such as `license = "MIT"` alone does not count | the page shows `NOT DECLARED` |
| Real CLI surface | the tool runs in a terminal with flags it can demonstrate | there is nothing to screenshot and nothing to list |

## Prepare the project

Run these in the project's repository, stopping at each unmet prerequisite.
Every step ends when its prerequisite passes.

1. **Publish the repository.** Make it public at the URL that will become
   `source_url` — GitHub: `gh repo edit <owner>/<repo> --visibility public
   --accept-visibility-change-consequences`; Forgejo: repository settings.
   Done when an anonymous request to the repository URL returns `200`.
2. **Declare a license.** Commit a license file in a form the host recognises
   (`LICENSE` at the root is the safe default). Done when the host reports the
   license — GitHub `licenseInfo.spdxId`, Forgejo `license` — whose name
   becomes the `license:` display string.
3. **Cut a pinnable ref.** Tag a release, or note the registry version or
   commit SHA the install will pin to. Done when the ref resolves
   (`git rev-parse <sha>`, `npm view <pkg>@<version>`).
4. **Verify the pinned install.** In a scratch directory under `/tmp/opencode`,
   run the exact command the entry will show, then `<tool> --help` and the
   README's first invocation. Done when the command and its real output are
   captured — this is the entry's `install:` and `usage:`.
5. **Capture one real visual.** A roughly 1200×675 SVG or PNG of an actual
   terminal session showing that output, not a mockup. Done when the file
   exists and matches the behaviour of the ref from step 3.
6. **Fill the README gaps.** Add whatever steps 4–5 exposed as missing: what
   the tool does in one line, how to install it, one invocation. Done when
   every field the entry needs can be read off the README, step 3's ref, or
   step 4's terminal output.

## Hand off to the catalog entry

With the project ready, write the entry here:
`.agents/skills/tooltown/references/new_tool.md` holds the schema, screenshot
placement, `scripts/build-site`, verification, and commit flow. Step 4's
command and output fill `install:` and `usage:`; step 5's file belongs in
`tools/<slug>/screenshots/`.

## When a project is not adoptable

A project that must stay private, or that has no CLI surface, cannot be
adopted. Report that to the user and stop — cataloging a redacted version
produces an entry whose install command fails for every reader.
