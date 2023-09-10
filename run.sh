rm log.txt
go build main.go
/usr/bin/go run main.go -mode=find-cc -algo=algo-mpi-$1 -file=tests/big/$2.csv
head log.txt