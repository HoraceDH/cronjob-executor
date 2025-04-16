package webresult

// Constants 文件描述
// Created in 2025-03-13 13:18.
// @author Horace

var SUCCESS = MsgObject{Code: 0, Msg: ""}
var ERROR_SIGN = MsgObject{Code: 6, Msg: "非法请求！"}
var ERROR_EXECUTE_SHUTDOWN = MsgObject{Code: 15, Msg: "执行器已关闭"}
var ERROR = MsgObject{Code: 1000, Msg: "操作失败"}
var ERROR_PARAMS = MsgObject{Code: 1001, Msg: "参数错误"}

// MsgObject 消息对象
type MsgObject struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Success 成功对象
func Success(data interface{}) MsgObject {
	var msgObject = new(MsgObject)
	msgObject.Code = SUCCESS.Code
	msgObject.Msg = SUCCESS.Msg
	msgObject.Data = data
	return *msgObject
}
