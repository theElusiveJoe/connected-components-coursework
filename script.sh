go build main.go && 
mpiexec -n $((1+$(($1))+$(($2)))) -oversubscribe main --mode=mpi -r="$1" -f="$3" 
