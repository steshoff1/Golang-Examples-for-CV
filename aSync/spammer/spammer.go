package main

import (
	"fmt"
	"sort"
	"sync"
)

func RunPipeline(cmds ...cmd) {
	in := make(chan interface{})
	wg := &sync.WaitGroup{}

	for _, com := range cmds {
		out := make(chan interface{})
		wg.Add(1)
		go func(in chan interface{}, out chan interface{}, com cmd, wg *sync.WaitGroup) {
			com(in, out)
			close(out)
			wg.Done()
		}(in, out, com, wg)
		in = out

	}
	wg.Wait()
}

func SelectUsers(in, out chan interface{}) {
	list := make(map[uint64]string)
	str := ""
	user := User{}
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for i := range in {
		str = fmt.Sprintf("%v", i)
		wg.Add(1)
		go func(str string, wg *sync.WaitGroup, list *map[uint64]string, user User) {
			user = GetUser(str)
			mu.Lock()
			_, ok := (*list)[user.ID]
			if !ok {
				(*list)[user.ID] = str
				out <- user
			}
			mu.Unlock()
			wg.Done()
		}(str, wg, &list, user)
	}
	wg.Wait()
}

func SelectMessages(in, out chan interface{}) {
	// 	in - User
	// 	out - MsgID
	flag := false
	users := make([]User, 0)
	wg := &sync.WaitGroup{}
	for i := range in {
		if flag {
			flag = false
			wg.Add(1)
			users = append(users, i.(User))
			go func(users []User, wg *sync.WaitGroup) {
				if res, err := GetMessages(users...); err != nil {
					fmt.Printf("Error: %v", err)
				} else {
					for _, j := range res {
						out <- j
					}
				}
				wg.Done()
			}(users, wg)
			users = make([]User, 0)
		} else {
			users = append(users, i.(User))
			flag = true
		}
	}
	if flag {
		if res, err := GetMessages(users...); err != nil {
			fmt.Printf("Error: %v", err)
		} else {
			for _, j := range res {
				out <- j
			}
		}
	}
	wg.Wait()
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData
	res := MsgData{}
	wg := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(in chan interface{}, res MsgData, wg *sync.WaitGroup) {
			for inp := range in {

				spam, err := HasSpam(inp.(MsgID))
				if err == nil {
					res.ID = inp.(MsgID)
					res.HasSpam = spam
					out <- res
				} else {
					fmt.Printf("Error: %v", err)
				}

			}
			wg.Done()
		}(in, res, wg)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
	mas := make([]MsgData, 0)
	for i := range in {
		mas = append(mas, i.(MsgData))
	}
	sort.Slice(mas, func(k, l int) bool {
		if mas[k].HasSpam == mas[l].HasSpam {
			return mas[k].ID < mas[l].ID
		}
		return mas[k].HasSpam
	})
	for _, j := range mas {
		out <- fmt.Sprintf("%v %v", j.HasSpam, j.ID)
	}

}
