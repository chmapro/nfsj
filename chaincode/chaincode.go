package main

import (
	"encoding/json"
	"fmt"

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
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
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
	//args["functionName", "blockappendixs", "json"]
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	event := outputEvent{
		EventName: "BlockAppendix",
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	err = stub.SetEvent("BlockAppendix", payload)
	return args[1], nil
}

//设置AcceptList
func setAcceptList(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	//args["functionName", "AcceptList", "list"]
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	event := outputEvent{
		EventName: "setAcceptList",
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	err = stub.SetEvent("AcceptList", payload)
	return args[1], nil
}

//设置RejectList
func setRejectList(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	//args["functionName", "RejectList", "list"]
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
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
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return string(value), nil
}

//获取数据拥有者
func getDataOwnerAddress(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	//args["functionName", "blockname", "DataTimestamp"]
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}

	var blockAppendixs BlockAppendixs

	json.Unmarshal(value, &blockAppendixs)

	for i := 0; i < len(blockAppendixs.BlockAppendixs); i++ {
		if args[1] == blockAppendixs.BlockAppendixs[i].BlockName.OwnerAccount {
			if args[2] == blockAppendixs.BlockAppendixs[i].BlockName.DataTimestamp {
				return blockAppendixs.BlockAppendixs[i].Owner.OwnerAccount, nil
			}
		}
	}

	return "", nil
}

//获取权限设置
func isBlockPermission(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	//args["functionName", "Owners"]
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}

	var blockAppendixs BlockAppendixs

	json.Unmarshal(value, &blockAppendixs)

	var acceptlist []Owner
	var rejectList []Owner

	for i := 0; i < len(blockAppendixs.BlockAppendixs); i++ {
		acceptlist = blockAppendixs.BlockAppendixs[i].AcceptList.Owners
		rejectList = blockAppendixs.BlockAppendixs[i].RejectList.Owners
	}

	for i := 0; i < len(acceptlist); {
		if args[1] == acceptlist[i].OwnerAddress {
			return string(1), nil
		}
	}

	for i := 0; i < len(rejectList); {
		if args[1] == rejectList[i].OwnerAddress {
			return string(0), nil
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
