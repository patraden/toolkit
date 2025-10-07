## copy from/to Vertica

```shell
# select csv from vertica
VERTICA_HOST="hostname1" \
VERTICA_PASSWORD="password1" \
./copy_stdout.sh copy_stdout_query.sql > out.csv

# copy csv into vertica
cat out.csv | \
VERTICA_HOST="hostname2" \
VERTICA_PASSWORD="password2" \
./copy_stdin.sh "schema.table"

# all in one
VERTICA_HOST="hostname1" \
VERTICA_PASSWORD="password1" \
./copy_stdout.sh copy_stdout_query.sql | \
VERTICA_HOST="hostname2" \
VERTICA_PASSWORD= \
./copy_stdin.sh "schema.table"
```