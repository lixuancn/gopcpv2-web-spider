package module

//组件的类型
type Type string

const TYPE_DOWNLOADER Type = "下载器"
const TYPE_ANALYZER Type = "分析器"
const TYPE_PIPELINE Type = "条目处理管道"

var legalTypeLetterMap = map[Type]string{
	TYPE_DOWNLOADER: "D",
	TYPE_ANALYZER: "A",
	TYPE_PIPELINE: "P",
}

//获取首字符的
// typeToLetter 用于根据给定的组件类型获得其字母代号。
// 若给定的组件类型不合法，则第一个结果值会是false。
func typeToLetter(moduleType Type) (bool, string) {
	if v, ok := legalTypeLetterMap[moduleType]; ok{
		return true, v
	}else{
		return false, ""
	}
}

//组件类型是否存在
// LegalType 用于判断给定的组件类型是否合法。
func LegalType(moduleType Type) bool {
	if _, ok := legalTypeLetterMap[moduleType]; ok {
		return true
	}
	return false
}

// getLetter 用于获取组件类型的字母代号。
func getLetter(moduleType Type) (bool, string) {
	if v, ok := legalTypeLetterMap[moduleType]; ok{
		return true, v
	}else{
		return false, ""
	}
}

//代表合法的字母-组件类型的映射。
var legalLetterTypeMap = map[string]Type{
	"D": TYPE_DOWNLOADER,
	"A": TYPE_ANALYZER,
	"P": TYPE_PIPELINE,
}

//用首字母来获取组件类型
func letterToType(letter string) (bool, Type) {
	if v, ok := legalLetterTypeMap[letter]; ok{
		return true, v
	}else{
		return false, ""
	}
}

// GetType 用于获取组件的类型。
// 若给定的组件ID不合法则第一个结果值会是false。
func GetType(mid MID) (bool, Type) {
	parts, err := SplitMID(mid)
	if err != nil {
		return false, ""
	}
	mt, ok := legalLetterTypeMap[parts[0]]
	return ok, mt
}

// CheckType 用于判断组件实例的类型是否匹配。
func CheckType(moduleType Type, module Module) bool {
	if moduleType == "" || module == nil {
		return false
	}
	switch moduleType {
	case TYPE_DOWNLOADER:
		if _, ok := module.(Downloader); ok {
			return true
		}
	case TYPE_ANALYZER:
		if _, ok := module.(Analyzer); ok {
			return true
		}
	case TYPE_PIPELINE:
		if _, ok := module.(Pipeline); ok {
			return true
		}
	}
	return false
}
