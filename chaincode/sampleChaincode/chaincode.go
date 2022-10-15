package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	AppendixObj = "appendix"
	DataInfoObj = "dataInfo"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// BlockName 账户名和时间戳组合为 blockName
type BlockName struct {
	OwnerAccount  string `json:"owner_account"`  // 账户名
	DataTimestamp int    `json:"data_timestamp"` // 时间戳
}

// BlockAppendix 附录文件
type BlockAppendix struct {
	BlockName    BlockName `json:"block_name"`
	BlockHash    string    `json:"block_hash"`    // 数据块的哈希值
	OwnerAddress string    `json:"owner_address"` // 数据所有者的账户地址
}

// BlockDataInfo 数据块信息
type BlockDataInfo struct {
	BlockData  string   `json:"block_data"`  // 数据块
	AcceptList []string `json:"accept_list"` // 接受列表
	RejectList []string `json:"reject_list"` // 拒绝列表
}

// UploadBlockAppendix 上传附录文件
// 将 blockName 与加密数据块的哈希值、数据所有者的账户地址打包到附录文件中,最后将附录文件上传到区块链上
// 注： 论文中未说明 owner_address 如何得到，故在此设定为从输入得到
// 参数： 账户名、账户地址、时间戳、数据块
// 返回： 数据块哈希（后期可去除）
func (s *SmartContract) UploadBlockAppendix(ctx contractapi.TransactionContextInterface, ownerAccount string, ownerAddress string, dataTimestamp int, blockData string) (string, error) {
	// 数据所有者的账户名和数据块生成的时间戳组合为 blockName,
	blockName := BlockName{
		OwnerAccount:  ownerAccount,
		DataTimestamp: dataTimestamp,
	}
	// 计算数据块的哈希值
	blockHash := GetHash(blockData)
	// 打包为附录文件
	blockAppendix := BlockAppendix{
		BlockName:    blockName,
		BlockHash:    blockHash,
		OwnerAddress: ownerAddress,
	}
	// 将附录文件对象序列化
	blockAppendixJSON, err := json.Marshal(blockAppendix)
	if err != nil {
		return "", err
	}

	// 创建数据块信息
	blockDataInfo := BlockDataInfo{blockData, []string{}, []string{}}
	// 将数据块信息对象序列化
	blockDataInfoJSON, err := json.Marshal(blockDataInfo)
	if err != nil {
		return "", err
	}
	// 构造附录文件的主键
	aKey := ConstructKey(AppendixObj, blockHash)
	// 将附录文件写入区块链
	err = ctx.GetStub().PutState(aKey, blockAppendixJSON)
	if err != nil {
		return "", err
	}
	// 构造数据块信息的主键
	dKey := ConstructKey(DataInfoObj, blockHash)
	// 将数据块信息写入区块链
	err = ctx.GetStub().PutState(dKey, blockDataInfoJSON)
	if err != nil {
		return "", err
	}
	return blockHash, nil
}

// GetOwnerAddress 查询数据所有者地址
// 根据数据块的哈希值在区块链上的附录文件中搜索出对应的数据所有者地址。
// 参数： 数据块哈希值
func (s *SmartContract) GetOwnerAddress(ctx contractapi.TransactionContextInterface, blockHash string) (string, error) {
	// 构造主键
	key := ConstructKey(AppendixObj, blockHash)
	// 从区块链中获取该数据块对应的附录文件
	blockAppendixJSON, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", err
	} else if blockAppendixJSON == nil {
		return "", fmt.Errorf("the Appendix of blockHash %s doesn't exist", blockHash)
	}
	// 反序列化附录文件对象
	blockAppendix := new(BlockAppendix)
	err = json.Unmarshal(blockAppendixJSON, blockAppendix)
	if err != nil {
		return "", err
	}
	// 返回数据所有者的地址
	return blockAppendix.OwnerAddress, nil
}

// UploadKLicenseFile 上传许可文件
// 如果消费者被授予访问权限,则其账户地址将记录在 acceptList 中;如果访问权限被拒绝,则消费者将被添加至 rejectList
// 参数： 数据块哈希值、账户地址、访问权限（0或1）
func (s *SmartContract) UploadKLicenseFile(ctx contractapi.TransactionContextInterface, blockHash string, accountAddress string, blockPermission int) error {
	// 构造主键
	key := ConstructKey(DataInfoObj, blockHash)
	// 从区块链中获取该数据块对应的数据块信息
	blockDataInfoJSON, err := ctx.GetStub().GetState(key)
	if err != nil {
		return err
	} else if blockDataInfoJSON == nil {
		return fmt.Errorf("the DataInfo of blockHash %s doesn't exist", blockHash)
	}
	blockDataInfo := new(BlockDataInfo)
	// 反序列化
	err = json.Unmarshal(blockDataInfoJSON, blockDataInfo)
	if err != nil {
		return err
	}

	// 判断权限，并将账户地址记录到对应的列表中
	if blockPermission == 1 {
		blockDataInfo.AcceptList = append(blockDataInfo.AcceptList, accountAddress)
	} else {
		blockDataInfo.RejectList = append(blockDataInfo.RejectList, accountAddress)
	}
	// 序列化对象
	blockDataInfoJSON, err = json.Marshal(blockDataInfo)
	if err != nil {
		return err
	}
	// 写入区块链
	return ctx.GetStub().PutState(key, blockDataInfoJSON)
}

// VerifyAccess 验证访问权限
// 通过搜索消费者是否已被记录在对应数据块的 acceptList 来验证其访问权限
// 参数： 数据块哈希值、消费者账户地址
func (s *SmartContract) VerifyAccess(ctx contractapi.TransactionContextInterface, blockHash string, accountAddress string) (bool, error) {
	// 构造主键
	key := ConstructKey(DataInfoObj, blockHash)
	// 从区块链中获取该数据块对应的数据块信息
	blockDataInfoJSON, err := ctx.GetStub().GetState(key)
	if err != nil {
		return false, err
	} else if blockDataInfoJSON == nil {
		return false, fmt.Errorf("blockHash %s doesn't exist", blockHash)
	}
	blockDataInfo := new(BlockDataInfo)
	// 反序列化
	err = json.Unmarshal(blockDataInfoJSON, &blockDataInfo)
	if err != nil {
		return false, err
	}

	// 验证消费者账户地址是否已被记录在对应数据块的 acceptList
	return IsContain(accountAddress, blockDataInfo.AcceptList), nil
}

// VerifyHashValue 验证哈希值
// 计算加密数据块的哈希值,并将计算结果与记录在区块链上的哈希标识符进行对比,验证数据块的正确性和完整性。
// 参数： 数据块
func (s *SmartContract) VerifyHashValue(ctx contractapi.TransactionContextInterface, blockData string) (bool, error) {
	// 计算数据块的哈希值
	blockHash := GetHash(blockData)
	// 构造主键
	key := ConstructKey(AppendixObj, blockHash)
	// 从区块链中获取该数据块对应的附录文件
	blockAppendixJSON, err := ctx.GetStub().GetState(key)
	if err != nil {
		return false, err
	} else if blockAppendixJSON == nil {
		return false, fmt.Errorf("blockHash %s doesn't exist", blockHash)
	}
	// 反序列化附录文件对象
	blockAppendix := new(BlockAppendix)
	err = json.Unmarshal(blockAppendixJSON, blockAppendix)
	if err != nil {
		return false, err
	}

	// 验证哈希值是否一致
	return blockAppendix.BlockHash == blockHash, nil
}

// GetHash 获取哈希值
func GetHash(data string) string {
	_sha1 := sha256.New()
	_sha1.Write([]byte(data))
	return hex.EncodeToString(_sha1.Sum([]byte("")))
}

// IsContain 判断某个字符串是否在字符串切片中
func IsContain(str string, slice []string) bool {
	for _, v := range slice {
		if str == v {
			return true
		}
	}
	return false
}

// ConstructKey 构造主键
func ConstructKey(objType string, value string) string {
	return fmt.Sprintf("%s_%s", objType, value)
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}
