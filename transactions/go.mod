module github.com/nimbo1999/financeiro/transactions

go 1.24.0

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/go-chi/chi v1.5.5
	github.com/go-chi/chi/v5 v5.2.3
	github.com/go-chi/cors v1.2.2
	github.com/jackc/pgx/v5 v5.7.6
	github.com/nimbo1999/financeiro/commons v0.0.0
	github.com/nimbo1999/financeiro/migrator v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.11.1
	gorm.io/driver/postgres v1.6.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang-migrate/migrate/v4 v4.19.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.29.0 // indirect
	gorm.io/gorm v1.31.0
)

replace (
	github.com/nimbo1999/financeiro/commons => ../commons
	github.com/nimbo1999/financeiro/migrator => ../modules/migrator
	github.com/nimbo1999/financeiro/users => ../users
)
