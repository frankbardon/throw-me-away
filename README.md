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

### Date syntax

`--due` accepts:

- `today`, `tomorrow`
- weekday names (`monday`, `next friday`)
- relative spans (`in 3 days`, `in 2 weeks`)
- RFC3339 timestamps (`2026-12-31T00:00:00Z`)

### Priorities

`low`, `medium` (default), `high`.

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
