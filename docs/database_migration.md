# Database Migration

We've used the golang-migrate library to manage the database migration.

### Install golang-migrate

```shell
$ brew install golang-migrate
```

### Commands

#### Create new database sequence

```shell
$ migrate create -ext sql -dir . -seq file_name
```

#### Up version database

```shell
$ migrate -source file://. -database "postgres://postgres:$PASSWORD@localhost:5432/postgres?sslmode=disable" up
```

#### Down version database 1 version

```shell
$ migrate -source file://. -database "postgres://postgres:$PASSWORD@localhost:5432/postgres?sslmode=disable" down 1
```

### References:

- Golang-Migrate: https://github.com/golang-migrate
- Connection string: https://www.connectionstrings.com/postgresql/
