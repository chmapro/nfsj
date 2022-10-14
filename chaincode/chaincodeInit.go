package main

type BlockAppendixs struct {
	BlockAppendixs []BlockAppendix `json:"blockAppendix"`
}

type BlockAppendix struct {
	BlockName  BlockName `json:"blockName"`
	Owner      Owner     `json:"owner"`
	BlockHash  string    `json:"blockHash"`
	AcceptList []string  `json:"accept_list"`
	RejectList []string  `json:"reject_list"`
}

type BlockName struct {
	OwnerAccount  string `json:"owner_account"`
	DataTimestamp string `json:"data_timestamp"`
	BlockNameHash string `json:"block_name_hash"`
}

type Owner struct {
	OwnerAccount string `json:"owner_account"`
	OwnerAddress string `json:"owner_address"`
}

//type AcceptList struct {
//	Owners []Owner `json:"accept_owner"`
//}
//
//type RejectList struct {
//	Owners []Owner `json:"reject_owner"`
//}
