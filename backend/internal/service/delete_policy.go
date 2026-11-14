package service

// DeleteDeckMode 描述一次删除请求的实际执行方式。
type DeleteDeckMode string

const (
	// DeleteHard 彻底删除题库及其全部关联数据（含市场版本）。
	DeleteHard DeleteDeckMode = "hard"
	// DeleteArchived 仅把题库从「我的题库」移除，公开版本继续留在市场。
	DeleteArchived DeleteDeckMode = "archived"
)

// PlanDeckDeletion 纯函数：
//   - 已归档题库再次删除时，一律彻底删除（这是「清除市场保留版」的入口）；
//   - 私有题库没有市场版本，一律彻底删除；
//   - 公开题库且调用方选择「保留市场版本」时转为归档，否则彻底删除。
func PlanDeckDeletion(visibility string, archived, removeFromMarket bool) DeleteDeckMode {
	if archived || visibility != "public" || removeFromMarket {
		return DeleteHard
	}
	return DeleteArchived
}
