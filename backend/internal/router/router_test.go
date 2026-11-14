package router

import (
	"testing"

	"github.com/wordtile/backend/internal/config"
)

// 静态路由 /decks/import 与参数路由 /decks/:id 共存时，gin 旧版本会 panic。
// 此测试保证路由树可以正常构建，且三个导入导出端点都已注册。
func TestImportExportRoutesRegistered(t *testing.T) {
	r := New(&Handlers{}, &config.Config{CORSOrigins: "*", JWTSecret: "test-secret"})
	want := map[string]bool{
		"POST /api/v1/decks/import":     false,
		"POST /api/v1/decks/:id/import": false,
		"GET /api/v1/decks/:id/export":  false,
	}
	for _, ri := range r.Routes() {
		key := ri.Method + " " + ri.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for path, found := range want {
		if !found {
			t.Errorf("路由未注册: %s", path)
		}
	}
}
