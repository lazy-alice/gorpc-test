package msg

type Request struct {
	ServicePath string            //请求的服务路径
	Metadata    map[string][]byte //透传的数据
	Payload     []byte            //请求体
}

type Response struct {
	RetCode  uint32            //返回码 0-正常 非0-错误
	RetMsg   string            //返回消息 success-正常 其他为错误详情
	Metadata map[string][]byte //透传的数据
	Payload  []byte
} //返回体
