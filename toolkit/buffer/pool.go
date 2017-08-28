package buffer

type Pool interface {
	//统一的缓冲器的容量
	BufferCap() uint32
	//缓冲器的最大数量
	MaxBufferNumber() uint32
	//缓冲器的数量
	BufferNumber() uint32
	//缓冲池中的数据总数
	Total() uint64
	//入，阻塞的
	Put(interface{}) error
	//取，阻塞的
	Get(interface{}, error)
	Close() bool
	Closed() bool

}