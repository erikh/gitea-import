# gitea-import: import repositories into Gitea in an automated fashion

I wrote this in a hurry as my existing Gitea died hard; and I needed to
reimport without a database attached; that is, I had all my repositories, just
not the surrounding database data.

This tool was born in a few hours and reads like it. Really, what you should do
with it, if anything, is read and modify it to suit your needs.

The tool requires a manifest to import the directories (called a "change list"
here in the docs) and you'll have to generate this yourself. There are some
tips at the end of this README that help with generation.

This tool will:

- Span repositories and organizations
- Create as necessary, and ignore on multiple runs
- Handles private repositories (see the change list)
- Detects default branch (current `HEAD` of git repo) properly pushes the
  content through git and not by any other means.

## Installing

```bash
go get github.com/erikh/gitea-import
```

## Running

First, create a Gitea token through the API for your **new** gitea instance. Then, run
this in the `/data/git/repositories` directory in your **old/existing** gitea instance:

```bash
gitea-import <url> <api token> <change list filename>
```

## Example Change List

It contains the repository name, qualified by org/username and the privacy
setting (true is private). These are transposed into directory names in the
Gitea `repositories` data directory.

```
erikh/ldhcpd false
erikh/secret true
erikh/duct false
```

## Generating the repository list

You can do this easily with the following command, but you must go in and
edit+annotate each line with a `true` or `false` to indicate private repositories.

```bash
find . -mindepth 2 -maxdepth 2 -type d | sed -e 's!^./!!' | sed -e 's!.git$!!' >changelist
```

## Author

Erik Hollensbe <github@hollensbe.org>
