cd ../

make docker-down
make docker-up

go run main.go ingest testdata/quicktest/libB.json
go run main.go ingest testdata/quicktest/libA.json
go run main.go ingest testdata/quicktest/dep1.json
