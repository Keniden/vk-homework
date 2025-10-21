package user

import (
	"fmt"
	"strings"

	"github.com/Keniden/vk-homework/game/item"
	"github.com/Keniden/vk-homework/game/room"
)

type User struct {
	InPlace  *room.Room
	Items    []*item.Item
	Backpack bool
}

func NewUser(InPlace *room.Room) *User {
	return &User{
		InPlace:  InPlace,
		Items:    make([]*item.Item, 0),
		Backpack: false,
	}
}

func (u *User) Look() string {
	r := u.InPlace
	toGo := make([]string, 0, len(r.ToGO))
	for _, next := range r.ToGO {
		if next != nil {
			toGo = append(toGo, next.Name)
		}
	}
	items := make([]string, 0, len(r.Items))
	for _, it := range r.Items {
		if it != nil && it.Name != "" {
			items = append(items, it.Name)
		}
	}

	var mainPart string
	if r.Name == "комната" && len(items) == 0 && !r.Backpack {
		mainPart = "пустая комната."
	} else {
		tablePart := "на столе: "
		if len(items) == 0 {
			tablePart += "ничего"
		} else {
			tablePart += strings.Join(items, ", ")
		}

		if r.Backpack {
			tablePart += ", на стуле: рюкзак"
		}
		mainPart = r.LookDesc + tablePart

		mission := r.MissionText
		if r.Name == "кухня" {
			if u.Backpack {
				mission = "надо идти в универ."
			} else {
				mission = "надо собрать рюкзак и идти в универ."
			}
		}

		if mission != "" {
			if !strings.HasSuffix(mainPart, ".") {
				mainPart += ", " + mission
			} else {
				mainPart += " " + mission
			}
		}

		if !strings.HasSuffix(mainPart, ".") {
			mainPart += "."
		}
	}

	exits := "можно пройти - "
	if len(toGo) == 0 {
		exits += "некуда"
	} else {
		exits += strings.Join(toGo, ", ")
	}

	return mainPart + " " + exits
}

func (u *User) GoTo(place string) string {
	for _, p := range u.InPlace.ToGO {
		if p.Name == place {

			if place == "улица" && !u.InPlace.Door {
				return "дверь закрыта"
			}

			u.InPlace = p

			toGo := []string{}
			for _, r := range u.InPlace.ToGO {
				toGo = append(toGo, r.Name)
			}
			if place == "улица" && u.InPlace.Name == "улица" {
				return "на улице весна. можно пройти - домой"
			}
			return fmt.Sprintf("%sможно пройти - %s", p.GoDesc, strings.Join(toGo, ", "))
		}
	}
	return fmt.Sprintf("нет пути в %s", place)
}

func (u *User) PutOnBackpack() string {
	if u.InPlace.Backpack {
		u.InPlace.Backpack = false
		u.Backpack = true
		return "вы надели: рюкзак"
	}
	return "нет рюкзака"
}

func (u *User) AddInInventory(item *item.Item) {
	if u.Backpack {
		u.Items = append(u.Items, item)
	}
}

func (u *User) Take(item string) string {
	if !u.Backpack {
		return "некуда класть"
	}
	for idx, i := range u.InPlace.Items {
		if i.Name == item {
			u.AddInInventory(i)
			u.InPlace.Items = append(u.InPlace.Items[:idx], u.InPlace.Items[idx+1:]...)
			return fmt.Sprintf("предмет добавлен в инвентарь: %s", item)
		}
	}

	return "нет такого"
}

func (u *User) Use(item1, item2 string) string {
	hasItem := false
	for _, i := range u.Items {
		if i.Name == item1 {
			hasItem = true
			break
		}
	}
	if !hasItem {
		return fmt.Sprintf("нет предмета в инвентаре - %s", item1)
	}

	if item1 == "ключи" && item2 == "дверь" {
		u.InPlace.UnlockDoor()
		return "дверь открыта"
	}
	return "не к чему применить"

}
