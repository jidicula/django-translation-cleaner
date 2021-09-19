[![Build](https://github.com/jidicula/django-translation-cleaner/actions/workflows/build.yml/badge.svg)](https://github.com/jidicula/django-translation-cleaner/actions/workflows/build.yml) [![Latest Release](https://github.com/jidicula/django-translation-cleaner/actions/workflows/release-draft.yml/badge.svg)](https://github.com/jidicula/django-translation-cleaner/actions/workflows/release-draft.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/jidicula/django-translation-cleaner)](https://goreportcard.com/report/github.com/jidicula/django-translation-cleaner) [![Go Reference](https://pkg.go.dev/badge/github.com/jidicula/django-translation-cleaner.svg)](https://pkg.go.dev/github.com/jidicula/django-translation-cleaner)

# django-translation-cleaner

Command django-translation-cleaner is a tool for cleaning unused translations from .po files in a Django project. django-translation-cleaner returns a zero exit code when no unused translations are found, and a non-zero exit code when there are unused translations.


### Do you find this useful?

Star this repo!

### Do you find this *really* useful?

You can sponsor me [here](https://github.com/sponsors/jidicula)!

# Usage

1. Get the tool with `go install github.com/jidicula/django-translation-cleaner@latest` or grab a binary for your OS and arch [here](https://github.com/jidicula/django-translation-cleaner/releases/latest).

2. Clean unused translations from a Django project:

```
django-translation-cleaner /path/to/repo
```

3. Check for unused translations in a Django project. If there are any unused
translations, prints them to stdout and returns non-zero exit code:

```
django-translation-cleaner --check /path/to/repo
```

# GitHub Action
Coming soon!
