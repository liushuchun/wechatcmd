package wechat

import (
	"fmt"
	"testing"
)

func Test_GetUUID(t *testing.T) {
	we := NewWechat()
	we.GetUUID()
	fmt.Println(we.Uuid)
	t.Logf("%v", we.Uuid)
}
