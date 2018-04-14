tamate
============
Reading and Getting diffs between table-based data (CSV, SQL, Google Spreadsheets, etc...)

## Install

```sh
go get github.com/Mitu217/tamate
```

## Usage

### Generate DataSource Config.
```
tamate generate:config -t <datasource type> [-o <ouptut path>]

e.g. tamate generate:config -t SQL -o sql_config.json
```

### Generate Schema.
```
tamate generate:schema -t <datasource type> -c <config path> [-o <output path>]

e.g. tamate generate:schema -t SQL -c sql_config.json -o sql_schema.json
```

### Dump.
```
tamate dump <input datasource config path> [<output datasource config path>]

e.g. tamate dump sql_config.json  // SQL -> STDOUT
e.g. tamate dump sql_config.json spreadsheets_config.json  // SQL -> SpreadSheets
```

### Diff
```
tamate diff <schema> <left table> <right table>

e.g. tamate diff schema.json table1.json table2.json
```

## Contribution

### Requirements for development

- [docker](https://www.docker.com/) (For MySQL tests)
- [dep](https://github.com/golang/dep)

### Getting started

```sh
go get github.com/Mitu217/tamate
cd $GOPATH/src/github.com/Mitu217/tamate
dep ensure

# Run unit tests
go test ./...
```
