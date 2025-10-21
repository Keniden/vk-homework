package main

import (
    "fmt"
    "sort"
    "sync"
)

func RunPipeline(cmds ...cmd) {
    in := make(chan interface{})
    close(in)
    for _, c := range cmds {
        out := make(chan interface{})
        go func(in, out chan interface{}, c cmd) {
            defer close(out)
            c(in, out)
        }(in, out, c)
        in = out
    }
    for range in {
    }
}

func SelectUsers(in, out chan interface{}) {
    var wg sync.WaitGroup
    var seen sync.Map
    for v := range in {
        email := v.(string)
        wg.Add(1)
        go func(email string) {
            defer wg.Done()
            user := GetUser(email)
            key := user.Email
            if _, loaded := seen.LoadOrStore(key, struct{}{}); !loaded {
                out <- user
            }
        }(email)
    }
    wg.Wait()
}

func SelectMessages(in, out chan interface{}) {
    // 	in - User
    // 	out - MsgID
    var wg sync.WaitGroup
    buffer := make([]User, 0, GetMessagesMaxUsersBatch)
    for v := range in {
        buffer = append(buffer, v.(User))
        if len(buffer) == GetMessagesMaxUsersBatch {
            batch := make([]User, len(buffer))
            copy(batch, buffer)
            wg.Add(1)
            go func(users []User) {
                defer wg.Done()
                msgs, _ := GetMessages(users...)
                for _, msg := range msgs {
                    out <- msg
                }
            }(batch)
            buffer = buffer[:0]
        }
    }
    if len(buffer) > 0 {
        batch := make([]User, len(buffer))
        copy(batch, buffer)
        wg.Add(1)
        go func(users []User) {
            defer wg.Done()
            msgs, _ := GetMessages(users...)
            for _, msg := range msgs {
                out <- msg
            }
        }(batch)
    }
    wg.Wait()
}

func CheckSpam(in, out chan interface{}) {
    // in - MsgID
    // out - MsgData
    var wg sync.WaitGroup
    for i := 0; i < HasSpamMaxAsyncRequests; i++ {
        wg.Add(1)
        go WorkerCheckSpam(in, out, &wg)
    }
    wg.Wait()
}

func WorkerCheckSpam(in, out chan interface{}, wg *sync.WaitGroup) {
    defer wg.Done()
    for v := range in {
        msgID := v.(MsgID)
        hasSpam, _ := HasSpam(msgID)
        out <- MsgData{
            ID:      msgID,
            HasSpam: hasSpam,
        }
    }
}

func CombineResults(in, out chan interface{}) {
    // in - MsgData
    // out - string
    s := make([]MsgData, 0, 100)
    for v := range in {
        s = append(s, v.(MsgData))
    }

    sort.Slice(s, func(i, j int) bool {
        if s[i].HasSpam == s[j].HasSpam {
            return s[i].ID < s[j].ID
        }
        return s[i].HasSpam && !s[j].HasSpam
    })
    for _, msg := range s {
        str := fmt.Sprintf("%v %d", msg.HasSpam, msg.ID)
        out <- str
    }
}

