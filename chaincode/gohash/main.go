package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

/*
  使用MD5对数据进行哈希运算方法1：使用md5.Sum()方法
*/
func getMD5String_1(b []byte) (result string) {
	//给哈希算法添加数据
	res := md5.Sum(b) //返回值：[Size]byte 数组
	/*//方法1：
	result=fmt.Sprintf("%x",res)   //通过fmt.Sprintf()方法格式化数据
	*/
	//方法2：
	result = hex.EncodeToString(res[:]) //对应的参数为：切片，需要将数组转换为切片。
	return
}

/*
使用MD5对数据进行哈希运算方法2：使用md5.new()方法
*/
func getMD5String_2(b []byte) (result string) {
	//1、创建Hash接口
	myHash := md5.New() //返回 Hash interface
	//2、添加数据
	myHash.Write(b) //写入数据
	//3、计算结果
	/*
	  执行原理为：myHash.Write(b1)写入的数据进行hash运算  +  myHash.Sum(b2)写入的数据进行hash运算
	              结果为：两个hash运算结果的拼接。若myHash.Write()省略或myHash.Write(nil) ，则默认为写入的数据为“”。
	              根据以上原理，一般不采用两个hash运算的拼接，所以参数为nil
	*/
	res := myHash.Sum(nil) //进行运算
	//4、数据格式化
	result = hex.EncodeToString(res) //转换为string
	return
}
func main() {
	str := []byte("jiangzhou")
	res := getMD5String_1(str)
	fmt.Println("方法1运算结果：", res)

	res = getMD5String_2(str)
	fmt.Println("方法2运算结果：", res)

}
