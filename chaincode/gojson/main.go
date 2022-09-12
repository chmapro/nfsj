package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

func main() {
	jsonInfo, err := os.Open("/Users/changye/project/nfsj/chaincode/gojson/users.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened users.json")
	defer jsonInfo.Close()

	byteValue, _ := ioutil.ReadAll(jsonInfo)

	var blockAppendixs BlockAppendixs

	json.Unmarshal(byteValue, &blockAppendixs)

	for i := 0; i < len(blockAppendixs.BlockAppendixs); i++ {
		fmt.Println("Owner Account: " + blockAppendixs.BlockAppendixs[i].BlockName.OwnerAccount)
		fmt.Println("User Age: " + strconv.Itoa(123))
		fmt.Println("Zhihu Url: " + blockAppendixs.BlockAppendixs[i].AcceptList.Owners[i].OwnerAddress)
		fmt.Println("Weibo Url: " + blockAppendixs.BlockAppendixs[i].BlockHash)
	}
}
