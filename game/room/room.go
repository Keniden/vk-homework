package room

type Room struct {
	Name     string
	LookDesc string
	GoDesc   string
	Items    []string
	ToGO     []*Room
}
func NewRoom(Name string, LookDesc string, GoDesc string, Items []string) *Room {
    return &Room{
        Name:     Name,
        LookDesc: LookDesc,
        GoDesc:   GoDesc,
        Items:    Items,
    }
}



func (r *Room) AddRout(add *Room) {
	r.ToGO = append(r.ToGO, add)
}
