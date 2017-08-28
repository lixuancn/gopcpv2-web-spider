package buffer

type Buffer interface {
	Cap() uint32
	Len() uint32
	//非阻塞
	Put(interface{})error
	//非阻塞
	Get(interface{}, error)
	Close()
	Closed()
}