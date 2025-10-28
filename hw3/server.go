package main

import (
    "encoding/json"
    "encoding/xml"
    "errors"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "sync"
)

type xmlUser struct {
    ID        int    `xml:"id"`
    GUID      string `xml:"guid"`
    IsActive  string `xml:"isActive"`
    Balance   string `xml:"balance"`
    Picture   string `xml:"picture"`
    Age       int    `xml:"age"`
    EyeColor  string `xml:"eyeColor"`
    FirstName string `xml:"first_name"`
    LastName  string `xml:"last_name"`
    Gender    string `xml:"gender"`
    Company   string `xml:"company"`
    Email     string `xml:"email"`
    Phone     string `xml:"phone"`
    Address   string `xml:"address"`
    About     string `xml:"about"`
}

type xmlRoot struct {
    Rows []xmlUser `xml:"row"`
}

var (
    usersOnce    sync.Once
    usersCache   []User
    usersLoadErr error
)

func datasetPaths() []string {
    return []string{
        "dataset.xml",
        filepath.Join("hw3", "dataset.xml"),
    }
}

func loadUsers() ([]User, error) {
    usersOnce.Do(func() {
        var data []byte
        var readErr error
        for _, p := range datasetPaths() {
            data, readErr = os.ReadFile(p)
            if readErr == nil {
                break
            }
        }
        if readErr != nil {
            usersLoadErr = fmt.Errorf("cannot read dataset.xml: %w", readErr)
            return
        }

        var root xmlRoot
        if err := xml.Unmarshal(data, &root); err != nil {
            usersLoadErr = fmt.Errorf("cannot unmarshal dataset.xml: %w", err)
            return
        }

        out := make([]User, 0, len(root.Rows))
        for _, ru := range root.Rows {
            out = append(out, User{
                ID:     ru.ID,
                Name:   strings.TrimSpace(ru.FirstName + " " + ru.LastName),
                Age:    ru.Age,
                About:  ru.About,
                Gender: ru.Gender,
            })
        }
        usersCache = out
    })
    return usersCache, usersLoadErr
}

const (
    orderByAsc  = 1
    orderByAsIs = 0
    orderByDesc = -1
)

const errorBadOrderField = "OrderField invalid"

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
    writeJSON(w, http.StatusBadRequest, map[string]string{"Error": msg})
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    if strings.TrimSpace(r.Header.Get("AccessToken")) == "" {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    users, err := loadUsers()
    if err != nil {
        http.Error(w, "SearchServer fatal error", http.StatusInternalServerError)
        return
    }

    q := r.URL.Query()
    limit, err := atoiDefault(q.Get("limit"), 0)
    if err != nil || limit < 0 {
        badRequest(w, "limit must be > 0")
        return
    }
    offset, err := atoiDefault(q.Get("offset"), 0)
    if err != nil || offset < 0 {
        badRequest(w, "offset must be > 0")
        return
    }
    query := strings.TrimSpace(q.Get("query"))
    orderField := strings.TrimSpace(q.Get("order_field"))
    orderBy, _ := atoiDefault(q.Get("order_by"), orderByAsIs)

    if orderField == "" {
        orderField = "Name"
    }

    filtered := filterUsers(users, query)

    if orderBy != orderByAsIs {
        lessFunc, err := makeLess(orderField)
        if err != nil {
            badRequest(w, errorBadOrderField)
            return
        }
        sort.SliceStable(filtered, func(i, j int) bool {
            if orderBy == orderByAsc {
                return lessFunc(filtered[i], filtered[j])
            }
            return lessFunc(filtered[j], filtered[i])
        })
    }

    if offset > len(filtered) {
        filtered = []User{}
    } else {
        filtered = filtered[offset:]
    }
    if limit > 0 && limit < len(filtered) {
        filtered = filtered[:limit]
    }

    writeJSON(w, http.StatusOK, filtered)
}

func atoiDefault(s string, def int) (int, error) {
    if strings.TrimSpace(s) == "" {
        return def, nil
    }
    n, err := strconv.Atoi(s)
    if err != nil {
        return 0, err
    }
    return n, nil
}

func filterUsers(users []User, query string) []User {
    if query == "" {
        out := make([]User, len(users))
        copy(out, users)
        return out
    }
    q := strings.ToLower(query)
    out := make([]User, 0, len(users))
    for _, u := range users {
        if strings.Contains(strings.ToLower(u.Name), q) || strings.Contains(strings.ToLower(u.About), q) {
            out = append(out, u)
        }
    }
    return out
}

func makeLess(field string) (func(a, b User) bool, error) {
    switch field {
    case "Id", "ID", "id":
        return func(a, b User) bool { return a.ID < b.ID }, nil
    case "Age", "age":
        return func(a, b User) bool { return a.Age < b.Age }, nil
    case "Name", "name":
        return func(a, b User) bool { return strings.ToLower(a.Name) < strings.ToLower(b.Name) }, nil
    case "":
        return nil, errors.New("empty field")
    default:
        return nil, errors.New("invalid field")
    }
}

func main() {
    http.HandleFunc("/", SearchServer)
    _ = http.ListenAndServe(":8080", nil)
}
