# todo

A small terminal task tracker. Stores tasks in a JSON file under your XDG data
directory and lets you add, list, complete, and delete them.

## Install

```sh
go install github.com/frankbardon/todo/cmd/todo@latest
```

Or build from source:

```sh
git clone https://github.com/frankbardon/todo
cd todo
make build
./todo --help
```

## Usage

```sh
todo add "buy milk" --priority high --due tomorrow --tag shopping
todo list
todo list --status todo --tag shopping
todo show 1
todo done 1
todo delete 1
```

### Priorities

`low`, `medium` (default), `high`.

## Due dates

Set a due date on a new task with `add --due <expr>`, or change one later with
`edit --due <expr>`. Drop a due date with `edit --clear-due`.

```sh
todo add "pay rent" --due "next friday"
todo edit 3 --due tomorrow
todo edit 3 --clear-due
```

### Supported date forms

`--due` and `--due-before` accept the same expressions:

- `today`, `tomorrow`, `yesterday`
- weekday names: `monday`, `tuesday`, ..., `sunday` (also short forms like `mon`, `fri`)
- `next <weekday>` (forces the following week even if today matches)
- `in N days`, `in N weeks`
- `YYYY-MM-DD` (e.g. `2026-12-31`)
- RFC3339 timestamps (e.g. `2026-12-31T00:00:00Z`)

### Triage

`list` has three flags for working through due work:

- `--overdue` — only tasks that are past due and not done.
- `--due-before <expr>` — only tasks due before the given date.
- `--sort due` — order by due date ascending; undated tasks sort last.

Overdue rows are shown in red when stdout is a terminal, and prefixed with `!`
when piped to a file or another command.

### Example

```sh
todo add "pay rent" --due "next friday" --priority high
todo add "buy milk" --due tomorrow --tag shopping
todo add "read book"

todo list
todo list --overdue
todo list --sort due
todo list --due-before "in 7 days"
```

## Configuration

Tasks are stored at `$XDG_DATA_HOME/todo/tasks.json`, defaulting to
`~/.local/share/todo/tasks.json`. Override with `--config`.

## Development

```sh
make test       # go test -race -cover ./...
make vet        # go vet ./...
make cover      # writes coverage.html
```

## License

MIT.
