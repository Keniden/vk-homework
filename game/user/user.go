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
	backpack := ""
	if u.InPlace.Backpack{
		backpack = ", на стуле: рюкзак"
	}
	toGo := []string{}
	for _, r := range u.InPlace.ToGO {
		toGo = append(toGo, r.Name)
	}
	items := []string{}
	for _, i := range u.InPlace.Items{
		items = append(items, i.Name)
	}

	return fmt.Sprintf("%sна столе: %s, надо собрать рюкзак и идти в универ. можно пройти - %s", u.InPlace.LookDesc, strings.Join(items, ", "), strings.Join(toGo, ", "))
}

func (u *User) GoTo(place string) string {
	for _, p := range u.InPlace.ToGO {
		if p.Name == place {
			
			if place == "улица" && !u.InPlace.Door{
				return "дверь закрыта"
			}
			
			u.InPlace = p

			toGo := []string{}
			for _, r := range u.InPlace.ToGO {
				toGo = append(toGo, r.Name)
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
		}
	}

	return fmt.Sprintf("предмет добавлен в инвентарь: %s", item)
}

func (u *User) Use(item1, item2 string) string {
	if !u.Backpack {
		return "нет рюкзака"
	}
	isItem1 := false

	for _, i := range u.Items {
		if i.Name == item1 {
			isItem1 = true
		} else {
			return fmt.Sprintf("нет предмета в инвентаре - %s", item1)
		}
	}

	if isItem1 && (item1 == "ключи" && item2 == "дверь") {
		u.InPlace.UnlockDoor()
		return "дверь открыта"
	} else {
		return "не к чему применить"
	}

}
