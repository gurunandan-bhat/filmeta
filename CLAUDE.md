# CLAUDE.md — `filmeta`

Notes for working in this repo. Written from a full read of the code on 2026-07-31.

## What this is

`filmeta` is a personal Cobra CLI that feeds the **FCG Reviews** website
(`https://www.fcgreviews.com`), a Hugo site whose source lives at `$HOME/repos/guild`.
It is a **maintainer's toolbox**, not a service: single-user, run by hand, no CI,
no README. Commands pull metadata from TMDB, write JSON/images into the Hugo tree,
and emit CSV reports from Hugo's generated JSON.

Data flows one way in each direction:

```
TMDB API ──> filmeta ──> guild/assets/meta/*.json   (film metadata, consumed by Hugo)
                     ──> guild/assets/... posters/backdrops
                     ──> guild/data/freescores.json (critic scores)

guild/public/**/index.json ──> filmeta ──> CSV on stdout (reports)
```

**There is no database.** MySQL persistence backed a legacy feature that has since
been retired; the `model` package, the `metaModel.Save` call in `import`, the `db`
block in the config struct, and the `go-sql-driver/mysql` + `jmoiron/sqlx`
dependencies were all removed on 2026-07-31. The tool is now purely files-in,
files-out. Do not reintroduce a DB dependency without a specific reason —
everything downstream reads the JSON.

Note the asymmetry: **reports read Hugo's *built output* (`public/`), not its source.**
Run `hugo` in `$HOME/repos/guild` before running `topScores` / `criticReviews` /
`missingMeta`, or you are reporting on stale data.

## Layout

| Path | Role |
|---|---|
| [main.go](main.go) | 3-line shim to `cmd.Execute()` |
| [cmd/root.go](cmd/root.go) | root command; loads config in `PersistentPreRunE` |
| [cmd/types.go](cmd/types.go) | all shared decode/encode structs for the `cmd` package |
| [cmd/*.go](cmd/) | one file per subcommand |
| [cmd/daterange.go](cmd/daterange.go) | shared `--from-date`/`--to-date` handling for the reports |
| [config/config.go](config/config.go) | viper-backed config singleton + ISO-639 language lookup |
| [tmdb/](tmdb/) | TMDB v3 API client |
| [guild/guild.go](guild/guild.go) | **dead code** — nothing imports `filmeta/guild` |
| [filmeta.sql](filmeta.sql) | mysqldump of the **retired** schema; kept only as a record. Nothing in this repo reads or writes it any more — note it also contains `mdbsession`, which belongs to the separate web app sharing `~/.filmeta.json`, so check with the site before dropping the database itself. |

## Configuration

Single JSON file, default `$HOME/.filmeta.json`, overridable with `-c/--config-file`.
It is **shared with another application** — the `db`, `security`, `session`, and
`algolia` sections are unused by this CLI and have no matching struct fields. Viper
silently ignores keys with no field, so the file needs no edit and the web app keeps
working. Two keys drive almost everything:

- `appRoot` → this repo (`$HOME/repos/filmeta`); only used to find `config/languages.json`
- `hugoRoot` → **`$HOME/repos/guild/public`**, i.e. the *built* site, not the site root

That last point is the single most common source of confusion. Paths are derived
from it by walking upward:

| Code | Resolves to |
|---|---|
| `filepath.Join(hugoRoot, "/../data/freescores.json")` | `guild/data/freescores.json` |
| `filepath.Join(hugoRoot, "../assets", "meta")` | `guild/assets/meta/` |
| `hugoRoot + "/mreviews/index.json"` | `guild/public/mreviews/index.json` |
| `hugoRoot + "/critics/index.json"` | `guild/public/critics/index.json` |
| `hugoRoot + "/guild/index.json"` | `guild/public/guild/index.json` |

### The viper/json-tag trap

`config.Config` is tagged with `json:"..."` but populated via `viper.Unmarshal`,
which reads **mapstructure** tags. It currently works only because viper lowercases
keys and falls back to case-insensitive *field-name* matching (`apiKey` → `apikey` →
`APIKey`). **Any new field whose JSON key isn't just a case variant of the Go field
name will silently stay zero.** If you add a config field, either name it to match or
add a `mapstructure:` tag — and verify by printing it, not by reading the struct.

`Configuration()` memoizes into a package-level `c` guarded by `if (c == Config{})`.
Consequence: the *first* call wins. `initLang()` calls `Configuration()` with no args,
so if anything ever calls `ISOLanguage` before the root command runs, the `-c` flag is
silently ignored.

## The md5 filename/key convention

This is the load-bearing contract between filmeta and Hugo, and it is undocumented
anywhere else. **Entities are keyed by the lowercase hex MD5 of their `LinkTitle`**
(the human-readable title/name, exactly as it appears in the Hugo content).

- `guild/assets/meta/<md5(LinkTitle)>.json` — TMDB metadata blob per film
- `public/mreviews/index.json` — `map[md5(LinkTitle)] → film object`
- `public/critics/index.json` — `map[md5(LinkTitle)] → critic object`
- `public/critics/<review-url>/index.json` — `map[md5(film LinkTitle)] → review object`
- `public/guild/index.json` — a **JSON array**, not a map (of guild members)

Verified against live data: all three map forms hash `LinkTitle` correctly.
The codebase spells the same hash two ways — `fmt.Sprintf("%x", md5.Sum(b))` and
`hex.EncodeToString(h[:])`. They are identical; don't "fix" one into the other
thinking it changes behaviour.

Because the key is the title, **renaming a film in Hugo orphans its metadata file**
and it will resurface in `missingMeta`.

The Hugo templates producing these files live in
`$HOME/repos/guild/themes/guild/layouts/{mreviews,critics,guild,reviews}/*.json`.
If a report suddenly decodes to zero rows, check whether one of those templates
changed shape — git history here shows this has happened twice
(`Update to handle new json output format`, `Handle conversion of json link from a
slice to a map`).

## Commands

| Command | Touches TMDB? | Purpose |
|---|---|---|
| `film <id>` | yes | dump one film + credits as JSON to stdout; `-t` for TV |
| `import <input.json>` | yes | bulk: fetch films, write metadata JSON + posters/backdrops |
| `createPosts` | yes | generate Hugo `index.md` posts with TOML front matter |
| `score <value> -a <critic> -f <film>` | no | record a critic's score into `freescores.json` |
| `missingMeta` | no | list films in Hugo with no `assets/meta` JSON |
| `topScores` | no | CSV of film/language/lastmod/score in a date range |
| `criticReviews` | no | CSV of critic/orgs/review-count in a date range |

`root.go`'s `PersistentPreRunE` now only loads config, so no command needs any
external service beyond TMDB (and only the three that fetch need even that).

`score` uses `PreRunE` (not `PersistentPreRunE`), so it correctly runs *in addition to*
root's hook. Keep it that way — a subcommand defining `PersistentPreRunE` would
**replace** root's and skip config loading entirely.

`missingMeta` and `topScores`/`criticReviews` are the routine reporting loop:
build the site, find gaps, run `import` to fill them.

## Verified defects

These were confirmed against the live database and checked-in schema, not just read.

### 1. `REPLACE INTO film` never replaced — OBSOLETE, code deleted

*Resolved 2026-07-31 by removing the database entirely, not by fixing the SQL.*

Recorded for context, since the `filmeta` database still exists on disk. `model.Save`
relied on `REPLACE INTO film ... (iTMDBID, ...)` being idempotent, but the `film`
table's only index is `PRIMARY KEY (iFilmID)` (auto-increment, never supplied).
`REPLACE` de-duplicates on PRIMARY/UNIQUE conflicts only, so with no unique key on
`iTMDBID` it degraded to a plain `INSERT`. Every re-import added a duplicate film row
plus a duplicate cast+crew set, because the `DELETE FROM film_credit` used the newly
minted `iFilmID`.

**The leftover data is still there**: as last measured, 384 film rows for 374 distinct
`iTMDBID`, and 1,045 orphaned `film_credit` rows. Nothing reads them now. If the site
never needs those tables again they can be dropped — but confirm against the web app
first, since it shares the same database for `mdbsession`.

### 2. `criticReviews` panicked on a member with no organizations — FIXED

*Fixed 2026-07-31 in [cmd/criticReviews.go](cmd/criticReviews.go).*

The branch taken when a critic's `index.json` was absent indexed
`member.Organizations[0]` directly, so a guild member with an empty
`Organizations` array crashed the command with `index out of range [0] with
length 0`. All 56 current members have at least one organization, so this was
latent — but the branch itself is live (Meena Iyer has no review index), and
adding one organization-less member would have triggered it.

The root cause was that the two branches rendered the same column differently:
the missing-index path emitted `Organizations[0]` (the *first* organization only),
while the normal path emitted the full `orgMap` joined. They are now unified —
a `switch` on the open error, one shared `joinOrgs(orgMap)` append — so the
missing-index case reports every organization instead of just the first, and
an empty set renders as an empty column.

`joinOrgs` also **sorts**, because `orgMap` is a map and its iteration order is
random; the organization column previously varied between runs for the 3 members
with more than one. Blank organization strings are now dropped rather than
becoming an empty element, and trimming moved from `strings.Trim(s, " ")` to
`strings.TrimSpace` so tabs and newlines are handled too.

Verified against the live site: same 56 members, identical review counts and
identical organization *sets*; only the ordering within the column changed, and
the command's full output is now byte-identical across runs.

### 3. Date filters excluded the end day — FIXED

*Fixed 2026-07-31 in [cmd/daterange.go](cmd/daterange.go); kept here because the
underlying pflag behaviour will bite again if new date flags are added.*

`--from-date` / `--to-date` parse with layout `2006-01-02`. pflag's `timeValue.Set`
calls `time.Parse`, which for a layout with no zone returns **midnight UTC** — not
local midnight. Three bugs followed:

- the whole of `--to-date` was dropped, since every timestamp that day is after
  its midnight;
- with `hugoRoot` timestamps at `+05:30`, anything on `--from-date` before 05:30
  IST was also dropped (54 films in the current data set);
- the flag *defaults* are `time.Now()` / `now.AddDate(-1,0,0)` — local wall-clock
  values carrying a time of day — so defaults and explicit flags behaved differently.

Both commands now call `dateRange(cmd)`, which re-anchors each bound to local
midnight by calendar date and returns the half-open interval `[start, end)` with
`end` advanced a day. Filter with `inRange(t, start, end)`; report with
`describeRange(start, end)`, which renders the inclusive pair back to the user.
An inverted window is now a hard error rather than silently empty output.

**If you add another date flag, use `addDateRangeFlags`/`dateRange` rather than
`Flags().TimeP` directly**, or you will reintroduce the UTC skew.

Regression tests live in [cmd/daterange_test.go](cmd/daterange_test.go) — the only
tests in the repo. They pin `time.Local` to a fixed `+05:30` zone via `TestMain`
so they pass on any host; **anything else added to package `cmd` inherits that
`TestMain`.**

### 4. `import`'s readability check rejected private files — FIXED

*Fixed 2026-07-31 in [cmd/import.go](cmd/import.go).*

`dataIsAvailable` used `info.Mode().Perm()&0444 != 0444`, which demands the file be
readable by owner **and** group **and** world. A `0600` file owned by the invoking
user — the normal mode for anything written by a private script — was rejected as
"not readable".

Mode bits cannot answer "can this process read this file?" on their own; the answer
depends on ownership, group membership and ACLs. The check now just opens the file
and closes it, which is the only honest test, plus an `IsDir` guard so a directory
argument fails with a clear message instead of an `EISDIR` further in.

**Prefer opening over `Stat`-and-inspect-mode anywhere else this pattern shows up.**

### 5. `ShowSearch` can loop forever

[tmdb/search.go:33](tmdb/search.go#L33) breaks only on `results.Page == results.TotalPages`.
A zero-result TMDB search returns `total_pages: 0`, which page 1 never equals — infinite
paging. (Currently unreachable: nothing calls `ShowSearch`.)

### 6. `GetConfiguration` targets a non-existent endpoint

[tmdb/configuration.go:37](tmdb/configuration.go#L37) requests `{base}/faces?limit=&page=`.
TMDB's endpoint is `/configuration` and takes no such params. This looks copy-pasted
from another project. Also unused.

## Rough edges (not bugs, but know them)

- **`gofmt` fails on [main.go](main.go)** — stray blank line in the header comment.
  `gofmt -l .` should be clean; it isn't. `go vet ./...` passes.
- **Hardcoded absolute path**: [cmd/createPosts.go:33](cmd/createPosts.go#L33) pins
  `rootFolder = "/Users/nandan/mog"`. The comment says this is deliberate (to keep the
  command away from the live guild site), but it makes the command machine-specific and
  it silently writes outside `hugoRoot`. It also `os.Create`s into
  `<rootFolder>/assets/metadata/` without `MkdirAll`, so it fails if that dir is absent —
  inconsistent with every other command, which creates its output dirs.
- **`defer` inside a loop**: [createPosts.go:111](cmd/createPosts.go#L111) accumulates
  open file handles for the whole run, and `log.Fatalf` in the deferred close would
  kill the process mid-cleanup.
- **`score` can't record 0.0** — [score.go:47](cmd/score.go#L47) rejects it alongside
  parse errors. `ParseFloat` already reports malformed input, so the `score == 0.0`
  guard only removes a legitimate value.
- **`score -o` is always joined to `hugoRoot`**, so you cannot pass an absolute output
  path. `defaultScoreFile` is the odd-looking `"/../data/freescores.json"` for the
  same reason.
- **`entityExists` swallows errors**: [score.go:137](cmd/score.go#L137) returns `""`
  identically for "file unreadable" and "not found", printing to stdout without a
  newline. A typo'd critic name and a missing site build are indistinguishable.
- **`ShowSearch` returns `*[]Film`** — pointer-to-slice, un-idiomatic.
- **Dead code**: the entire `guild` package, `ShowSearch`, `GetConfiguration`,
  `ISOLanguage`/`initLang` (and therefore `appRoot` and `config/languages.json`).
  `guild.Film` duplicates `cmd.Film`. The `model` package was in this list and has
  now been deleted; the rest are still candidates.
- **Cobra scaffolding noise**: generated `Copyright © 2025 NAME HERE <EMAIL ADDRESS>`
  headers, commented-out flag boilerplate in every `init()`, `"A brief description of
  your command"` shorts on `score`/`topScores`, and root's `--config-file` help text
  reading `"Help message for toggle"`.
- **Debug `fmt.Println` in output paths**: `import` still prints `"import called"` to
  stdout. Harmless there (it emits no machine-readable output), but the same pattern
  in `topScores`/`criticReviews` corrupted their CSV on redirect; **fixed 2026-07-31**
  by dropping the `"<cmd> called"` scaffolding line and sending the date-range banner
  to stderr. **Keep stdout reserved for the CSV in any new reporting command.**
- **`sendRequest` treats 3xx as success** ([tmdb/client.go:48](tmdb/client.go#L48)).
  Harmless today because `http.Client` follows redirects.
- **`ids.txt`** is committed test input with stray whitespace and a duplicate;
  `getIDsFromFile` tolerates it (trims, skips blanks, logs bad lines).

## Conventions to follow

- **Errors wrap with `%w`** and a lowercase context phrase; commands return errors from
  `RunE` rather than exiting. `log.Fatalf`/`panic` appear only in `init()` for
  programmer errors (a required flag that doesn't exist).
- **New subcommand** = new file in `cmd/`, `rootCmd.AddCommand(...)` in its `init()`,
  shared structs into [cmd/types.go](cmd/types.go). Add a hyphenated `Aliases` entry
  for camelCase names (`createPosts` → `create-posts`), as the newer commands do.
- **`tmdb.Film.FCGTitle`** (`json:"fcg_title"`) is *synthesized by filmeta*, not returned
  by TMDB — `import` copies the Hugo `LinkTitle` into it so the written metadata file
  records which site entry it belongs to. TMDB's own `title` often differs from the
  title the site uses.
- **Movie vs TV**: TMDB puts the title in `title` for movies and `name` for TV.
  `showType` is the literal path segment `"movie"` or `"tv"`, taken from `FilmOut.ShowType`
  in the input file. Note the fallback from `title` to `name` lived in the deleted
  `model.Save`; **`createPosts` still uses `tmdbFilm.Title` unguarded**, which is empty
  for a TV show — harmless only because it hardcodes `"movie"`.

## Working here

```bash
go build ./...        # binary `filmeta` is gitignored
go vet ./...          # clean
go test ./...         # cmd/ only; the other packages have no tests
gofmt -l .            # currently reports main.go
```

Test coverage is limited to the date-range helpers, `import`'s argument validator,
and `criticReviews`' organization handling — all in package `cmd`.

Nothing this tool does is transactional any more, but `import` and `createPosts`
still **overwrite files in place** in the Hugo tree. Point them at a scratch
`--output-dir` when experimenting. The reports are read-only, and all of them
depend on `hugo` having been run in `$HOME/repos/guild`.

Every command runs against a config with no `db` block, so a throwaway config is an
easy way to test against fixture data:

```bash
python3 -c "import json,os; c=json.load(open(os.path.expanduser('~/.filmeta.json'))); \
  c['hugoRoot']='/tmp/fake/root'; json.dump(c, open('/tmp/fake/config.json','w'))"
./filmeta criticReviews -c /tmp/fake/config.json
```

A useful way to check a change to the reports is to build the previous revision
into a scratch dir and diff the CSV output against the new one:

```bash
git worktree add /tmp/old HEAD && (cd /tmp/old && go build -o /tmp/filmeta-old .)
diff <(/tmp/filmeta-old topScores -f 2025-01-01 -t 2025-12-31 2>/dev/null) \
     <(./filmeta        topScores -f 2025-01-01 -t 2025-12-31 2>/dev/null)
```

Both report commands now keep stdout to the CSV alone, so `2>/dev/null` is enough to
isolate it — no `tail` needed.

`criticReviews` output is fully deterministic (it iterates the `guild/index.json`
array, and `joinOrgs` sorts). **`topScores` is not** — it ranges over a
`map[string]Film`, so its row order varies between runs. `sort` both sides before
diffing it. Giving `topScores` a `slices.SortFunc` before the CSV write is an easy
outstanding improvement.
