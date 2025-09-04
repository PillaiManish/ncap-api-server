Testing frameworks and libraries:

- The Go standard library testing package (`testing`)
- SQL mocking using `github.com/DATA-DOG/go-sqlmock`
- `bcrypt` used for password hashing checks in `LoginUser` tests

These tests focus on the `database/mysql` methods introduced or modified in the PR diff. They cover happy paths, edge cases, and failure conditions by asserting the SQL statements, arguments, and transaction behavior. External dependencies (actual MySQL) are mocked.