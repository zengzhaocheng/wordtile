package database

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/wordtile/backend/internal/model"
)

// SeedDemo 首次启动注入演示账号与示例公开题库（幂等）。
func SeedDemo(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.User{}).Where("username = ?", "demo").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("demo1234"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	demo := model.User{
		ID:           uuid.New(),
		Username:     "demo",
		Email:        "demo@wordtile.app",
		PasswordHash: string(hash),
		DisplayName:  "演示账号",
	}
	if err := db.Create(&demo).Error; err != nil {
		return err
	}

	entries := seedEntries()
	deck := model.Deck{
		ID:          uuid.New(),
		OwnerID:     demo.ID,
		Name:        "基础英语词汇（示例）",
		Description: "内置示例题库：数字 / 食物 / 旅游 / 动物 / 家庭，共 " + itoa(len(entries)) + " 条。可直接练习、fork 后编辑。",
		Visibility:  "public",
	}
	if err := db.Create(&deck).Error; err != nil {
		return err
	}

	changes := model.EntryChanges{Added: make([]model.EntryChange, 0, len(entries))}
	for _, e := range entries {
		logical := uuid.New()
		if err := db.Create(&model.Entry{
			ID:         uuid.New(),
			DeckID:     deck.ID,
			LogicalID:  logical,
			Term:       e.term,
			Definition: e.definition,
			Tags:       e.tags,
		}).Error; err != nil {
			return err
		}
		changes.Added = append(changes.Added, model.EntryChange{
			ID: logical, Term: e.term, Definition: e.definition, Tags: e.tags,
		})
	}

	hashStr := uuid.NewString()
	commit := model.Commit{
		ID:           uuid.New(),
		DeckID:       deck.ID,
		Hash:         hashStr,
		Seq:          1,
		ParentHashes: []string{},
		AuthorID:     demo.ID,
		Changes:      changes,
		Message:      "初始提交：导入示例词条",
	}
	if err := db.Create(&commit).Error; err != nil {
		return err
	}
	deck.HeadCommitID = &hashStr
	return db.Save(&deck).Error
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [8]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

type seedEntry struct {
	term       string
	definition string
	tags       []string
}

func seedEntries() []seedEntry {
	return []seedEntry{
		{"一", "one", []string{"数字"}},
		{"二", "two", []string{"数字"}},
		{"三", "three", []string{"数字"}},
		{"四", "four", []string{"数字"}},
		{"五", "five", []string{"数字"}},
		{"六", "six", []string{"数字"}},
		{"七", "seven", []string{"数字"}},
		{"八", "eight", []string{"数字"}},
		{"九", "nine", []string{"数字"}},
		{"十", "ten", []string{"数字"}},
		{"苹果", "apple", []string{"食物"}},
		{"面包", "bread", []string{"食物"}},
		{"米饭", "rice", []string{"食物"}},
		{"鸡蛋", "egg", []string{"食物"}},
		{"牛奶", "milk", []string{"食物"}},
		{"鱼", "fish", []string{"食物", "动物"}},
		{"面条", "noodle", []string{"食物"}},
		{"鸡肉", "chicken", []string{"食物"}},
		{"奶酪", "cheese", []string{"食物"}},
		{"葡萄", "grape", []string{"食物"}},
		{"旅行", "travel", []string{"旅游"}},
		{"酒店", "hotel", []string{"旅游"}},
		{"机场", "airport", []string{"旅游"}},
		{"机票", "ticket", []string{"旅游"}},
		{"护照", "passport", []string{"旅游"}},
		{"地图", "map", []string{"旅游"}},
		{"海滩", "beach", []string{"旅游"}},
		{"博物馆", "museum", []string{"旅游"}},
		{"行李", "luggage", []string{"旅游"}},
		{"签证", "visa", []string{"旅游"}},
		{"猫", "cat", []string{"动物"}},
		{"狗", "dog", []string{"动物"}},
		{"鸟", "bird", []string{"动物"}},
		{"老虎", "tiger", []string{"动物"}},
		{"大象", "elephant", []string{"动物"}},
		{"猴子", "monkey", []string{"动物"}},
		{"熊猫", "panda", []string{"动物"}},
		{"马", "horse", []string{"动物"}},
		{"家庭", "family", []string{"家庭"}},
		{"父亲", "father", []string{"家庭"}},
		{"母亲", "mother", []string{"家庭"}},
		{"兄弟", "brother", []string{"家庭"}},
		{"姐妹", "sister", []string{"家庭"}},
		{"爷爷", "grandfather", []string{"家庭"}},
		{"奶奶", "grandmother", []string{"家庭"}},
		{"家", "home", []string{"家庭"}},
	}
}
