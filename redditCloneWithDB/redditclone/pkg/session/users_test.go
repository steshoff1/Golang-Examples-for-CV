package session

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/user"
	"reflect"
	"strings"
	"testing"
	"time"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func ErrCheck(mock sqlmock.Sqlmock, err error, t *testing.T) {
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
}

func TestFind(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	// good query
	rows := sqlmock.NewRows([]string{"id", "username", "password"})
	expect := []*user.User{
		{
			Id:   "1",
			Log:  "1",
			Pass: "strongpassword1",
		},
		{
			Id:   "2",
			Log:  "abracadabra",
			Pass: "strongpassword2",
		},
	}
	for _, usr := range expect {
		rows = rows.AddRow(usr.Id, usr.Log, usr.Pass)
	}
	U := &SqlUsers{DB: db}
	for i, u := range expect {
		mock.
			ExpectQuery("SELECT id, username, password FROM users WHERE").
			WithArgs(u.Log).
			WillReturnRows(rows)
		usr, err := U.Find(u.Log)

		ErrCheck(mock, err, t)
		if !reflect.DeepEqual(usr, *expect[i]) {
			t.Errorf("results not match, want %v, have %v", *expect[0], usr)
			return
		}
	}
	// query error
	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE").
		WithArgs("abc").
		WillReturnError(fmt.Errorf("bad query"))

	_, err = U.Find("abc")

	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	// row scan error
	rows = sqlmock.NewRows([]string{"id", "title"}).
		AddRow(1, "title")

	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE").
		WithArgs("abc").
		WillReturnRows(rows)

	_, err = U.Find("abc")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

}

func TestCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	U := &SqlUsers{
		DB: db,
	}
	login := "title"
	password := "description"
	expect := []*user.User{
		{
			Id:   "1",
			Log:  login,
			Pass: password,
		},
	}
	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(fmt.Sprintf(`{"username":"%s","password":"%s"}`, login, password)))

	//ok query
	mock.
		ExpectExec(`INSERT INTO users`).
		WithArgs(login, password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	usr, err := U.CreateUser(req)

	ErrCheck(mock, err, t)
	if !reflect.DeepEqual(usr, expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect[0], usr)
		return
	}
	// query error
	req = httptest.NewRequest("POST", "/api/register", strings.NewReader(fmt.Sprintf(`{"username":"%s","password":"%s"}`, login, password)))
	mock.
		ExpectExec(`INSERT INTO users`).
		WithArgs(login, password).
		WillReturnError(fmt.Errorf("bad query"))

	_, err = U.CreateUser(req)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	// result error(last insert id idk why it)
	req = httptest.NewRequest("POST", "/api/register", strings.NewReader(fmt.Sprintf(`{"username":"%s","password":"%s"}`, login, password)))
	mock.
		ExpectExec(`INSERT INTO users`).
		WithArgs(login, password).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("bad_result")))

	_, err = U.CreateUser(req)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	// user has already registered
	req = httptest.NewRequest("POST", "/api/register", strings.NewReader(fmt.Sprintf(`{"username":"%s","password":"%s"}`, login, password)))
	rows := sqlmock.NewRows([]string{"id", "username", "password"})
	rows = rows.AddRow(expect[0].Id, login, password)
	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE").
		WithArgs(login).
		WillReturnRows(rows)

	_, err = U.CreateUser(req)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	// wrong map
	requests := []*http.Request{
		httptest.NewRequest("POST", "/api/register", strings.NewReader(fmt.Sprintf(`{"abs""%s","password":"%s"}`, login, password))),
		httptest.NewRequest("POST", "/api/register", strings.NewReader(fmt.Sprintf(`{"username":"%s","password":""}`, login))),
		httptest.NewRequest("POST", "/api/register", strings.NewReader(fmt.Sprintf(`{"username":"","password":"%s"}`, password))),
		httptest.NewRequest("POST", "/api/register", strings.NewReader(`{"username":"username","passwo":""}`)),
	}
	for i, v := range requests {
		_, err = U.CreateUser(v)
		if err == nil {
			t.Errorf("wrong map : test case [%d] : expected error, got nil", i)
			return
		}
	}
}

type TestCaseValid struct {
	token string
	exp   time.Duration
	res   bool
	err   error
}

func TestValid(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	U := &SqlUsers{
		DB: db,
	}
	cases := []TestCaseValid{
		{
			token: "1",
			exp:   time.Duration(time.Now().AddDate(0, 0, 7).Unix()),
			res:   true,
		},
		{
			token: "2",
			exp:   time.Duration(time.Now().AddDate(0, 0, -7).Unix()),
			err:   fmt.Errorf("error"),
			res:   false,
		},
		{
			token: "3",
			exp:   time.Duration(time.Now().AddDate(0, 0, -7).Unix()),
			res:   false,
		},
	}
	rows := sqlmock.NewRows([]string{"exp"})

	for _, tcv := range cases {
		rows = rows.AddRow(tcv.exp)
	}

	for i, tcv := range cases {
		mock.
			ExpectQuery("SELECT exp FROM sessions WHERE").
			WithArgs(tcv.token).
			WillReturnRows(rows)
		if tcv.res == false {
			mock.
				ExpectExec("DELETE FROM sessions WHERE").
				WithArgs(tcv.token).WillReturnResult(sqlmock.NewResult(0, 1)).WillReturnError(tcv.err)
		}

		ok, err := U.Valid(tcv.token)
		if err != nil && tcv.err == nil {
			t.Errorf("case [%d]: unexpected err: %s", i, err)
			return
		}

		if err == nil && tcv.err != nil {
			t.Errorf("case [%d]: expected errror %s got nil", i, tcv.err)
			return
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
			return
		}

		if tcv.res != ok && tcv.err == nil {
			t.Errorf("case [%d]: results not match, want %v, have %v", i, tcv.res, ok)
			return
		}
	}
	// not found
	mock.
		ExpectQuery("SELECT exp FROM sessions WHERE").
		WithArgs("abc").
		WillReturnError(fmt.Errorf("bad query"))

	_, err = U.Valid("abc")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	// row scan error
	rows = sqlmock.NewRows([]string{"exp"}).
		AddRow("fdbdfbdf")

	mock.
		ExpectQuery("SELECT exp FROM sessions WHERE").
		WithArgs("abc").
		WillReturnRows(rows)

	_, err = U.Valid("abc")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

func TestAddSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	U := &SqlUsers{
		DB: db,
	}
	expect := []*user.User{
		{
			Id:   "1",
			Log:  "title",
			Pass: "description",
		},
	}
	//ok query
	token, err := MakeJwt(user.User{
		Id:   expect[0].ID(),
		Log:  expect[0].Login(),
		Pass: expect[0].Password(),
	})

	if err != nil {
		t.Errorf("must make JWT bat cant : %s", err)
		return
	}
	mock.
		ExpectExec(`INSERT INTO sessions`).
		WithArgs(sqlmock.AnyArg(), token).
		WillReturnResult(sqlmock.NewResult(1, 1))
	_, err = U.AddSession(expect[0])

	ErrCheck(mock, err, t)
	// query error
	mock.
		ExpectExec(`INSERT INTO sessions`).
		WithArgs(sqlmock.AnyArg(), token).
		WillReturnError(fmt.Errorf("bad query"))

	_, err = U.AddSession(expect[0])
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
