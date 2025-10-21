package main

import (
	"fmt"
	"strings"

	"github.com/Keniden/vk-homework/game/item"
	"github.com/Keniden/vk-homework/game/room"
	"github.com/Keniden/vk-homework/game/user"
)

/*
	код писать в этом файле
	наверняка у вас будут какие-то структуры с методами, глобальные переменные ( тут можно ), функции
*/

var gamer *user.User

func initGame() {
	/*
		эта функция инициализирует игровой мир - все комнаты
		если что-то было - оно корректно перезатирается
	*/
	street := room.NewRoom("улица", "", "на улице весна. ", "", []*item.Item{})
	kitchen := room.NewRoom("кухня", "ты находишься на кухне, ", "кухня, ничего интересного. ", "надо собрать рюкзак и идти в универ.", []*item.Item{})
	myRoom := room.NewRoom("комната", "", "ты в своей комнате. ", "", []*item.Item{})
	hall := room.NewRoom("коридор", "", "ничего интересного. ", "", []*item.Item{})

	myRoom.AddBackpack()

	hall.ItHall()

	street.AddRout(hall)
	kitchen.AddRout(hall)
	myRoom.AddRout(hall)
	hall.AddRout(kitchen)
	hall.AddRout(myRoom)
	hall.AddRout(street)

	kitchen.AddItem("чай")

	myRoom.AddItem("ключи")
	myRoom.AddItem("конспекты")

	gamer = user.NewUser(kitchen)

}

func handleCommand(command string) string {
	/*
		данная функция принимает команду от "пользователя"
		и наверняка вызывает какой-то другой метод или функцию у "мира" - списка комнат
	*/
	cmd := strings.Split(command, " ")

	switch cmd[0] {
	case "осмотреться":
		return gamer.Look()
	case "идти":
		return gamer.GoTo(cmd[1])

	case "надеть":
		return gamer.PutOnBackpack()
	case "взять":
		return gamer.Take(cmd[1])

	case "применить":
		return gamer.Use(cmd[1], cmd[2])

	default:
		return "неизвестная команда"
	}

}

func main() {
	/*
		в этой функции можно ничего не писать,
		но тогда у вас не будет работать через go run main.go
		очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/
	initGame()
	fmt.Println(handleCommand("идти улица"))
}
