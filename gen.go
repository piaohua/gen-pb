package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	Init()
	Gen()
}

//TODO 内部协议通信

var (
	protoPacket = make(map[string]uint32) //响应协议
	protoUnpack = make(map[string]uint32) //请求协议
	packetPath  = "./packet.go"           //打包协议文件路径
	unpackPath  = "./unpack.go"           //解包协议文件路径
	rPacketPath = "./rpacket.go"          //机器人打包协议
	rUnpackPath = "./runpack.go"          //机器人解包协议
	luaPath     = "./MsgID.lua"           //lua文件
	packetFunc  = "Packet"                //打包协议函数名字
	unpackFunc  = "Unpack"                //解包协议函数名字
	rPacketFunc = "Rpacket"               //机器人打包协议函数名字
	rUnpackFunc = "Runpack"               //机器人解包协议函数名字
	jsPath      = "./MsgID.js"            //js文件
	jsonPath    = "./MsgID.json"          //json文件
)

type proto struct {
	name string
	code uint32
}

var protosUnpack = []proto{
	//login
	{name: "CLogin", code: 1000},
	{name: "CRegist", code: 1002},
	{name: "CWxLogin", code: 1004},
}

var protosPacket = []proto{
	//login
	{name: "SLogin", code: 1000},
	{name: "SRegist", code: 1002},
	{name: "SWxLogin", code: 1004},
	{name: "SLoginOut", code: 1006},
}

//初始化
func Init() {
	//request
	for _, v := range protosUnpack {
		registUnpack(v.name, v.code)
	}
	//response
	for _, v := range protosPacket {
		registPacket(v.name, v.code)
	}
	//最后生成MsgID.lua文件
	genMsgID()
	genjsMsgID()
	genjsonMsgID()
}

func registUnpack(key string, code uint32) {
	if _, ok := protoUnpack[key]; ok {
		panic(fmt.Sprintf("%s registered: %d", key, code))
	}
	protoUnpack[key] = code
}

func registPacket(key string, code uint32) {
	if _, ok := protoPacket[key]; ok {
		panic(fmt.Sprintf("%s registered: %d", key, code))
	}
	protoPacket[key] = code
}

//生成文件
func Gen() {
	gen_packet()
	gen_unpack()
	//client
	gen_client_packet()
	gen_client_unpack()
}

//生成打包文件
func gen_packet() {
	var str string
	str += head_packet()
	str += body_packet()
	str += end_packet()
	err := ioutil.WriteFile(packetPath, []byte(str), 0644)
	if err != nil {
		panic(fmt.Sprintf("write file err -> %v\n", err))
	}
}

//生成解包文件
func gen_unpack() {
	var str string
	str += head_unpack()
	str += body_unpack()
	str += end_unpack()
	err := ioutil.WriteFile(unpackPath, []byte(str), 0644)
	if err != nil {
		panic(fmt.Sprintf("write file err -> %v\n", err))
	}
}

func body_unpack() string {
	var str string
	for k, v := range protoUnpack {
		//str += fmt.Sprintf("case %d:\n\t\tmsg := new(%s)\n\t\t%s\n\t", v, k, result_unpack())
		str += fmt.Sprintf("case %d:\n\t\tmsg := new(%s)\n\t\t%s\n\t\t%s\n\t", v, k, body_unpack_code(v), result_unpack())
	}
	return str
}

func body_packet() string {
	var str string
	for k, v := range protoPacket {
		//str += fmt.Sprintf("case *%s:\n\t\tb, err := msg.(*%s).Marshal()\n\t\t%s\n\t", k, k, result_packet(v))
		str += fmt.Sprintf("case *%s:\n\t\t%s\n\t\tb, err := msg.(*%s).Marshal()\n\t\t%s\n\t", k, body_packet_code(v, k), k, result_packet(v))
	}
	return str
}

func body_unpack_code(code uint32) (str string) {
	str = fmt.Sprintf("//msg.Code = %d", code)
	return
}

func body_packet_code(code uint32, name string) (str string) {
	str = fmt.Sprintf("//msg.(*%s).Code = %d", name, code)
	return
}

func head_packet() string {
	return fmt.Sprintf(`// Code generated by tool/gen.go.
// DO NOT EDIT!

package pb

import (
	"errors"
)

//打包消息
func Packet(msg interface{}) (uint32, []byte, error) {
	switch msg.(type) {
	`)
}

func head_unpack() string {
	return fmt.Sprintf(`// Code generated by tool/gen.go.
// DO NOT EDIT!

package pb

import (
	"errors"
)

//解包消息
func Unpack(id uint32, b []byte) (interface{}, error) {
	switch id {
	`)
}

func head_rpacket() string {
	return fmt.Sprintf(`// Code generated by tool/gen.go.
// DO NOT EDIT!

package pb

import (
	"errors"
)

//打包消息
func Rpacket(msg interface{}) (uint32, []byte, error) {
	switch msg.(type) {
	`)
}

func head_runpack() string {
	return fmt.Sprintf(`// Code generated by tool/gen.go.
// DO NOT EDIT!

package pb

import (
	"errors"
)

//解包消息
func Runpack(id uint32, b []byte) (interface{}, error) {
	switch id {
	`)
}

func result_packet(code uint32) string {
	return fmt.Sprintf("return %d, b, err", code)
}

func result_unpack() string {
	return fmt.Sprintf(`err := msg.Unmarshal(b)
		return msg, err`)
}

func end_packet() string {
	return fmt.Sprintf(`default:
		return 0, []byte{}, errors.New("unknown message")
	}
}`)
}

func end_unpack() string {
	return fmt.Sprintf(`default:
		return nil, errors.New("unknown message")
	}
}`)
}

//生成lua文件
func genMsgID() {
	var str string
	str += fmt.Sprintf("msgID = {")
	for k, v := range protoUnpack {
		str += fmt.Sprintf("\n\t%s = %d,", k, v)
	}
	str += fmt.Sprintf("\n")
	for k, v := range protoPacket {
		str += fmt.Sprintf("\n\t%s = %d,", k, v)
	}
	str += fmt.Sprintf("\n}")
	err := ioutil.WriteFile(luaPath, []byte(str), 0666)
	if err != nil {
		panic(fmt.Sprintf("write file err -> %v\n", err))
	}
}

//生成js文件
func genjsMsgID() {
	var str string
	str += fmt.Sprintf("msgID = {")
	for k, v := range protoUnpack {
		str += fmt.Sprintf("\n\t%s : %d,", k, v)
	}
	str += fmt.Sprintf("\n")
	length := len(protoPacket)
	var i int
	for k, v := range protoPacket {
		i += 1
		if i == length {
			str += fmt.Sprintf("\n\t%s : %d", k, v)
		} else {
			str += fmt.Sprintf("\n\t%s : %d,", k, v)
		}
	}
	str += fmt.Sprintf("\n}")
	err := ioutil.WriteFile(jsPath, []byte(str), 0666)
	if err != nil {
		panic(fmt.Sprintf("write file err -> %v\n", err))
	}
}

//
//{
//	3028:{type:"room",        sendType:"protocol.CChat",            revType:"protocol.SChat",           },
//}
func genjsonMsgID() {
	var str string
	str += fmt.Sprintf("{")
	length := len(protoPacket)
	var i int
	for k, v := range protoPacket { //响应
		rsp := ""
		for k2, v2 := range protoUnpack { //请求
			if v == v2 {
				rsp = k2
				break
			}
		}
		i += 1
		if i == length {
			if len(rsp) == 0 {
				str += fmt.Sprintf("\n\t%d:{type:\"room\",\t\tsendType:\"%s\",\t\trevType:\"pb.%s\",\t\t}", v, rsp, k)
			} else {
				str += fmt.Sprintf("\n\t%d:{type:\"room\",\t\tsendType:\"pb.%s\",\t\trevType:\"pb.%s\",\t\t}", v, rsp, k)
			}
		} else {
			if len(rsp) == 0 {
				str += fmt.Sprintf("\n\t%d:{type:\"room\",\t\tsendType:\"%s\",\t\trevType:\"pb.%s\",\t\t},", v, rsp, k)
			} else {
				str += fmt.Sprintf("\n\t%d:{type:\"room\",\t\tsendType:\"pb.%s\",\t\trevType:\"pb.%s\",\t\t},", v, rsp, k)
			}
		}
	}
	str += fmt.Sprintf("\n}")
	err := ioutil.WriteFile(jsonPath, []byte(str), 0666)
	if err != nil {
		panic(fmt.Sprintf("write file err -> %v\n", err))
	}
}

//生成机器人打包文件
func gen_client_packet() {
	var str string
	str += head_rpacket()
	str += body_client_packet()
	str += end_packet()
	err := ioutil.WriteFile(rPacketPath, []byte(str), 0644)
	if err != nil {
		panic(fmt.Sprintf("write file err -> %v\n", err))
	}
}

//生成机器人解包文件
func gen_client_unpack() {
	var str string
	str += head_runpack()
	str += body_client_unpack()
	str += end_unpack()
	err := ioutil.WriteFile(rUnpackPath, []byte(str), 0644)
	if err != nil {
		panic(fmt.Sprintf("write file err -> %v\n", err))
	}
}

func body_client_packet() string {
	var str string
	for k, v := range protoUnpack {
		//str += fmt.Sprintf("case *%s:\n\t\tb, err := msg.(*%s).Marshal()\n\t\t%s\n\t", k, k, result_packet(v))
		str += fmt.Sprintf("case *%s:\n\t\t%s\n\t\tb, err := msg.(*%s).Marshal()\n\t\t%s\n\t", k, body_client_packet_code(v, k), k, result_packet(v))
	}
	return str
}

func body_client_unpack() string {
	var str string
	for k, v := range protoPacket {
		//str += fmt.Sprintf("case %d:\n\t\tmsg := new(%s)\n\t\t%s\n\t", v, k, result_unpack())
		str += fmt.Sprintf("case %d:\n\t\tmsg := new(%s)\n\t\t%s\n\t\t%s\n\t", v, k, body_client_unpack_code(v), result_unpack())
	}
	return str
}

func body_client_unpack_code(code uint32) (str string) {
	str = fmt.Sprintf("//msg.Code = %d", code)
	return
}

func body_client_packet_code(code uint32, name string) (str string) {
	str = fmt.Sprintf("//msg.(*%s).Code = %d", name, code)
	return
}
