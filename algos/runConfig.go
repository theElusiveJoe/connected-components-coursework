package algos

import (
	"encoding/json"
)

type RunConfig struct {
	TestFile   string `json:"testFile"`
	ResultDir  string `json:"resultDir"`
	RoutersNum int    `json:"routesNum"`
	Slavesnum  int    `json:"slavesNum"`
	HashNum    int    `json:"hashNum"`
	Id         string `json:"id"`
}

func StrToConfig(s string) *RunConfig {
	var conf RunConfig
	if err := json.Unmarshal([]byte(s), &conf); err != nil {
		panic(err)
	}
	return &conf
}

func (conf *RunConfig) ConfigToStr() string {
	b, err := json.Marshal(conf)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (c *RunConfig) GetCopy() *RunConfig {
	c2 := new(RunConfig)
	c2.TestFile = c.TestFile
	c2.ResultDir = c.ResultDir
	c2.RoutersNum = c.RoutersNum
	c2.Slavesnum = c.Slavesnum
	c2.HashNum = c.HashNum
	c2.Id = c.Id
	return c2
}
