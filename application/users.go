package application

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/GehirnInc/crypt"
	"github.com/pkg/errors"

	_ "github.com/GehirnInc/crypt/md5_crypt"
	_ "github.com/GehirnInc/crypt/sha256_crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
)

type userManager struct {
}

func NewUserManager() *userManager {
	return &userManager{}
}

type User struct {
	UserName        string
	UserID, GroupID int
	UserInfo        string
	HomeDir         string
	Shell           string

	PasswordHash string
}

func (u *User) verifyUserPassword(password string) (ok bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("Panic: %v", r)
		}
	}()

	err = crypt.NewFromHash(u.PasswordHash).Verify(u.PasswordHash, []byte(password))
	if err != nil {
		if err == crypt.ErrKeyMismatch {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (um *userManager) getPasswordData() (map[string]string, error) {
	shadowData, err := ioutil.ReadFile("/etc/shadow")
	if err != nil {
		return nil, err
	}

	r := bufio.NewReader(bytes.NewReader(shadowData))

	data := map[string]string{}

	for {
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				return data, nil
			}

			return nil, err
		}

		lineParts := strings.Split(string(line), ":")

		data[lineParts[0]] = lineParts[1]
	}
}

func (um *userManager) getUsers() ([]*User, error) {
	var users []*User

	passwords, err := um.getPasswordData()
	if err != nil {
		return nil, err
	}

	passwdData, err := ioutil.ReadFile("/etc/passwd")
	if err != nil {
		return nil, err
	}

	r := bufio.NewReader(bytes.NewReader(passwdData))

	for {
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				return users, nil
			}

			return nil, err
		}

		lineParts := strings.Split(string(line), ":")

		u := &User{
			UserName: lineParts[0],
			UserInfo: lineParts[4],
			HomeDir:  lineParts[5],
			Shell:    lineParts[6],
		}

		if lineParts[1] == "x" {
			u.PasswordHash = passwords[u.UserName]
		} else {
			u.PasswordHash = lineParts[1]
		}

		u.UserID, err = strconv.Atoi(lineParts[2])
		if err != nil {
			return nil, err
		}

		u.GroupID, err = strconv.Atoi(lineParts[3])
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}
}

func (um *userManager) getUserByUserName(name string) (*User, error) {
	users, err := um.getUsers()
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if u.UserName == name {
			return u, nil
		}
	}

	return nil, nil
}
