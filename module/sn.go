package module

type SNGenertor interface {
	//最小序列号
	Start() uint64
	Max() uint64
	//下一个序列号
	Next()uint64
	//用户获取循环次数，即从前两名方法给定范围的循环次数
	CycleCount() uint64
	//获取一个序列号并准备下一个序列号
	Get()uint64
}