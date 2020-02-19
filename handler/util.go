package handler

import (
	"encoding/json"
	"net/http"
)

// 对于 POST/DELETE/PUT 等写请求，handler 需要将请求结果（由 code 和 msg 组成，绝大多数情况下无 data）返回给用户
func writeJsonResponse(w http.ResponseWriter, jsonResponse JsonResponse) {
	bytes, _ := json.Marshal(jsonResponse)
	w.Write(bytes)
}

// 对于 GET 类读请求，如果读取成功，则直接返回读取到的数据
func writeJsonData(w http.ResponseWriter, data interface{}) {
	var jsonResponse = JsonResponse{
		Code: 0,
		Msg:  "",
		Data: data,
	}
	bytes, _ := json.Marshal(jsonResponse)
	w.Write(bytes)
}
