package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

type outputEvent struct {
	EventName string
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Printf("init...")
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "set" {
		result, err = set(stub, args)
	} else if fn == "setBlockAppendix" {
		result, err = setBlockAppendix(stub, args)
	} else if fn == "setAcceptList" {
		result, err = setAcceptList(stub, args)
	} else if fn == "setRejectList" {
		result, err = setRejectList(stub, args)
	} else if fn == "get" { // assume 'get' even if fn is nil
		result, err = get(stub, args)
	} else if fn == "getDataOwnerAddress" {
		result, err = getDataOwnerAddress(stub, args)
	} else if fn == "isBlockPermission" {
		result, err = isBlockPermission(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("failed to set asset: %s", args[0])
	}
	event := outputEvent{
		EventName: "set",
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	err = stub.SetEvent("chaincode-event", payload)
	return args[1], nil
}

//初始化区块附录文件
func setBlockAppendix(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	// args["functionName", "ownerAccount"， "blockData"]
	// 所需参数包括： "ownerAccount"， "ownerAddress"， "dataTimestamp"， "blockData"
	// ownerAddress 由 hash(ownerAccount + dataTimeStamp) 生成
	// 总长为3

	if len(args) != 2 {
		return "", fmt.Errorf("incorrect arguments. Expecting a key and a value")
	}

	// 获取时间戳
	dataTimeStamp := strconv.FormatInt(time.Now().Unix(), 10)

	// 生成数据块的哈希值以及用户地址
	blockHash := GetHash(args[2])
	ownerAddress := GetHash(args[1] + dataTimeStamp)
	blocNameHash := GetHash(ownerAddress + blockHash)

	// 数据所有者的账户名和数据块生成的时间戳组合为 blockName,
	blockName := BlockName{
		OwnerAccount:  args[1],
		DataTimestamp: dataTimeStamp,
		BlockNameHash: blocNameHash,
	}

	owner := Owner{
		OwnerAccount: args[1],
		OwnerAddress: ownerAddress,
	}

	// 打包为附录文件
	blockAppendix := BlockAppendix{
		BlockName:  blockName,
		BlockHash:  blockHash,
		Owner:      owner,
		AcceptList: []string{ownerAddress},
		RejectList: []string{},
	}
	// 将附录文件对象序列化
	blockAppendixJSON, err := json.Marshal(blockAppendix)
	if err != nil {
		return "", err
	}

	err = stub.PutState(blocNameHash, blockAppendixJSON)
	if err != nil {
		return "", fmt.Errorf("failed to set asset: %s", args[0])
	}
	event := outputEvent{
		EventName: "BlockAppendix",
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	err = stub.SetEvent("BlockAppendix", payload)
	if err != nil {
		return "", err
	}
	return blocNameHash, nil
}

// 设置AcceptList
func setAcceptList(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	// args["functionName", "blocNameHash", "ownerAddress"]	其中ownerAddress为被验证用户的地址
	if len(args) != 2 {
		return "", fmt.Errorf("incorrect arguments. Expecting a key and a value")
	}

	// 获取附录文件
	blockAppendixJSON, err := stub.GetState(args[1])
	//owner = Owner()

	if err != nil {
		return "", err
	} else if blockAppendixJSON == nil {
		return "", fmt.Errorf("the Appendix of blockHash %s doesn't exist", args[1])
	}

	blockAppendix := new(BlockAppendix)
	// 反序列化
	err = json.Unmarshal(blockAppendixJSON, blockAppendix)
	if err != nil {
		return "", err
	}
	blockAppendix.AcceptList = append(blockAppendix.AcceptList, args[2])

	blockAppendixJSON, err = json.Marshal(blockAppendix)
	if err != nil {
		return "", err
	}

	err = stub.PutState(args[1], blockAppendixJSON)
	if err != nil {
		return "", fmt.Errorf("failed to set asset: %s", args[0])
	}
	event := outputEvent{
		EventName: "setAcceptList",
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	err = stub.SetEvent("AcceptList", payload)
	return fmt.Sprintf("setAcceptList add %s successful.", args[1]), nil

}

//设置RejectList
func setRejectList(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	//args["functionName", "blocNameHash", "ownerAddress"]
	if len(args) != 2 {
		return "", fmt.Errorf("incorrect arguments. Expecting a key and a value")
	}

	// 获取附录文件
	blockAppendixJSON, err := stub.GetState(args[1])

	if err != nil {
		return "", err
	} else if blockAppendixJSON == nil {
		return "", fmt.Errorf("the Appendix of blockHash %s doesn't exist", args[1])
	}

	blockAppendix := new(BlockAppendix)
	// 反序列化
	err = json.Unmarshal(blockAppendixJSON, blockAppendix)
	if err != nil {
		return "", err
	}
	blockAppendix.RejectList = append(blockAppendix.RejectList, args[2])

	blockAppendixJSON, err = json.Marshal(blockAppendix)
	if err != nil {
		return "", err
	}

	err = stub.PutState(args[1], blockAppendixJSON)
	if err != nil {
		return "", fmt.Errorf("failed to set asset: %s", args[0])
	}
	event := outputEvent{
		EventName: "setRejectList",
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	err = stub.SetEvent("RejectList", payload)
	return args[1], nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("asset not found: %s", args[0])
	}
	return string(value), nil
}

//获取数据拥有者
func getDataOwnerAddress(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	//args["functionName", "blocNameHash"]
	if len(args) != 1 {
		return "", fmt.Errorf("incorrect arguments. Expecting a key")
	}
	// 获取附录文件
	blockAppendixJSON, err := stub.GetState(args[1])

	if err != nil {
		return "", err
	} else if blockAppendixJSON == nil {
		return "", fmt.Errorf("the Appendix of blockHash %s doesn't exist", args[1])
	}

	blockAppendix := new(BlockAppendix)
	// 反序列化
	err = json.Unmarshal(blockAppendixJSON, blockAppendix)
	if err != nil {
		return "", err
	}

	dataUser := blockAppendix.Owner.OwnerAddress
	return dataUser, nil
}

func GetHash(data string) string {
	_sha1 := sha256.New()
	_sha1.Write([]byte(data))
	return hex.EncodeToString(_sha1.Sum([]byte("")))
}

//获取权限设置
func isBlockPermission(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	//args["functionName", "Owners"]
	if len(args) != 2 {
		return "", fmt.Errorf("incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("asset not found: %s", args[0])
	}

	var blockAppendixs BlockAppendixs

	err = json.Unmarshal(value, &blockAppendixs)
	if err != nil {
		return "", err
	}

	var acceptlist []string
	var rejectList []string

	for i := 0; i < len(blockAppendixs.BlockAppendixs); i++ {
		acceptlist = blockAppendixs.BlockAppendixs[i].AcceptList
		rejectList = blockAppendixs.BlockAppendixs[i].RejectList
	}

	for i := 0; i < len(acceptlist); {
		if args[1] == acceptlist[i] {
			return strconv.Itoa(1), nil
		}
	}

	for i := 0; i < len(rejectList); {
		if args[1] == rejectList[i] {
			return strconv.Itoa(0), nil
		}
	}
	return "not found", nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
