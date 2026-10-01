# Page templates

One skeleton per page type. Section names in `<angle brackets>` are placeholders; headings without
them are fixed. Drop a section only when it would be empty. Wrap prose at 100 columns.

## Home (`docs/README.md`)

```markdown
# rtunk

<One sentence: what rtunk is and for whom.>

<One short paragraph: the problem it solves, and the single most important property.>

| rtunk is   | rtunk is not |
| ---------- | ------------ |
| <property> | <non-goal>   |

## Getting started

<One line: who this group is for.>

- [<Page title>](<Page-File>.md) — <what the reader gets from it>

## <Next group…>
```

## Guide (task-oriented: `Checking-Code`, `Formatting-Code`, …)

````markdown
# <Page Title>

<Short version: one or two sentences, then the minimal command block that does the common case.>

## <Task 1, as the reader would phrase it>

<When/why, one or two sentences.>

```console
$ rtunk <command>
<real output>
```

<What the output means; what to do next.>

> [!NOTE]
> <Exception, platform caveat or gotcha — only if there is one.>

## <Task 2…>

## Where to go next

- [<Reference page>](<File>.md#<anchor>) — every flag of the commands used here
- [<Related guide>](<File>.md) — <why the reader would go there>
````

## Reference (`Command-Reference`, `Configuration-Reference`)

````markdown
# <Page Title>

<Scope in one sentence, and what it was verified against (e.g. `rtunk help --all`).>

| <Entry>                | <Purpose>  |
| ---------------------- | ---------- |
| [`<entry>`](#<anchor>) | <one line> |

## `<entry>`

<One-sentence purpose.>

```text
<synopsis>
```

| Flag / key | Type | Default | Description |
| ---------- | ---- | ------- | ----------- |

<Example with real output, when it clarifies.>

## Where to go next
````

## FAQ

```markdown
# FAQ

## <Question exactly as a user would ask it?>

<Direct answer first sentence. Then detail or a link to the page that owns the topic.>
```

## Internals (explanation / design spec: `Architecture`, `CLI-Design`, …)

```markdown
# <Page Title>

> [!NOTE]
> Contributor documentation. <One line on audience, e.g. "Describes the design rtunk is built to,
> not how to use it — see [Command Reference](Command-Reference.md) for that.">

<Short version: what this part of the system is responsible for, in two or three sentences.>

## <Concept / component>

<Motivation and responsibility; Mermaid diagram for flows; trade-offs stated explicitly.>

## Where to go next
```

## `_Sidebar.md`

```markdown
**[Home](README.md)**

**Getting started**

- [<Title>](<File>.md)

**<Group>**

- [<Title>](<File>.md)
```

## `_Footer.md`

```markdown
[Home](README.md) · [FAQ](FAQ.md) · [Report an issue](https://github.com/xunleii/rtunk/issues) ·
Source of this page: `docs/` in the [rtunk repository](https://github.com/xunleii/rtunk)
```
