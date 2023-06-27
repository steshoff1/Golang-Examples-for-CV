package main

import (
	"strings"
)

var player = Player{"ты", Room{}, nil, nil, nil, false}

/*
код писать в этом файле
наверняка у вас будут какие-то структуры с методами, глобальные перменные ( тут можно ), функции
*/
type Nameable interface {
	getName() string
}

func nameDifGender(n Nameable) string {
	switch n.getName() {
	case "стол":
		return "столе"
	case "стул":
		return "стуле"
	}
	return n.getName()
}

type Player struct {
	Name       string
	Place      Room
	Tasks      []string
	Inventory  []*Thing
	FitOn      []*Thing
	ableToTake bool
}

func (p *Player) addTask(task string) {
	p.Tasks = append(p.Tasks, task)
}

func (p *Player) setPlace(place Room) {
	p.Place = place
}

func (p *Player) playerMove(name string) string {
	idx := 0
	flag := false
	res := ""
	i := 0
	for i = range p.Place.Doors {
		if p.Place.Doors[i].Name == name {
			break
		}
	}
	if p.Place.Doors[i].Status {
		for idx < len(p.Place.NextRooms) {
			if p.Place.NextRooms[idx].getName() == name {
				flag = true
				break
			}
			idx++
		}
	} else {
		return "дверь закрыта"
	}
	if flag {
		p.Place = *p.Place.NextRooms[idx]
	} else {
		return "нет пути в " + name
	}
	res = p.Place.InfoMoved + ". " + p.ableToGo()
	return res
}

func (p *Player) ableToGo() string {
	res := "можно пройти - "
	for i := 0; i < len(p.Place.NextRooms); i++ {
		res += player.Place.NextRooms[i].getName() + ", "
	}
	res = res[:len(res)-2]
	return res
}

func (r *Room) getInfo() string {
	res := r.Info
	flag := false
	if res != "" {
		res += " "
	}
	if len(r.FurnitureInIt) != 0 {
		for i := range r.FurnitureInIt {
			if len(r.FurnitureInIt[i].ThingsOnIt) != 0 {
				res += "на " + nameDifGender(r.FurnitureInIt[i]) + ": "
				if len(r.FurnitureInIt[i].ThingsOnIt) != 0 {
					flag = true
					for j := range r.FurnitureInIt[i].ThingsOnIt {
						res += nameDifGender(r.FurnitureInIt[i].ThingsOnIt[j])
						if j != len(r.FurnitureInIt[i].ThingsOnIt)-1 {
							res += ", "
						} else if i != len(r.FurnitureInIt) {
							res += ", "
						}
					}
				}
			} else {
				continue
			}

		}
	}
	if !flag {
		res = "пустая комната  "
	}
	if r.Name == "кухня" {
		res += player.Tasks[0]
	}
	res = res[:len(res)-2] + ". "
	res += player.ableToGo()
	return res
}

// ты находишься на кухне, на столе: чай, надо собрать рюкзак и идти в универ. можно пройти - коридор
// на столе: ключи, конспекты, на стуле: рюкзак. можно пройти - коридор
func (p *Player) lookUp() string {
	res := ""
	res = p.Place.getInfo()

	return res
}

func (p *Player) fitOnYourself(item string) string {
	flag := false
	res := "вы надели: "
	for i := 0; i < len(p.Place.FurnitureInIt); i++ {
		for j := 0; j < len(p.Place.FurnitureInIt[i].ThingsOnIt); j++ {
			if p.Place.FurnitureInIt[i].ThingsOnIt[j].getName() == item {
				p.ableToTake = true
				flag = true
				p.FitOn = append(p.FitOn, p.Place.FurnitureInIt[i].ThingsOnIt[j])
				if len(p.Place.FurnitureInIt[i].ThingsOnIt) == 1 {
					p.Place.FurnitureInIt[i].ThingsOnIt = nil
				} else {
					p.Place.FurnitureInIt[i].ThingsOnIt[j] = p.Place.FurnitureInIt[i].ThingsOnIt[len(p.Place.FurnitureInIt[i].ThingsOnIt)-1]
					tmp := Thing{"", nil}
					p.Place.FurnitureInIt[i].ThingsOnIt[len(p.Place.FurnitureInIt[i].ThingsOnIt)-1] = &tmp
					p.Place.FurnitureInIt[i].ThingsOnIt = p.Place.FurnitureInIt[i].ThingsOnIt[:len(p.Place.FurnitureInIt[i].ThingsOnIt)-1]
				}
			}
		}
	}
	if !flag {
		return "нет такого"
	}
	res += p.FitOn[len(p.FitOn)-1].getName()

	return res

}

func (p *Player) putInInventory(item string) string {
	res := ""
	flag := false
	if p.ableToTake {
		for i := 0; i < len(p.Place.FurnitureInIt); i++ {
			for j := 0; j < len(p.Place.FurnitureInIt[i].ThingsOnIt); j++ {
				if p.Place.FurnitureInIt[i].ThingsOnIt[j].getName() == item {
					flag = true
					p.Inventory = append(p.Inventory, p.Place.FurnitureInIt[i].ThingsOnIt[j])
					if len(p.Place.FurnitureInIt[i].ThingsOnIt) == 1 {
						p.Place.FurnitureInIt[i].ThingsOnIt = nil
					} else {
						p.Place.FurnitureInIt[i].ThingsOnIt[j] = p.Place.FurnitureInIt[i].ThingsOnIt[len(p.Place.FurnitureInIt[i].ThingsOnIt)-1]
						tmp := Thing{"", nil}
						p.Place.FurnitureInIt[i].ThingsOnIt[len(p.Place.FurnitureInIt[i].ThingsOnIt)-1] = &tmp
						p.Place.FurnitureInIt[i].ThingsOnIt = p.Place.FurnitureInIt[i].ThingsOnIt[:len(p.Place.FurnitureInIt[i].ThingsOnIt)-1]
					}
				}
			}
		}
		if !flag {
			return "нет такого"
		} else {
			res = "предмет добавлен в инвентарь: " + p.Inventory[len(p.Inventory)-1].getName()
			p.Tasks[0] = "надо идти в универ. "
		}

	} else {
		return "некуда класть"
	}
	return res
}

func (p *Player) useThing(thing string, item string) string {
	res := ""
	flag := false
	able := false
	for i := range p.Inventory {
		if p.Inventory[i].getName() == thing {
			able = true
		}
	}
	if able {
		for j := range p.Place.Doors {
			if p.Place.Doors[j].Name == item {
				p.Place.Doors[j].Status = true
				flag = true
				res = "дверь открыта"
			}
		}
	} else {
		return "нет предмета в инвентаре - " + thing
	}

	if !flag {
		res = "не к чему применить"
	}
	return res
}

type Door struct {
	Name        string
	Status      bool
	RoomConnect [2]*Room
}

type Room struct {
	Name          string
	NextRooms     []*Room
	FurnitureInIt []Furniture
	Doors         []*Door
	Info          string
	InfoMoved     string
}

func (r Room) getName() string {
	return r.Name
}

func (r *Room) addRoom(room *Room) {
	r.NextRooms = append(r.NextRooms, room)
}

func (r *Room) addFurniture(furniture ...Furniture) {
	r.FurnitureInIt = append(r.FurnitureInIt, furniture...)

}

func roomConnect(door *Door, room1 *Room, room2 ...*Room) {
	for _, val := range room2 {
		room1.addRoom(val)
		val.addRoom(room1)
		room1.Doors = append(room1.Doors, door)
		val.Doors = append(val.Doors, door)
	}
}

func roomConnectSet(door *Door, room1 *Room, room2 *Room) {
	tmp := []*Room{room2}
	room1.NextRooms = tmp
	room1.Doors = append(room1.Doors, door)
	room2.Doors = append(room2.Doors, door)
}

type Furniture struct {
	Name       string
	ThingsOnIt []*Thing
}

func (f Furniture) getName() string {
	return f.Name
}

func (f *Furniture) addThing(thing ...*Thing) {
	f.ThingsOnIt = append(f.ThingsOnIt, thing...)
}

type Thing struct {
	Name     string
	ActionTo []Furniture
}

func (t Thing) getName() string {
	return t.Name
}

func main() {
	/*
		в этой функции можно ничего не писать
		но тогда у вас не будет работать через go run main.go
		очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/
}

func initGame() {
	/*
		эта функция инициализирует игровой мир - все команты
		если что-то было - оно корректно перезатирается
	*/
	player = Player{"ты", Room{}, nil, nil, nil, false}
	var hall = Room{"коридор", nil, nil, nil, "ничего интересного", "ничего интересного"}
	var room = Room{"комната", nil, nil, nil, "", "ты в своей комнате"}
	var outside = Room{"улица", nil, nil, nil, "", "на улице весна"}
	var inside = Room{"домой", nil, nil, nil, "", ""}
	var kitchen = Room{"кухня", nil, nil, nil, "ты находишься на кухне,", "кухня, ничего интересного"}
	var doorKitchenHall = Door{"кухня", true, [2]*Room{&hall, &kitchen}}
	var doorRoomHall = Door{"комната", true, [2]*Room{&hall, &room}}
	var doorOutside = Door{"дверь", false, [2]*Room{&outside, &inside}}

	roomConnect(&doorKitchenHall, &hall, &kitchen)
	roomConnect(&doorRoomHall, &hall, &room)
	roomConnect(&doorOutside, &hall, &outside)
	roomConnectSet(&doorOutside, &outside, &inside)

	var tableKithcen = Furniture{"стол", nil}
	tea := Thing{"чай", []Furniture{tableKithcen}}
	tableKithcen.addThing(&tea)
	kitchen.addFurniture(tableKithcen)
	tableRoom := Furniture{"стол", nil}
	keys := Thing{"ключи", []Furniture{tableRoom}}
	conspects := Thing{"конспекты", []Furniture{tableRoom}}
	tableRoom.addThing(&keys, &conspects)
	backpack := Thing{"рюкзак", nil}
	chairRoom := Furniture{"стул", []*Thing{&backpack}}
	room.addFurniture(tableRoom, chairRoom)

	player.addTask("надо собрать рюкзак и идти в универ. ")

	player.setPlace(kitchen)
}

func handleCommand(command string) string {
	/*
		данная функция принимает команду от "пользователя"
		и наверняка вызывает какой-то другой метод или функцию у "мира" - списка комнат
	*/
	commands := strings.Split(command, " ")
	switch commands[0] {
	case "осмотреться":
		return player.lookUp()
	case "идти":
		return player.playerMove(commands[1])
	case "надеть":
		return player.fitOnYourself(commands[1])
	case "взять":
		return player.putInInventory(commands[1])
	case "применить":
		return player.useThing(commands[1], commands[2])
	default:
		return "неизвестная команда"
	}

}
