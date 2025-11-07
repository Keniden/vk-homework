package main

// сюда писать код

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

var (
	// @BotFather в телеграме даст вам токен. Если захотите потыкать своего бота через телегу - используйте именно его
	BotToken = "XXX"

	// Урл, в который будет стучаться телега при получении сообщения от пользователя.
	// Может быть как айпишником личной виртуалки, так и просто выдан сервисом для деплоя
	WebhookURL = "https://525f2cb5.ngrok.io"
	store      = NewTaskStore()
)

type Task struct {
	ID               int
	Title            string
	OwnerID          int64
	OwnerUsername    string
	AssigneeID       int64
	AssigneeUsername string
}

type TaskStore struct {
	mu     sync.Mutex
	items  map[int]*Task
	nextID int
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		items:  make(map[int]*Task),
		nextID: 1,
	}
}

func (t *TaskStore) New(ownerID int64, ownerUsername, title string) *Task {
	t.mu.Lock()
	defer t.mu.Unlock()

	id := t.nextID
	t.nextID++

	task := &Task{
		ID:            id,
		Title:         title,
		OwnerID:       ownerID,
		OwnerUsername: ownerUsername,
	}

	t.items[id] = task

	return task
}

func (t *TaskStore) Get(id int) (*Task, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	task, ok := t.items[id]
	return task, ok
}

func (t *TaskStore) Assign(id int, userID int, username string) (*Task, int64, string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	task, ok := t.items[id]
	if !ok {
		return nil, 0, "", false
	}

	prevAssigneeID := task.AssigneeID
	prevAssigneeUsername := task.AssigneeUsername

	task.AssigneeID = int64(userID)
	task.AssigneeUsername = username
	return task, prevAssigneeID, prevAssigneeUsername, true
}

func (t *TaskStore) Unassign(id int, userID int) (*Task, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	task, ok := t.items[id]
	if !ok {
		return nil, false
	}
	if task.AssigneeID != int64(userID) {
		return nil, false
	}
	task.AssigneeID = 0
	task.AssigneeUsername = ""
	return task, true
}

func (t *TaskStore) Resolve(id int, userID int64) (*Task, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	task, ok := t.items[id]
	if !ok {
		return nil, false
	}
	if task.AssigneeID != userID {
		return nil, false
	}
	delete(t.items, id)
	return task, true
}

func (t *TaskStore) ListSorted() []int {
	t.mu.Lock()
	defer t.mu.Unlock()
	ids := make([]int, 0, len(t.items))
	for id := range t.items {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}

func formatTask(t *Task, me int64) string {
	line1 := fmt.Sprintf("%d. %s by @%s", t.ID, t.Title, t.OwnerUsername)

	if t.AssigneeID == 0 {
		return fmt.Sprintf("%s\n/assign_%d", line1, t.ID)
	}

	if t.AssigneeID == me {
		return fmt.Sprintf("%s\nassignee: я\n/unassign_%d /resolve_%d", line1, t.ID, t.ID)
	}

	return fmt.Sprintf("%s\nassignee: @%s", line1, t.AssigneeUsername)
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := update.Message
	user := msg.From
	text := strings.TrimSpace(msg.Text)

	switch {
	case strings.HasPrefix(text, "/new"):
		title := strings.TrimSpace(strings.TrimPrefix(text, "/new"))
		if title == "" {
			send(bot, user.ID, "Введите название задачи")
			return
		}
		task := store.New(user.ID, user.UserName, title)
		send(bot, user.ID, fmt.Sprintf(`Задача "%s" создана, id=%d`, title, task.ID))
		return

	case text == "/tasks":
		ids := store.ListSorted()
		if len(ids) == 0 {
			send(bot, user.ID, "Нет задач")
			return
		}
		var sb strings.Builder
		for i, id := range ids {
			t, _ := store.Get(id)
			if i > 0 {
				sb.WriteString("\n\n")
			}
			sb.WriteString(formatTask(t, user.ID))
		}
		send(bot, user.ID, sb.String())
		return

	case strings.HasPrefix(text, "/assign_"):
		idStr := strings.TrimPrefix(text, "/assign_")
		id, _ := strconv.Atoi(idStr)
		t, prevID, _, ok := store.Assign(id, int(user.ID), user.UserName)
		if !ok {
			send(bot, user.ID, "Задача не найдена")
			return
		}
		send(bot, user.ID, fmt.Sprintf(`Задача "%s" назначена на вас`, t.Title))
		if prevID != 0 && prevID != user.ID {
			send(bot, prevID, fmt.Sprintf(`Задача "%s" назначена на @%s`, t.Title, user.UserName))
		} else {
			if t.OwnerID != user.ID {
				send(bot, t.OwnerID, fmt.Sprintf(`Задача "%s" назначена на @%s`, t.Title, user.UserName))
			}
		}
		return

	case strings.HasPrefix(text, "/unassign_"):
		idStr := strings.TrimPrefix(text, "/unassign_")
		id, _ := strconv.Atoi(idStr)
		t, ok := store.Unassign(id, int(user.ID))
		if !ok {
			send(bot, user.ID, "Задача не на вас")
			return
		}
		send(bot, user.ID, "Принято")
		if t.OwnerID != user.ID {
			send(bot, t.OwnerID, fmt.Sprintf(`Задача "%s" осталась без исполнителя`, t.Title))
		}
		return

	case strings.HasPrefix(text, "/resolve_"):
		idStr := strings.TrimPrefix(text, "/resolve_")
		id, _ := strconv.Atoi(idStr)
		t, ok := store.Resolve(id, user.ID)
		if !ok {
			send(bot, user.ID, "Задача не на вас")
			return
		}
		send(bot, user.ID, fmt.Sprintf(`Задача "%s" выполнена`, t.Title))
		if t.OwnerID != user.ID {
			send(bot, t.OwnerID, fmt.Sprintf(`Задача "%s" выполнена @%s`, t.Title, user.UserName))
		}
		return

	case text == "/my":
		ids := store.ListSorted()
		var sb strings.Builder
		for _, id := range ids {
			t, _ := store.Get(id)
			if t.AssigneeID == user.ID {
				if sb.Len() > 0 {
					sb.WriteString("\n\n")
				}
				sb.WriteString(fmt.Sprintf("%d. %s by @%s\n/unassign_%d /resolve_%d",
					t.ID, t.Title, t.OwnerUsername, t.ID, t.ID))
			}
		}
		if sb.Len() == 0 {
			send(bot, user.ID, "Нет задач")
		} else {
			send(bot, user.ID, sb.String())
		}
		return

	case text == "/owner":
		ids := store.ListSorted()
		var sb strings.Builder
		for _, id := range ids {
			t, _ := store.Get(id)
			if t.OwnerID == user.ID {
				if sb.Len() > 0 {
					sb.WriteString("\n\n")
				}
				sb.WriteString(fmt.Sprintf("%d. %s by @%s\n/assign_%d",
					t.ID, t.Title, t.OwnerUsername, t.ID))
			}
		}
		if sb.Len() == 0 {
			send(bot, user.ID, "Нет задач")
		} else {
			send(bot, user.ID, sb.String())
		}
		return
	}
}

func startTaskBot(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		return err
	}

	wh, err := tgbotapi.NewWebhook(WebhookURL)
	if err != nil {
		return err
	}
	_, err = bot.Request(wh)
	if err != nil {
		return err
	}

	updates := bot.ListenForWebhook("/")

	srv := &http.Server{
		Addr: "127.0.0.1:8081",
	}

	go func() {
		_ = srv.ListenAndServe()
	}()

	go func() {
		for upd := range updates {
			handleCommand(bot, upd)
		}
	}()

	<-ctx.Done()
	_ = srv.Shutdown(context.Background())
	return nil
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
