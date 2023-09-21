package storage

import (
	"fmt"

	"gorm.io/gorm"
	//"github.com/kindenko/gophermart/internal/api"
)

type UserDB struct {
	db *gorm.DB
}

type UserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u UserDB) GetUserByLogin(login string) {
	_ = "select u.login from users u where u.login =%1"

	var user = UserRequest{Login: login}

	row := u.db.Find(user)

	fmt.Println(row)
	//row := u.db.First(&api.UserRequest, api.UserRequest{Login: login})

	// row := u.db.QueryRow(query, login)
	// if err := row.Scan(&long, &isDeleted); err != nil {
	// 	log.Println("Failed to get link from db")
	// 	log.Println(err)
	// 	return "Error in Get from db", 0, err
	// }

	// return long, isDeleted, nil
}
