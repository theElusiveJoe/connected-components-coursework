package fastSVMpi

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"
import (
	"bytes"
	"connectedComponents/algos"
	"encoding/gob"
	"fmt"
	"os"
	"path"
)

// true, если нужно прогрнать еще один круг, иначе false
func runStep6SaveResult(tr *transRole, conf *algos.RunConfig) {
	switch tr.role {
	// оставлю тут заглушки - мб потом понадобятся
	case MASTER:
		runStep6Master(tr)
	case ROUTER:
		runStep6Router(tr)
	case SLAVE:
		runStep6Slave(tr, conf)
	}
}

func runStep6Master(tr *transRole) {
}

func runStep6Router(tr *transRole) {
}

func runStep6Slave(tr *transRole, conf *algos.RunConfig) {
	fmt.Println("DIR:", conf.ResultDir)
	if err := os.MkdirAll(conf.ResultDir, os.ModePerm); err != nil {
		panic(err)
	}

	name := conf.Id + "_" + fmt.Sprintf("%d", tr.rank-tr.routersNum) + ".mapbin"
	p := path.Join(conf.ResultDir, name)
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	if err := e.Encode(tr.slave.f); err != nil {
		panic("result encoding failed")
	}
	os.Create(p)
	if err := os.WriteFile(p, b.Bytes(), 0755); err != nil {
		panic(err)
	}
}
