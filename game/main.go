package main

import (
	"fmt"
	"strings"

	"gitlab.vk-golang.ru/vk-golang/lectures/01_intro/99_hw/game/room"
	"gitlab.vk-golang.ru/vk-golang/lectures/01_intro/99_hw/game/user"
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
	street := room.NewRoom("улица", "", "", []string{""})
	kitchen := room.NewRoom("кухня", "ты находишься на кухне, ", "", []string{"чай"})
	myRoom := room.NewRoom("комната", "", "ты в своей комнате. ", []string{"чай"})
	hall := room.NewRoom("коридор", "", "ничего интересного. ", []string{})

	street.AddRout(hall)
	kitchen.AddRout(hall)
	myRoom.AddRout(hall)
	hall.AddRout(kitchen)
	hall.AddRout(myRoom)
	hall.AddRout(street)

	gamer = user.NewUser(kitchen, []string{})

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
	fmt.Println(handleCommand("осмотреться"))
	fmt.Println(handleCommand("идти коридор"))
	fmt.Println(handleCommand("идти комната"))
	fmt.Println(handleCommand("осмотреться"))
	fmt.Println(handleCommand("осмотреться"))
}
