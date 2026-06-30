package services

import "sync"

// userLocks 对单个用户的所有权威写入进行序列化，以便
// 并发请求（客户端现在使用多个并行的 HTTPRequest 节点）
// 不会交错执行“检查后变更”的序列——例如，两次种植都读取
// 相同的金币余额并分别扣除它。
//
// 单实例假设：这是一个进程内锁。如果服务器以后
// 扩展到多个实例，请将其替换为基于 Redis 的分布式锁。
var userLocks sync.Map // map[uint]*sync.Mutex

// lockUser 获取每个用户的锁并返回解锁函数。
//
//	defer lockUser(userID)()
func lockUser(userID uint) func() {
	actual, _ := userLocks.LoadOrStore(userID, &sync.Mutex{})
	mu := actual.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}
