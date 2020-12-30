# gauge-inprogress

A Gauge plugin to keep track of specs/scenarios that are marked with work in progress (wip) tags.

The aim of this plugin is to make sure that any specs or scenarios tagged with in-progress or wip tags are trackable.
And all scenarios marked as `in-progress` tags are followed by a reason so that it's easy to revisit these accordingly.

[![Gauge Badge](https://gauge.org/Gauge_Badge.svg)](https://gauge.org)

All notable changes to this project are documented in [CHANGELOG.md](CHANGELOG.md).
The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## :hammer_and_pick: Installation

* Install the plugin

```sh
gauge install inprogress
```

### Offline installation

* Download the plugin from [Releases](../../releases)

```sh
gauge install inprogress --file inprogress-${version}.zip
```

### Using the plugin

Add `inprogress` to your project's `manifest.json`.

```json
{
  "Language": "java",
  "Plugins": [
    "html-report",
    "inprogress"
  ]
}
```