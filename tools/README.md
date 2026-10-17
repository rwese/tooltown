# Tool catalog entries

Create one directory per tool. The directory name becomes its URL slug.

```text
tools/
  example/
    tooltown.yaml
    screenshots/
      overview.webp
```

Required `tooltown.yaml` fields:

- `name`
- `summary`
- `source_url`

Supported optional fields include `tagline`, `description` (Markdown), `category`, `status`, `visual` (`signal`, `pocket`, or `drift`), `language`, `license`, `platforms`, `features`, `install`, `usage`, and `screenshots`. Each screenshot needs a path relative to the tool directory and meaningful alt text. `source_url` may point at GitHub or Forgejo; the detail page labels the link from the URL host.

Run `scripts/build-site` after changing metadata or assets. Hugo validates required fields and referenced screenshots while generating `static/`.
