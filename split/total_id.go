package split

type TotalId struct {
	Split
	Million int
}

// 创建分表
func (f *TotalId) CreateSplitTable() error {
	return nil
}

// 同步数据
func (f *TotalId) SyncData() error {
	return nil
}
