# Keeping Tools Up To Date

`rtunk renovate enable` adds Renovate annotations above the version pins in your configuration, so
[Renovate](https://docs.renovatebot.com/) can open pull requests when a linter or plugin source has
a new release.

```bash
rtunk renovate config      # add this output to your Renovate config (once)
rtunk renovate enable      # annotate the pins in your rtunk.yaml
```

## Teach Renovate about rtunk

Renovate does not understand `rtunk.yaml` on its own. `rtunk renovate config` prints a custom
regex manager that matches the annotations:

```console
$ rtunk renovate config
{
  "customManagers": [
    {
      "customType": "regex",
      "managerFilePatterns": ["/(^|/)\\.trunk/trunk\\.yaml$/", "/(^|/)\\.rtunk/rtunk\\.yaml$/"],
      "matchStrings": [
        "# renovate: datasource=(?<datasource>\\S+) depName=(?<depName>\\S+)(?:\\s+extractVersion=(?<extractVersion>\\S+))?\\s*\\n\\s*(?:-\\s*\\S+@|ref:\\s*)(?<currentValue>\\S+)"
      ]
    }
  ]
}
```

Merge the `customManagers` entry into the Renovate configuration at the root of your repository
(for example `renovate.json`). The manager covers both `.rtunk/rtunk.yaml` and
`.trunk/trunk.yaml`.

## Annotate the pins

```console
$ rtunk renovate enable
warning: no Renovate regexManager found for the annotations; add the output of `rtunk renovate config` to your Renovate config
annotated lint/shellcheck
annotated plugins.sources/trunk

2 annotated, 0 skipped
```

The warning appears while no Renovate configuration file at the repository root contains the
manager; it disappears once you added it. The annotations are YAML comments placed above each pin:

```yaml
plugins:
  sources:
    - id: trunk
      uri: https://github.com/trunk-io/plugins
      # renovate: datasource=github-tags depName=trunk-io/plugins
      ref: v1.11.0
lint:
  enabled: [
      # renovate: datasource=github-releases depName=koalaman/shellcheck
      shellcheck@0.11.0,
    ]
```

Each annotation names the Renovate datasource and the upstream repository (`depName`) to watch.
rtunk annotates plugin sources and enabled linters; the count it prints (`2 annotated, 0 skipped`)
reports how many pins it handled. In the example above, `shellcheck` had no explicit version, so
`enable` wrote the resolved version `0.11.0`.

> [!NOTE]
> `enable` rewrites the file, so it may reformat the entries it touches (as above, where the
> `enabled` list became a bracketed list). Review the diff before committing.

## Remove the annotations

```console
$ rtunk renovate disable
2 annotation(s) removed
```

## Where to go next

- [Command Reference](Command-Reference.md#renovate) — the three `renovate` commands
- [Managing Linters](Managing-Linters.md) — pin a linter with `linters enable <id>@<version>`
- [Configuration Reference](Configuration-Reference.md) — where plugin source `ref` pins live
