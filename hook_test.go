package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/techxmind/go-utils/object"
)

func testTrimJsonComment(t *testing.T) {
	ast := assert.New(t)

	jsonStr := []byte(`{
"__debug" : true,
"qconf" : {
	"zk" : "192.168.111.6:2181", //开发机ZK
	// 基础配置
	"basic_conf" : "/conf/dev/go/basic"
}
}`)

	expect := []byte(`{
"__debug" : true,
"qconf" : {
	"zk" : "192.168.111.6:2181",
	"basic_conf" : "/conf/dev/go/basic"
}
}`)

	jsonStr = trimJsonComment(jsonStr, T_JSON)

	var v1, v2 interface{}

	err1 := json.Unmarshal(jsonStr, &v1)
	err2 := json.Unmarshal(expect, &v2)

	ast.Nil(err1)
	ast.Nil(err2)
	ast.Equal(v1, v2)
	v, _ := object.GetValue(v1, "__debug")
	ast.Equal(true, v)
}
