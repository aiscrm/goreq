package sonic

import (
	"fmt"
	"testing"

	"github.com/aiscrm/goreq/codec"
)

func TestSonic(t *testing.T) {
	c := NewCodec(
		codec.WithEscapeHTML(true),
	)
	user1 := struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	}{
		ID:   11111111,
		Name: "哈哈哈",
		URL:  "https://meetings.feishu.cn/s/1i38ftnck0f18?src_type=3&name=哈哈",
	}
	data1, err := c.Marshal(user1)
	if err != nil {
		panic(err)
	}
	fmt.Println("==" + string(data1) + "==")
	user2 := &struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	}{}
	if err = c.Unmarshal(data1, user2); err != nil {
		panic(err)
	}
	if user2.ID != user1.ID || user2.Name != user1.Name || user2.URL != user1.URL {
		t.Fail()
	}

	user3 := struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
	}{
		ID:   22222222,
		Name: "哈哈哈2",
	}
	data3, err := c.Marshal(user3)
	if err != nil {
		panic(err)
	}
	fmt.Println("==" + string(data3) + "==")
	user4 := &struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
	}{}
	if err = c.Unmarshal(data3, user4); err != nil {
		panic(err)
	}
	if user3.ID != user4.ID || user3.Name != user4.Name {
		t.Fail()
	}
}
