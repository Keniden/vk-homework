package room

import "github.com/Keniden/vk-homework/game/item"

type Room struct {
	Name     string
	LookDesc string
	GoDesc   string
	Items    []*item.Item
	ToGO     []*Room
	Backpack bool
	IsHall   bool
	Door     bool
}

func NewRoom(Name string, LookDesc string, GoDesc string, Items []*item.Item) *Room {
	return &Room{
		Name:     Name,
		LookDesc: LookDesc,
		GoDesc:   GoDesc,
		Items:    make([]*item.Item, 0),
	}
}

func (r *Room) AddItem(item1 string) {
	newItem := &item.Item{Name: item1}
	r.Items = append(r.Items, newItem)
}

func (r *Room) AddRout(add *Room) {
	r.ToGO = append(r.ToGO, add)
}

func (r *Room) AddBackpack() {
	r.Backpack = true
}

func (r *Room) ItHall() {
	r.IsHall = true
}

func (r *Room) UnlockDoor() {
	if r.IsHall {
		r.Door = true
	}
}
