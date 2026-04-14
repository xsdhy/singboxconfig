package storage

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"singboxconfig/entity"
	"strings"
	"testing"
	"time"
)

// mockPostgREST 创建一个模拟 PostgREST 的 HTTP 服务器
func mockPostgREST(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *SupabaseStorage) {
	t.Helper()
	server := httptest.NewServer(handler)
	// NewSupabaseStorage 会追加 /rest/v1，所以 baseURL 需要去掉这个后缀
	store := &SupabaseStorage{
		baseURL:    server.URL,
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}
	return server, store
}

// assertAuthHeaders 验证请求携带了正确的认证 Header
func assertAuthHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Header.Get("apikey") != "test-api-key" {
		t.Errorf("expected apikey header 'test-api-key', got %q", r.Header.Get("apikey"))
	}
	if r.Header.Get("Authorization") != "Bearer test-api-key" {
		t.Errorf("expected Authorization header 'Bearer test-api-key', got %q", r.Header.Get("Authorization"))
	}
}

// --- Subscribe Tests ---

func TestSupabaseStorage_CreateSubscribe(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuthHeaders(t, r)
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/subscribes") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Prefer") != "return=representation" {
			t.Errorf("expected Prefer header, got %q", r.Header.Get("Prefer"))
		}
		var sb supabaseSubscribe
		json.NewDecoder(r.Body).Decode(&sb)
		if sb.Name != "test_sub" || sb.URL != "https://example.com" {
			t.Errorf("unexpected body: %+v", sb)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode([]supabaseSubscribe{sb})
	})
	defer server.Close()

	err := store.CreateSubscribe(&entity.Subscribe{
		Name: "test_sub",
		URL:  "https://example.com",
	})
	if err != nil {
		t.Fatalf("CreateSubscribe failed: %v", err)
	}
}

func TestSupabaseStorage_CreateSubscribe_Conflict(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	})
	defer server.Close()

	err := store.CreateSubscribe(&entity.Subscribe{Name: "dup"})
	if err == nil {
		t.Fatal("expected error for conflict")
	}
}

func TestSupabaseStorage_GetSubscribe(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuthHeaders(t, r)
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.RawQuery, "name=eq.test_sub") {
			t.Errorf("expected name filter in query, got %s", r.URL.RawQuery)
		}
		if r.Header.Get("Accept") != "application/vnd.pgrst.object+json" {
			t.Errorf("expected object Accept header, got %q", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(supabaseSubscribe{
			Name:      "test_sub",
			URL:       "https://example.com",
			UserAgent: "ua",
			Status:    true,
		})
	})
	defer server.Close()

	sub, err := store.GetSubscribe("test_sub")
	if err != nil {
		t.Fatalf("GetSubscribe failed: %v", err)
	}
	if sub.Name != "test_sub" || sub.URL != "https://example.com" || sub.UserAgent != "ua" || !sub.Status {
		t.Errorf("unexpected subscribe: %+v", sub)
	}
}

func TestSupabaseStorage_GetSubscribe_NotFound(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotAcceptable)
	})
	defer server.Close()

	_, err := store.GetSubscribe("nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSupabaseStorage_ListSubscribes(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuthHeaders(t, r)
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]supabaseSubscribe{
			{Name: "sub1", URL: "https://a.com"},
			{Name: "sub2", URL: "https://b.com", Status: true},
		})
	})
	defer server.Close()

	subs, err := store.ListSubscribes()
	if err != nil {
		t.Fatalf("ListSubscribes failed: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("expected 2 subscribes, got %d", len(subs))
	}
	if subs[0].Name != "sub1" || subs[1].Name != "sub2" {
		t.Errorf("unexpected subscribes: %+v", subs)
	}
}

func TestSupabaseStorage_ListSubscribes_Empty(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	})
	defer server.Close()

	subs, err := store.ListSubscribes()
	if err != nil {
		t.Fatalf("ListSubscribes failed: %v", err)
	}
	if len(subs) != 0 {
		t.Errorf("expected empty list, got %d", len(subs))
	}
}

func TestSupabaseStorage_UpdateSubscribe(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuthHeaders(t, r)
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if !strings.Contains(r.URL.RawQuery, "name=eq.test_sub") {
			t.Errorf("expected name filter, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]supabaseSubscribe{{Name: "test_sub", URL: "https://new.com"}})
	})
	defer server.Close()

	err := store.UpdateSubscribe(&entity.Subscribe{Name: "test_sub", URL: "https://new.com"})
	if err != nil {
		t.Fatalf("UpdateSubscribe failed: %v", err)
	}
}

func TestSupabaseStorage_UpdateSubscribe_NotFound(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	})
	defer server.Close()

	err := store.UpdateSubscribe(&entity.Subscribe{Name: "nonexistent"})
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSupabaseStorage_DeleteSubscribe(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuthHeaders(t, r)
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]supabaseSubscribe{{Name: "test_sub"}})
	})
	defer server.Close()

	err := store.DeleteSubscribe("test_sub")
	if err != nil {
		t.Fatalf("DeleteSubscribe failed: %v", err)
	}
}

func TestSupabaseStorage_DeleteSubscribe_NotFound(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	})
	defer server.Close()

	err := store.DeleteSubscribe("nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// --- NodeGroup Tests ---

func TestSupabaseStorage_CreateNodeGroup(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuthHeaders(t, r)
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/node_groups") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var sg supabaseNodeGroup
		json.NewDecoder(r.Body).Decode(&sg)
		if sg.Tag != "proxy" || sg.GroupType != "selector" {
			t.Errorf("unexpected body: %+v", sg)
		}
		w.WriteHeader(http.StatusCreated)
	})
	defer server.Close()

	err := store.CreateNodeGroup(&entity.NodeGroup{Tag: "proxy", GroupType: "selector"})
	if err != nil {
		t.Fatalf("CreateNodeGroup failed: %v", err)
	}
}

func TestSupabaseStorage_GetNodeGroup(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "tag=eq.proxy") {
			t.Errorf("expected tag filter, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(supabaseNodeGroup{
			Tag: "proxy", Name: "Proxy", GroupType: "selector", TestURL: "https://test.com",
		})
	})
	defer server.Close()

	ng, err := store.GetNodeGroup("proxy")
	if err != nil {
		t.Fatalf("GetNodeGroup failed: %v", err)
	}
	if ng.Tag != "proxy" || ng.GroupType != "selector" || ng.TestURL != "https://test.com" {
		t.Errorf("unexpected node group: %+v", ng)
	}
}

func TestSupabaseStorage_GetNodeGroup_NotFound(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotAcceptable)
	})
	defer server.Close()

	_, err := store.GetNodeGroup("nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSupabaseStorage_ListNodeGroups(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]supabaseNodeGroup{
			{Tag: "proxy", Name: "Proxy"},
			{Tag: "direct", Name: "Direct"},
		})
	})
	defer server.Close()

	groups, err := store.ListNodeGroups()
	if err != nil {
		t.Fatalf("ListNodeGroups failed: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2, got %d", len(groups))
	}
}

func TestSupabaseStorage_UpdateNodeGroup(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]supabaseNodeGroup{{Tag: "proxy"}})
	})
	defer server.Close()

	err := store.UpdateNodeGroup(&entity.NodeGroup{Tag: "proxy", Name: "Updated"})
	if err != nil {
		t.Fatalf("UpdateNodeGroup failed: %v", err)
	}
}

func TestSupabaseStorage_DeleteNodeGroup(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]supabaseNodeGroup{{Tag: "proxy"}})
	})
	defer server.Close()

	err := store.DeleteNodeGroup("proxy")
	if err != nil {
		t.Fatalf("DeleteNodeGroup failed: %v", err)
	}
}

// --- RuleSet Tests ---

func TestSupabaseStorage_CreateRuleSet(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuthHeaders(t, r)
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/rule_sets") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var sr supabaseRuleSet
		json.NewDecoder(r.Body).Decode(&sr)
		if sr.Tag != "geoip-cn" || sr.RuleSetType != "remote" {
			t.Errorf("unexpected body: %+v", sr)
		}
		w.WriteHeader(http.StatusCreated)
	})
	defer server.Close()

	err := store.CreateRuleSet(&entity.RuleSet{Tag: "geoip-cn", RuleSetType: "remote"})
	if err != nil {
		t.Fatalf("CreateRuleSet failed: %v", err)
	}
}

func TestSupabaseStorage_GetRuleSet(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "tag=eq.geoip-cn") {
			t.Errorf("expected tag filter, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(supabaseRuleSet{
			Tag: "geoip-cn", Name: "GeoIP CN", RuleSetType: "remote",
			Format: "binary", Outbound: "direct", Sort: 10,
		})
	})
	defer server.Close()

	rs, err := store.GetRuleSet("geoip-cn")
	if err != nil {
		t.Fatalf("GetRuleSet failed: %v", err)
	}
	if rs.Tag != "geoip-cn" || rs.RuleSetType != "remote" || rs.Sort != 10 {
		t.Errorf("unexpected rule set: %+v", rs)
	}
}

func TestSupabaseStorage_GetRuleSet_NotFound(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotAcceptable)
	})
	defer server.Close()

	_, err := store.GetRuleSet("nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSupabaseStorage_ListRuleSets(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]supabaseRuleSet{
			{Tag: "geoip-cn", Sort: 1},
			{Tag: "geosite-cn", Sort: 2},
		})
	})
	defer server.Close()

	sets, err := store.ListRuleSets()
	if err != nil {
		t.Fatalf("ListRuleSets failed: %v", err)
	}
	if len(sets) != 2 {
		t.Fatalf("expected 2, got %d", len(sets))
	}
}

func TestSupabaseStorage_UpdateRuleSet(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]supabaseRuleSet{{Tag: "geoip-cn"}})
	})
	defer server.Close()

	err := store.UpdateRuleSet(&entity.RuleSet{Tag: "geoip-cn", Name: "Updated"})
	if err != nil {
		t.Fatalf("UpdateRuleSet failed: %v", err)
	}
}

func TestSupabaseStorage_DeleteRuleSet(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]supabaseRuleSet{{Tag: "geoip-cn"}})
	})
	defer server.Close()

	err := store.DeleteRuleSet("geoip-cn")
	if err != nil {
		t.Fatalf("DeleteRuleSet failed: %v", err)
	}
}

// --- GlobalSetting Tests ---

func TestSupabaseStorage_SetGlobalSetting(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuthHeaders(t, r)
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/global_settings") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		prefer := r.Header.Get("Prefer")
		if !strings.Contains(prefer, "resolution=merge-duplicates") {
			t.Errorf("expected merge-duplicates in Prefer, got %q", prefer)
		}
		var sg supabaseGlobalSetting
		json.NewDecoder(r.Body).Decode(&sg)
		if sg.Key != "theme" || sg.Value != "dark" {
			t.Errorf("unexpected body: %+v", sg)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode([]supabaseGlobalSetting{sg})
	})
	defer server.Close()

	err := store.SetGlobalSetting("theme", "dark")
	if err != nil {
		t.Fatalf("SetGlobalSetting failed: %v", err)
	}
}

func TestSupabaseStorage_SetGlobalSetting_Update(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		// Upsert 更新已有记录返回 200
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]supabaseGlobalSetting{{Key: "theme", Value: "light"}})
	})
	defer server.Close()

	err := store.SetGlobalSetting("theme", "light")
	if err != nil {
		t.Fatalf("SetGlobalSetting (update) failed: %v", err)
	}
}

func TestSupabaseStorage_GetGlobalSetting(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.RawQuery, "key=eq.theme") {
			t.Errorf("expected key filter, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(supabaseGlobalSetting{Value: "dark"})
	})
	defer server.Close()

	val, err := store.GetGlobalSetting("theme")
	if err != nil {
		t.Fatalf("GetGlobalSetting failed: %v", err)
	}
	if val != "dark" {
		t.Errorf("expected 'dark', got %q", val)
	}
}

func TestSupabaseStorage_GetGlobalSetting_NotFound(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotAcceptable)
	})
	defer server.Close()

	_, err := store.GetGlobalSetting("nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSupabaseStorage_ListGlobalSettings(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]supabaseGlobalSetting{
			{Key: "theme", Value: "dark"},
			{Key: "lang", Value: "zh"},
		})
	})
	defer server.Close()

	settings, err := store.ListGlobalSettings()
	if err != nil {
		t.Fatalf("ListGlobalSettings failed: %v", err)
	}
	if len(settings) != 2 {
		t.Fatalf("expected 2, got %d", len(settings))
	}
	if settings["theme"] != "dark" || settings["lang"] != "zh" {
		t.Errorf("unexpected settings: %+v", settings)
	}
}

func TestSupabaseStorage_ListGlobalSettings_Empty(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	})
	defer server.Close()

	settings, err := store.ListGlobalSettings()
	if err != nil {
		t.Fatalf("ListGlobalSettings failed: %v", err)
	}
	if len(settings) != 0 {
		t.Errorf("expected empty map, got %d entries", len(settings))
	}
}

func TestSupabaseStorage_DeleteGlobalSetting(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]supabaseGlobalSetting{{Key: "theme", Value: "dark"}})
	})
	defer server.Close()

	err := store.DeleteGlobalSetting("theme")
	if err != nil {
		t.Fatalf("DeleteGlobalSetting failed: %v", err)
	}
}

func TestSupabaseStorage_DeleteGlobalSetting_NotFound(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	})
	defer server.Close()

	err := store.DeleteGlobalSetting("nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSupabaseStorage_DeleteDevice_ClearsBindingsFirst(t *testing.T) {
	requests := make([]string, 0, 2)
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		switch len(requests) {
		case 1:
			if r.Method != http.MethodDelete || !strings.Contains(r.URL.RawQuery, "device_code=eq.phone") {
				t.Fatalf("unexpected first request: %s", requests[0])
			}
			w.Write([]byte("[]"))
		case 2:
			if r.Method != http.MethodDelete || !strings.Contains(r.URL.RawQuery, "code=eq.phone") {
				t.Fatalf("unexpected second request: %s", requests[1])
			}
			json.NewEncoder(w).Encode([]supabaseDevice{{Code: "phone"}})
		default:
			t.Fatalf("unexpected extra request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	if err := store.DeleteDevice("phone"); err != nil {
		t.Fatalf("DeleteDevice failed: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("unexpected request count: %v", requests)
	}
}

func TestSupabaseStorage_DeleteWireGuard_ClearsDevicesAndPeersFirst(t *testing.T) {
	requests := make([]string, 0, 3)
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		switch len(requests) {
		case 1:
			if r.Method != http.MethodPatch || !strings.Contains(r.URL.RawQuery, "wire_guard_tag=eq.wg-main") {
				t.Fatalf("unexpected first request: %s", requests[0])
			}
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode first request body failed: %v", err)
			}
			if body["wire_guard_tag"] != "" || body["wire_guard_client_addr"] != "" || body["wire_guard_client_key"] != "" {
				t.Fatalf("unexpected first request body: %+v", body)
			}
			w.Write([]byte("[]"))
		case 2:
			if r.Method != http.MethodDelete || !strings.Contains(r.URL.RawQuery, "wire_guard_tag=eq.wg-main") {
				t.Fatalf("unexpected second request: %s", requests[1])
			}
			w.Write([]byte("[]"))
		case 3:
			if r.Method != http.MethodDelete || !strings.Contains(r.URL.RawQuery, "tag=eq.wg-main") {
				t.Fatalf("unexpected third request: %s", requests[2])
			}
			json.NewEncoder(w).Encode([]supabaseWireGuard{{Tag: "wg-main"}})
		default:
			t.Fatalf("unexpected extra request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	if err := store.DeleteWireGuard("wg-main"); err != nil {
		t.Fatalf("DeleteWireGuard failed: %v", err)
	}
	if len(requests) != 3 {
		t.Fatalf("unexpected request count: %v", requests)
	}
}

func TestSupabaseStorage_CreateExtraOutbound_OmitsZeroID(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuthHeaders(t, r)
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/outbounds") {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body failed: %v", err)
		}
		// 新建手工 Outbound 时不应显式发送 id=0，否则 Supabase 会按主键 0 插入。
		if _, exists := body["id"]; exists {
			t.Fatalf("request body should omit zero id, got %+v", body)
		}
		if body["tag"] != "manual-a" || body["source"] != string(entity.OutboundSourceManual) {
			t.Fatalf("unexpected request body: %+v", body)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode([]map[string]any{{"tag": "manual-a"}})
	})
	defer server.Close()

	if err := store.CreateExtraOutbound(&entity.Outbound{
		Tag:     "manual-a",
		Name:    "manual-a",
		Type:    "vmess",
		Enabled: true,
	}); err != nil {
		t.Fatalf("CreateExtraOutbound failed: %v", err)
	}
}

func TestSupabaseStorage_CreateOrUpdateOutbounds_OmitsZeroID(t *testing.T) {
	server, store := mockPostgREST(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuthHeaders(t, r)
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/outbounds") {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "on_conflict=tag") {
			t.Fatalf("expected on_conflict=tag, got %s", r.URL.RawQuery)
		}

		var body []map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body failed: %v", err)
		}
		if len(body) != 2 {
			t.Fatalf("unexpected body length: %d", len(body))
		}
		for _, item := range body {
			// 批量写入订阅缓存时，新增记录也必须省略 id，让数据库分配自增值。
			if _, exists := item["id"]; exists {
				t.Fatalf("request item should omit zero id, got %+v", item)
			}
		}
		if body[0]["tag"] != "sub-node-a" || body[1]["tag"] != "sub-node-b" {
			t.Fatalf("unexpected request body: %+v", body)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(body)
	})
	defer server.Close()

	current := time.Now().UTC()
	if err := store.CreateOrUpdateOutbounds([]*entity.Outbound{
		{
			Tag:            "sub-node-a",
			Name:           "sub-node-a",
			Type:           "vmess",
			Enabled:        true,
			Source:         entity.OutboundSourceSubscription,
			SubscribeName:  "sub-a",
			LastFetchTime:  &current,
			ConfigJSON:     `{"tag":"sub-node-a","type":"vmess"}`,
			VisibleDevices: "phone",
		},
		{
			Tag:            "sub-node-b",
			Name:           "sub-node-b",
			Type:           "trojan",
			Enabled:        true,
			Source:         entity.OutboundSourceSubscription,
			SubscribeName:  "sub-a",
			LastFetchTime:  &current,
			ConfigJSON:     `{"tag":"sub-node-b","type":"trojan"}`,
			VisibleDevices: "",
		},
	}); err != nil {
		t.Fatalf("CreateOrUpdateOutbounds failed: %v", err)
	}
}

// --- NewSupabaseStorage Test ---

func TestNewSupabaseStorage(t *testing.T) {
	store := NewSupabaseStorage("https://example.supabase.co", "my-key")
	if store.baseURL != "https://example.supabase.co/rest/v1" {
		t.Errorf("unexpected baseURL: %s", store.baseURL)
	}
	if store.apiKey != "my-key" {
		t.Errorf("unexpected apiKey: %s", store.apiKey)
	}
	if store.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}
