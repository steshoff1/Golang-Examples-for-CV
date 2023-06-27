package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

// тут писать код тестов
type TestCase struct {
	req        SearchRequest
	src        SearchClient
	resp       *SearchResponse
	path       string
	ExistError bool
	Error      string
}

type UserTest struct {
	User
	err complex64
}

func TimeOutServer(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 5)
}

func BadJSONFileServerGoodResponse(w http.ResponseWriter, r *http.Request) {
	// случай когда зареспонсился
	jsonBrokeData := `{"id : awdoido}`
	_, err := w.Write([]byte(jsonBrokeData))
	if err != nil {
	}

}

func BadJSONFileServerBadResponse(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "adwad", http.StatusBadRequest)
}

func TestSearchServer(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	testTimeOutServer := httptest.NewServer(http.HandlerFunc(TimeOutServer))
	testBadJSONServerGood := httptest.NewServer(http.HandlerFunc(BadJSONFileServerGoodResponse))
	testBadJSONServerBad := httptest.NewServer(http.HandlerFunc(BadJSONFileServerBadResponse))
	testStopServer := httptest.NewUnstartedServer(http.HandlerFunc(SearchServer))
	cases := []TestCase{

		{
			// проверка работоспособности Query
			req: SearchRequest{20, 0, "KaneSharp", "", 1},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path: "dataset.xml",
			resp: &SearchResponse{
				Users: []User{
					{
						ID:     34,
						Name:   "KaneSharp",
						Age:    34,
						About:  "Lorem proident sint minim anim commodo cillum. Eiusmod velit culpa commodo anim consectetur consectetur sint sint labore. Mollit consequat consectetur magna nulla veniam commodo eu ut et. Ut adipisicing qui ex consectetur officia sint ut fugiat ex velit cupidatat fugiat nisi non. Dolor minim mollit aliquip veniam nostrud. Magna eu aliqua Lorem aliquip.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			ExistError: false,
			Error:      "",
		},
		{
			// проверка сортировки при разных параметрах
			req: SearchRequest{50, 0, "laborum c", "Age", -1},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path: "dataset.xml",
			resp: &SearchResponse{
				Users: []User{
					{
						ID:     32,
						Name:   "ChristyKnapp",
						Age:    40,
						About:  "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n",
						Gender: "female",
					},
					{
						ID:     33,
						Name:   "TwilaSnow",
						Age:    36,
						About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
						Gender: "female",
					},
					{
						ID:     12,
						Name:   "CruzGuerrero",
						Age:    36,
						About:  "Sunt enim ad fugiat minim id esse proident laborum magna magna. Velit anim aliqua nulla laborum consequat veniam reprehenderit enim fugiat ipsum mollit nisi. Nisi do reprehenderit aute sint sit culpa id Lorem proident id tempor. Irure ut ipsum sit non quis aliqua in voluptate magna. Ipsum non aliquip quis incididunt incididunt aute sint. Minim dolor in mollit aute duis consectetur.\n",
						Gender: "male",
					},
					{
						ID:     4,
						Name:   "OwenLynn",
						Age:    30,
						About:  "Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			ExistError: false,
			Error:      "",
		},
		{
			// проверка сортировки при разных параметрах
			req: SearchRequest{3, 0, "laborum c", "ID", 1},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path: "dataset.xml",
			resp: &SearchResponse{
				Users: []User{
					{
						ID:     4,
						Name:   "OwenLynn",
						Age:    30,
						About:  "Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n",
						Gender: "male",
					},
					{
						ID:     12,
						Name:   "CruzGuerrero",
						Age:    36,
						About:  "Sunt enim ad fugiat minim id esse proident laborum magna magna. Velit anim aliqua nulla laborum consequat veniam reprehenderit enim fugiat ipsum mollit nisi. Nisi do reprehenderit aute sint sit culpa id Lorem proident id tempor. Irure ut ipsum sit non quis aliqua in voluptate magna. Ipsum non aliquip quis incididunt incididunt aute sint. Minim dolor in mollit aute duis consectetur.\n",
						Gender: "male",
					},
					{
						ID:     32,
						Name:   "ChristyKnapp",
						Age:    40,
						About:  "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n",
						Gender: "female",
					},
				},
				NextPage: true,
			},
			ExistError: false,
			Error:      "",
		},
		{
			// проверка сортировки при разных параметрах
			req: SearchRequest{3, 0, "laborum c", "Name", 0},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path: "dataset.xml",
			resp: &SearchResponse{
				Users: []User{
					{
						ID:     4,
						Name:   "OwenLynn",
						Age:    30,
						About:  "Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n",
						Gender: "male",
					},
					{
						ID:     12,
						Name:   "CruzGuerrero",
						Age:    36,
						About:  "Sunt enim ad fugiat minim id esse proident laborum magna magna. Velit anim aliqua nulla laborum consequat veniam reprehenderit enim fugiat ipsum mollit nisi. Nisi do reprehenderit aute sint sit culpa id Lorem proident id tempor. Irure ut ipsum sit non quis aliqua in voluptate magna. Ipsum non aliquip quis incididunt incididunt aute sint. Minim dolor in mollit aute duis consectetur.\n",
						Gender: "male",
					},
					{
						ID:     32,
						Name:   "ChristyKnapp",
						Age:    40,
						About:  "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n",
						Gender: "female",
					},
				},
				NextPage: true,
			},
			ExistError: false,
			Error:      "",
		},
		{
			// проверка сортировки при разных параметрах
			req: SearchRequest{3, 0, "laborum c", "", 1},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path: "dataset.xml",
			resp: &SearchResponse{
				Users: []User{
					{
						ID:     4,
						Name:   "OwenLynn",
						Age:    30,
						About:  "Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n",
						Gender: "male",
					},
					{
						ID:     33,
						Name:   "TwilaSnow",
						Age:    36,
						About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
						Gender: "female",
					},
					{
						ID:     32,
						Name:   "ChristyKnapp",
						Age:    40,
						About:  "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n",
						Gender: "female",
					},
				},
				NextPage: true,
			},
			ExistError: false,
			Error:      "",
		},
		{
			// проверка сортировки при разных параметрах
			req: SearchRequest{3, 0, "laborum c", "", 1},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path: "dataset.xml",
			resp: &SearchResponse{
				Users: []User{
					{
						ID:     4,
						Name:   "OwenLynn",
						Age:    30,
						About:  "Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n",
						Gender: "male",
					},
					{
						ID:     33,
						Name:   "TwilaSnow",
						Age:    36,
						About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
						Gender: "female",
					},
					{
						ID:     32,
						Name:   "ChristyKnapp",
						Age:    40,
						About:  "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n",
						Gender: "female",
					},
				},
				NextPage: true,
			},
			ExistError: false,
			Error:      "",
		},
		{
			// проверка сортировки при разных параметрах
			req: SearchRequest{3, 4, "ut", "ID", 1},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path: "dataset.xml",
			resp: &SearchResponse{
				Users: []User{
					{
						ID:     6,
						Name:   "JenningsMays",
						Age:    39,
						About:  "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
						Gender: "male",
					},
					{
						ID:     7,
						Name:   "LeannTravis",
						Age:    34,
						About:  "Lorem magna dolore et velit ut officia. Cupidatat deserunt elit mollit amet nulla voluptate sit. Quis aute aliquip officia deserunt sint sint nisi. Laboris sit et ea dolore consequat laboris non. Consequat do enim excepteur qui mollit consectetur eiusmod laborum ut duis mollit dolor est. Excepteur amet duis enim laborum aliqua nulla ea minim.\n",
						Gender: "female",
					},
					{
						ID:     9,
						Name:   "RoseCarney",
						Age:    36,
						About:  "Voluptate ipsum ad consequat elit ipsum tempor irure consectetur amet. Et veniam sunt in sunt ipsum non elit ullamco est est eu. Exercitation ipsum do deserunt do eu adipisicing id deserunt duis nulla ullamco eu. Ad duis voluptate amet quis commodo nostrud occaecat minim occaecat commodo. Irure sint incididunt est cupidatat laborum in duis enim nulla duis ut in ut. Cupidatat ex incididunt do ullamco do laboris eiusmod quis nostrud excepteur quis ea.\n",
						Gender: "female",
					},
				},
				NextPage: true,
			},
			ExistError: false,
			Error:      "",
		},

		{
			// случай если есть xml файл в котором два одинаковых айдишника
			req: SearchRequest{20, 0, "HildaMayer", "", 1},
			src: SearchClient{
				AccessToken: "tokns",
				URL:         testServer.URL,
			},
			path: "dataset_double.xml",
			resp: &SearchResponse{
				Users: []User{
					{
						ID:     1,
						Name:   "HildaMayer",
						Age:    21,
						About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
					{
						ID:     35,
						Name:   "HildaMayer",
						Age:    21,
						About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
				},
				NextPage: false,
			},
			ExistError: false,
			Error:      "",
		},
		{
			// проверка на токен
			req: SearchRequest{10, 1, "", "", -1},
			src: SearchClient{
				AccessToken: "",
				URL:         testServer.URL,
			},
			path:       "dataset.xml",
			resp:       nil,
			ExistError: true,
			Error:      "bad AccessToken",
		},
		{
			// проверка на неверный путь
			req: SearchRequest{10, 1, "", "", 1},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path:       "ne_dataset.xml",
			resp:       nil,
			ExistError: true,
			Error:      "SearchServer fatal error",
		},
		{
			// проверка на битый xml
			req: SearchRequest{10, 1, "ut", "ID", 1},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path:       "brokeXML.xml",
			resp:       nil,
			ExistError: true,
			Error:      "SearchServer fatal error",
		},
		{
			// проверка на отрицательный limit
			req: SearchRequest{-1, 1, "", "", 1},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path:       "dataset.xml",
			resp:       nil,
			ExistError: true,
			Error:      "limit must be > 0",
		},
		{
			// проверка на отрицательный offset
			req: SearchRequest{10, -1, "", "", 1},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path:       "dataset.xml",
			resp:       nil,
			ExistError: true,
			Error:      "offset must be > 0",
		},
		{
			// проверка на неверный OrderField
			req: SearchRequest{10, 1, "", "lol", -1},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path:       "dataset.xml",
			resp:       nil,
			ExistError: true,
			Error:      "OrderFeld lol invalid",
		},
		{
			// проверка на неверный OrderField
			req: SearchRequest{10, 1, "ne_vazhno", "Id", 5},
			src: SearchClient{
				AccessToken: "token",
				URL:         testServer.URL,
			},
			path:       "dataset.xml",
			resp:       nil,
			ExistError: true,
			Error:      "unknown bad request error: wrong value in orderBy",
		},
		{
			// проверка на timeout
			req: SearchRequest{0, 0, "", "", 0},
			src: SearchClient{
				AccessToken: "token",
				URL:         testTimeOutServer.URL,
			},
			path:       "dataset.xml",
			resp:       nil,
			ExistError: true,
			Error:      "timeout for limit=1&offset=0&order_by=0&order_field=&query=",
		},
		{
			// сервер вернул json w/ bad status
			req: SearchRequest{0, 0, "", "", 0},
			src: SearchClient{
				AccessToken: "token",
				URL:         testBadJSONServerBad.URL,
			},
			path:       "dataset.xml",
			resp:       nil,
			ExistError: true,
			Error:      "cant unpack error json: invalid character 'a' looking for beginning of value",
		},
		{
			// сервер вернул json w/ bad status
			req: SearchRequest{0, 0, "", "", 0},
			src: SearchClient{
				AccessToken: "token",
				URL:         testBadJSONServerGood.URL,
			},
			path:       "dataset.xml",
			resp:       nil,
			ExistError: true,
			Error:      "cant unpack result json: unexpected end of JSON input",
		},
		{
			// test for broken server
			req: SearchRequest{0, 0, "", "", 0},
			src: SearchClient{
				AccessToken: "",
				URL:         testStopServer.URL,
			},
			path:       "dataset.xml",
			resp:       nil,
			ExistError: true,

			Error: "unknown error Get \"?limit=1&offset=0&order_by=0&order_field=&query=\": unsupported protocol scheme \"\"",
		},
	}
	for caseNum, item := range cases {
		path = item.path

		sr, err := item.src.FindUsers(item.req)

		if err != nil && !item.ExistError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.ExistError {
			t.Errorf("[%d] expected error,\n got nil", caseNum)
		}

		if err != nil && item.ExistError {
			if err.Error() != item.Error {
				t.Errorf("[%d] wrong error, expected \n%#v,got \n%#v", caseNum, item.Error, err.Error())
			}
		}

		if !reflect.DeepEqual(item.resp, sr) {
			t.Errorf("[%d] wrong result, expected \n%#v,\ngot \n%#v", caseNum, item.resp, sr)
		}

	}
	testServer.Close()
	testTimeOutServer.Close()
	testBadJSONServerBad.Close()
	testBadJSONServerGood.Close()
	testStopServer.Close()

}

type TestCase2 struct {
	URL     url.Values
	isError bool
}

func TestBadRequest(t *testing.T) {

	testCases := []TestCase2{
		{
			URL: url.Values{
				"limit":       {"20"},
				"offset":      {"trouble"},
				"query":       {""},
				"order_field": {"ID"},
				"order_by":    {"1"},
			},
			isError: true,
		},
		{
			URL: url.Values{
				"limit":       {"trouble"},
				"offset":      {"1"},
				"query":       {""},
				"order_field": {"Name"},
				"order_by":    {"1"},
			},
			isError: true,
		},
		{
			URL: url.Values{
				"limit":       {"10"},
				"offset":      {"1"},
				"query":       {""},
				"order_field": {"Age"},
				"order_by":    {"trouble"},
			},
			isError: true,
		},
	}

	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))

	for caseNum, v := range testCases {
		searcherReq, _ := http.NewRequest("GET", testServer.URL+"?"+v.URL.Encode(), nil) //nolint:errcheck
		searcherReq.Header.Add("AccessToken", "token")
		_, err := client.Do(searcherReq)

		if err != nil && v.isError {
			t.Errorf("[%d] expected error,\n got nil", caseNum)
		}
		if err != nil && !v.isError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}

	}
	testServer.Close()
}
