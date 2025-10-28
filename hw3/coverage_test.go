package main

import (
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "net/http/httptest"
    "net/url"
    "strconv"
    "strings"
    "testing"
    "time"
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

func startSearchHTTPServer() (*httptest.Server, string, func()) {
    ts := httptest.NewServer(http.HandlerFunc(SearchServer))
    return ts, ts.URL, ts.Close
}

func startStubServer(h http.HandlerFunc) (*httptest.Server, string, func()) {
    ts := httptest.NewServer(h)
    return ts, ts.URL, ts.Close
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

func TestMakeLess_EmptyField(t *testing.T) {
    if _, err := makeLess(""); err == nil {
        t.Fatalf("expected error for empty field")
    }
}

func TestSearchServer_Unauthorized(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/", nil)
    rr := httptest.NewRecorder()
    SearchServer(rr, req)
    if rr.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401, got %d", rr.Code)
    }
}

func TestFindUsers_SuccessNextPage(t *testing.T) {
    setUsers(t, []User{{ID: 1, Name: "bbb"}, {ID: 2, Name: "Aaa"}, {ID: 3, Name: "ccc"}, {ID: 4, Name: "ddd"}})
    _, urlStr, closeFn := startSearchHTTPServer()
    defer closeFn()

    cli := &SearchClient{URL: urlStr, AccessToken: "token"}
    req := SearchRequest{Limit: 2, Offset: 0, Query: "", OrderField: "Name", OrderBy: OrderByAsc}
    resp, err := cli.FindUsers(req)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp == nil {
        t.Fatalf("nil response")
    }
    if !resp.NextPage {
        t.Fatalf("expected NextPage=true")
    }
    if len(resp.Users) != 2 {
        t.Fatalf("expected 2 users, got %d", len(resp.Users))
    }
    if strings.ToLower(resp.Users[0].Name) > strings.ToLower(resp.Users[1].Name) {
        t.Fatalf("expected users sorted by Name asc")
    }
}

func TestFindUsers_BadOrderField(t *testing.T) {
    setUsers(t, []User{{ID: 1, Name: "x"}, {ID: 2, Name: "y"}})
    _, urlStr, closeFn := startSearchHTTPServer()
    defer closeFn()

    cli := &SearchClient{URL: urlStr, AccessToken: "token"}
    req := SearchRequest{Limit: 2, Offset: 0, OrderField: "UnknownField", OrderBy: OrderByAsc}
    _, err := cli.FindUsers(req)
    if err == nil {
        t.Fatalf("expected error, got nil")
    }
    want := "OrderFeld UnknownField invalid"
    if err.Error() != want {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestFindUsers_BadAccessToken(t *testing.T) {
    setUsers(t, []User{{ID: 1, Name: "a"}})
    _, urlStr, closeFn := startSearchHTTPServer()
    defer closeFn()

    cli := &SearchClient{URL: urlStr, AccessToken: ""}
    req := SearchRequest{Limit: 1}
    _, err := cli.FindUsers(req)
    if err == nil || err.Error() != "bad AccessToken" {
        t.Fatalf("expected bad AccessToken error, got %v", err)
    }
}

func TestFindUsers_Timeout(t *testing.T) {
    h := func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(1500 * time.Millisecond)
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("[]"))
    }
    _, urlStr, closeFn := startStubServer(h)
    defer closeFn()

    cli := &SearchClient{URL: urlStr, AccessToken: "token"}
    req := SearchRequest{Limit: 1}
    _, err := cli.FindUsers(req)
    if err == nil || !strings.HasPrefix(err.Error(), "timeout for ") {
        t.Fatalf("expected timeout error, got %v", err)
    }
}

func TestFindUsers_Server500(t *testing.T) {
    h := func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "boom", http.StatusInternalServerError)
    }
    _, urlStr, closeFn := startStubServer(h)
    defer closeFn()

    cli := &SearchClient{URL: urlStr, AccessToken: "token"}
    req := SearchRequest{Limit: 1}
    _, err := cli.FindUsers(req)
    if err == nil || err.Error() != "SearchServer fatal error" {
        t.Fatalf("expected fatal error, got %v", err)
    }
}

func TestFindUsers_BadRequestInvalidJSON(t *testing.T) {
    h := func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest)
        _, _ = w.Write([]byte("oops"))
    }
    _, urlStr, closeFn := startStubServer(h)
    defer closeFn()

    cli := &SearchClient{URL: urlStr, AccessToken: "token"}
    req := SearchRequest{Limit: 1}
    _, err := cli.FindUsers(req)
    if err == nil || !strings.HasPrefix(err.Error(), "cant unpack error json:") {
        t.Fatalf("expected json unpack error, got %v", err)
    }
}

func TestFindUsers_BadRequestUnknownError(t *testing.T) {
    h := func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest)
        _ = json.NewEncoder(w).Encode(map[string]string{"Error": "some other"})
    }
    _, urlStr, closeFn := startStubServer(h)
    defer closeFn()

    cli := &SearchClient{URL: urlStr, AccessToken: "token"}
    _, err := cli.FindUsers(SearchRequest{Limit: 1})
    if err == nil || !strings.HasPrefix(err.Error(), "unknown bad request error:") {
        t.Fatalf("expected unknown bad request error, got %v", err)
    }
}

func TestFindUsers_BadResultJSON(t *testing.T) {
    h := func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("not-json"))
    }
    _, urlStr, closeFn := startStubServer(h)
    defer closeFn()

    cli := &SearchClient{URL: urlStr, AccessToken: "token"}
    req := SearchRequest{Limit: 1}
    _, err := cli.FindUsers(req)
    if err == nil || !strings.HasPrefix(err.Error(), "cant unpack result json:") {
        t.Fatalf("expected result json unpack error, got %v", err)
    }
}

func TestFindUsers_ParamValidation(t *testing.T) {
    cli := &SearchClient{URL: "http://example.com", AccessToken: "token"}

    _, err := cli.FindUsers(SearchRequest{Limit: -1})
    if err == nil || err.Error() != "limit must be > 0" {
        t.Fatalf("expected limit validation error, got %v", err)
    }

    _, err = cli.FindUsers(SearchRequest{Limit: 1, Offset: -2})
    if err == nil || err.Error() != "offset must be > 0" {
        t.Fatalf("expected offset validation error, got %v", err)
    }
}

func TestFindUsers_NextPageFalse_WhenServerReturnsLessThanLimit(t *testing.T) {
    h := func(w http.ResponseWriter, r *http.Request) {
        q := r.URL.Query()
        limStr := q.Get("limit")
        lim, _ := strconv.Atoi(limStr)
        if lim < 0 {
            lim = 0
        }
        n := 0
        if lim > 0 {
            n = lim - 1
        }
        type outUser struct {
            ID     int
            Name   string
            Age    int
            About  string
            Gender string
        }
        arr := make([]outUser, 0, n)
        for i := 0; i < n; i++ {
            arr = append(arr, outUser{ID: i + 1, Name: "User" + strconv.Itoa(i+1)})
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(arr)
    }
    _, urlStr, closeFn := startStubServer(h)
    defer closeFn()

    cli := &SearchClient{URL: urlStr, AccessToken: "token"}
    req := SearchRequest{Limit: 3}
    resp, err := cli.FindUsers(req)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.NextPage {
        t.Fatalf("expected NextPage=false, got true")
    }
}

func TestFindUsers_LimitCapAndIncrement(t *testing.T) {
    h := func(w http.ResponseWriter, r *http.Request) {
        u, _ := url.Parse(r.URL.String())
        limStr := u.Query().Get("limit")
        if limStr != "26" {
            t.Fatalf("expected transmitted limit=26, got %s", limStr)
        }
        w.Header().Set("Content-Type", "application/json")
        _, _ = w.Write([]byte("[]"))
    }
    _, urlStr, closeFn := startStubServer(h)
    defer closeFn()

    cli := &SearchClient{URL: urlStr, AccessToken: "token"}
    _, err := cli.FindUsers(SearchRequest{Limit: 100})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestFindUsers_UnknownError(t *testing.T) {
    cli := &SearchClient{URL: "http://127.0.0.1:1", AccessToken: "token"}
    _, err := cli.FindUsers(SearchRequest{Limit: 1})
    if err == nil || !strings.HasPrefix(err.Error(), "unknown error ") {
        t.Fatalf("expected unknown error prefix, got %v", err)
    }
}

