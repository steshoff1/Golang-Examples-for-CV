package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

var path string = "dataset.xml"

type Subject struct {
	ID            int    `xml:"id"`
	GuID          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favorite"`
}

type Subjects struct {
	XMLName  xml.Name  `xml:"root"`
	Subjects []Subject `xml:"row"`
}

func parseFromFile() (*Subjects, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return &Subjects{}, fmt.Errorf("parser: %w", err)
	}
	subjects := new(Subjects)
	err = xml.Unmarshal(data, subjects)
	if err != nil {
		return subjects, fmt.Errorf("parser: %w", err)
	}
	return subjects, nil

}

type DataErr struct {
	err string
}

func (d DataErr) Error() string {
	return d.err
}

func packError(err error) string {
	return `{"error" : "` + err.Error() + `"}`
}

func exportFromSR(sr *SearchRequest, url url.Values) error {
	var err error
	sr.Limit, err = strconv.Atoi(url.Get("limit"))
	if err != nil {
		return err
	}
	sr.Offset, err = strconv.Atoi(url.Get("offset"))
	if err != nil {
		return err
	}

	sr.OrderField = url.Get("order_field")
	sr.Query = url.Get("query")

	sr.OrderBy, err = strconv.Atoi(url.Get("order_by"))
	if err != nil {
		return err
	}

	return nil
}

func findByNameOrAbout(s *Subjects, substr string) []Subject {
	res := []Subject{}
	for _, subj := range s.Subjects {
		if strings.Contains(subj.About, substr) || strings.Contains(subj.FirstName+subj.LastName, substr) {
			res = append(res, subj)
		}
	}
	return res
}

func restruct(subjs []Subject) []User {
	users := make([]User, len(subjs))
	for i := 0; i < len(subjs); i++ {
		users[i].ID = subjs[i].ID
		users[i].Name = subjs[i].FirstName + subjs[i].LastName
		users[i].Age = subjs[i].Age
		users[i].Gender = subjs[i].Gender
		users[i].About = subjs[i].About
	}
	return users
}

func mySort(users []User, orderBy int, orderField string) error {
	if orderBy == OrderByAsIs {
		return nil
	}

	if orderBy != OrderByDesc && orderBy != OrderByAsc {
		return DataErr{"wrong value in orderBy"}
	}

	switch {
	case orderField == "Name" || orderField == "":
		sort.Slice(users, func(i, j int) bool {
			if len(users[i].Name) != len(users[j].Name) {
				return len(users[i].Name) < len(users[j].Name) == (orderBy == OrderByAsc)
			}
			for k := 0; k < len(users[i].Name); k++ {
				if users[i].Name[k] != users[j].Name[k] {
					return users[i].Name[k] < users[j].Name[k] == (orderBy == OrderByAsc)
				}
			}
			return false
		})

	case orderField == "Age":
		sort.Slice(users, func(i, j int) bool {
			return users[i].Age < users[j].Age == (orderBy == OrderByAsc)
		})

	case orderField == "ID":
		sort.Slice(users, func(i, j int) bool {
			return users[i].ID < users[j].ID == (orderBy == OrderByAsc)
		})
	default:
		return DataErr{"OrderField invalid"}
	}

	return nil
}

// тут писать SearchServer
func SearchServer(w http.ResponseWriter, r *http.Request) {
	accesToken := r.Header.Get("AccessToken")
	if accesToken == "" {
		http.Error(w, packError(DataErr{"bad AccessToken"}), http.StatusUnauthorized)
		return
	}

	url := r.URL.Query()
	var sr SearchRequest

	err := exportFromSR(&sr, url)
	if err != nil {
		http.Error(w, packError(err), http.StatusBadRequest)
		return
	}

	allSubjects, err := parseFromFile()
	if err != nil {
		http.Error(w, fmt.Errorf("SearchServer: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	subjs := findByNameOrAbout(allSubjects, sr.Query)
	users := restruct(subjs)

	err = mySort(users, sr.OrderBy, sr.OrderField)
	if err != nil {
		http.Error(w, packError(err), http.StatusBadRequest)
		return
	}
	len := len(users)
	if sr.Limit+sr.Offset < len {
		len = sr.Limit + sr.Offset
	}

	json, err := json.Marshal(users[sr.Offset:len])
	if err != nil {
		http.Error(w, fmt.Errorf("SearchServer: %w", err).Error(), http.StatusInternalServerError)
	}
	_, err = w.Write(json)
	if err != nil {
		http.Error(w, fmt.Errorf("SearchServer: %w", err).Error(), http.StatusInternalServerError)
	}

}
