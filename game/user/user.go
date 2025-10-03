package user

import (
	"fmt"
	"strings"

	"gitlab.vk-golang.ru/vk-golang/lectures/01_intro/99_hw/game/room"
)

type User struct {
	InPlace *room.Room
	Items   []string
}

func NewUser(InPlace *room.Room, Items []string) *User {
	return &User{
		InPlace: InPlace,
		Items:   Items,
	}
}

func (u *User) Look() string {
	toGo := []string{}
	for _, r := range u.InPlace.ToGO {
		toGo = append(toGo, r.Name)
	}
	return fmt.Sprintf("%sна столе: %s, надо собрать рюкзак и идти в универ. можно пройти - %s", u.InPlace.LookDesc, u.InPlace.Items[0], strings.Join(toGo, ", "))
}

func (u *User) GoTo(place string) string {
	for _, p := range u.InPlace.ToGO {
		if p.Name == place {
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
