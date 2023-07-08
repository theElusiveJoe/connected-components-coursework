echo "1 - routers num"
echo "2 - slaves num"
echo "3 - hash num"
echo "4 - test_file"

go build main.go && 
mpiexec -n $((1+$(($1))+$(($2)))) -oversubscribe main --mode=mpi-with-dist -routers="$1" -hash="$3" -file="$4" 
