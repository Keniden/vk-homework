package main

import (
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

func setUsers(t *testing.T, us []User) {
    t.Helper()
    _, _ = loadUsers() 
    usersCache = us
    usersLoadErr = nil
}

func doReq(t *testing.T, method, target string) *httptest.ResponseRecorder {
    t.Helper()
    req := httptest.NewRequest(method, target, nil)
    req.Header.Set("AccessToken", "token")
    rr := httptest.NewRecorder()
    SearchServer(rr, req)
    return rr
}

func TestSearchServer_MethodNotAllowed(t *testing.T) {
    rr := doReq(t, http.MethodPost, "/?limit=1")
    if rr.Code != http.StatusMethodNotAllowed {
        t.Fatalf("expected 405, got %d", rr.Code)
    }
}

func TestSearchServer_InvalidLimitAndOffset(t *testing.T) {
    rr := doReq(t, http.MethodGet, "/?limit=abc")
    if rr.Code != http.StatusBadRequest {
        t.Fatalf("expected 400 for limit, got %d", rr.Code)
    }
    var m map[string]string
    _ = json.Unmarshal(rr.Body.Bytes(), &m)
    if m["Error"] != "limit must be > 0" {
        t.Fatalf("unexpected error message for limit: %v", m)
    }

    rr = doReq(t, http.MethodGet, "/?offset=zzz")
    if rr.Code != http.StatusBadRequest {
        t.Fatalf("expected 400 for offset, got %d", rr.Code)
    }
    m = map[string]string{}
    _ = json.Unmarshal(rr.Body.Bytes(), &m)
    if m["Error"] != "offset must be > 0" {
        t.Fatalf("unexpected error message for offset: %v", m)
    }
}

func TestSearchServer_DefaultOrderByNameAsc(t *testing.T) {
    setUsers(t, []User{{ID: 1, Name: "bbb"}, {ID: 2, Name: "Aaa"}, {ID: 3, Name: "ccc"}})
    rr := doReq(t, http.MethodGet, "/?limit=10&offset=0&order_by=1")
    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
    var list []User
    _ = json.Unmarshal(rr.Body.Bytes(), &list)
    if len(list) != 3 {
        t.Fatalf("expected 3 users, got %d", len(list))
    }
    if !(strings.EqualFold(list[0].Name, "Aaa") && strings.EqualFold(list[1].Name, "bbb")) {
        t.Fatalf("unexpected order: %#v", list)
    }
}

func TestSearchServer_SortDescByAge(t *testing.T) {
    setUsers(t, []User{{ID: 1, Name: "a", Age: 10}, {ID: 2, Name: "b", Age: 30}, {ID: 3, Name: "c", Age: 20}})
    rr := doReq(t, http.MethodGet, "/?limit=10&offset=0&order_field=Age&order_by=-1")
    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
    var list []User
    _ = json.Unmarshal(rr.Body.Bytes(), &list)
    if !(list[0].Age == 30 && list[1].Age == 20 && list[2].Age == 10) {
        t.Fatalf("unexpected age order: %#v", list)
    }
}

func TestSearchServer_AsIsOrderAndPagination(t *testing.T) {
    setUsers(t, []User{{ID: 1, Name: "c"}, {ID: 2, Name: "a"}, {ID: 3, Name: "b"}})
    rr := doReq(t, http.MethodGet, "/?limit=1&offset=1&order_by=0")
    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
    var list []User
    _ = json.Unmarshal(rr.Body.Bytes(), &list)
    if len(list) != 1 || list[0].Name != "a" {
        t.Fatalf("unexpected page slice: %#v", list)
    }
}

func TestSearchServer_FilteringByNameAndAbout(t *testing.T) {
    setUsers(t, []User{{ID: 1, Name: "John Smith", About: "alpha"}, {ID: 2, Name: "Jane", About: "bravo"}})

    rr := doReq(t, http.MethodGet, "/?query=smith") 
    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
    var list []User
    _ = json.Unmarshal(rr.Body.Bytes(), &list)
    if len(list) != 1 || list[0].ID != 1 {
        t.Fatalf("unexpected filter by name: %#v", list)
    }

    rr = doReq(t, http.MethodGet, "/?query=bra")
    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
    list = nil
    _ = json.Unmarshal(rr.Body.Bytes(), &list)
    if len(list) != 1 || list[0].ID != 2 {
        t.Fatalf("unexpected filter by about: %#v", list)
    }
}

func TestSearchServer_OffsetBeyondLength(t *testing.T) {
    setUsers(t, []User{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}})
    rr := doReq(t, http.MethodGet, "/?offset=5")
    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
    body, _ := io.ReadAll(rr.Body)
    if strings.TrimSpace(string(body)) != "[]" {
        t.Fatalf("expected empty json array, got %s", string(body))
    }
}

func TestSearchServer_InvalidOrderField_MessageShape(t *testing.T) {
    setUsers(t, []User{{ID: 1, Name: "a"}})
    rr := doReq(t, http.MethodGet, "/?order_field=foo&order_by=1")
    if rr.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", rr.Code)
    }
    var m map[string]string
    _ = json.Unmarshal(rr.Body.Bytes(), &m)
    if m["Error"] != "OrderField invalid" {
        t.Fatalf("unexpected error json: %v", m)
    }
}

func TestSearchServer_InternalErrorPath(t *testing.T) {
    _, _ = loadUsers()
    usersLoadErr = errors.New("boom")
    rr := doReq(t, http.MethodGet, "/")
    if rr.Code != http.StatusInternalServerError {
        t.Fatalf("expected 500, got %d", rr.Code)
    }
    usersLoadErr = nil
}

func TestAtoiDefault(t *testing.T) {
    if v, err := atoiDefault("", 42); err != nil || v != 42 {
        t.Fatalf("expected default 42, got %d, err=%v", v, err)
    }
    if v, err := atoiDefault("10", 0); err != nil || v != 10 {
        t.Fatalf("expected 10, got %d, err=%v", v, err)
    }
    if _, err := atoiDefault("x", 0); err == nil {
        t.Fatalf("expected error for non-int")
    }
}

func TestMakeLess(t *testing.T) {
    nameLess, err := makeLess("Name")
    if err != nil || !nameLess(User{Name: "a"}, User{Name: "b"}) {
        t.Fatalf("name less failed")
    }
    ageLess, err := makeLess("Age")
    if err != nil || !ageLess(User{Age: 1}, User{Age: 2}) {
        t.Fatalf("age less failed")
    }
    idLess, err := makeLess("Id")
    if err != nil || !idLess(User{ID: 1}, User{ID: 2}) {
        t.Fatalf("id less failed")
    }
    if _, err := makeLess("Nope"); err == nil {
        t.Fatalf("expected invalid field error")
    }
}

