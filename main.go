package main

import (
	"errors"
	"fmt"
	"sync"
)

type User struct {
	ID   int
	Name string
}

type UserStore struct {
	mu    sync.RWMutex // combining the user with the mutex, so that the operations have the same mutex instance while dealing with the user
	users map[int]User
}

func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[int]User),
	}
}

func (s *UserStore) Create(user User) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; exists {
		return errors.New("user already exists")
	}

	s.users[user.ID] = user
	return nil
}

func (s *UserStore) Get(id int) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]

	if !exists {
		return User{}, errors.New("user not found")
	}
	return user, nil
}

func (s *UserStore) Update(id int, name string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]

	if !exists {
		return errors.New("user not found")
	}

	user.Name = name
	s.users[id] = user

	return nil
}

func (s *UserStore) Delete(id int) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.users[id]

	if !exists {
		return errors.New("user not found")
	}

	delete(s.users, id)

	return nil
}

func (s *UserStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]User, 0, len(s.users))

	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

func main() {

	store := NewUserStore()

	if err := store.Create(User{ID: 1, Name: "Dhwanil"}); err != nil {
		fmt.Println(err)
		return
	}

	user, err := store.Get(1)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("read:", user)

	if err := store.Update(1, "Alex"); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("all users:", store.List())

	if err := store.Delete(1); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("after deletion:", store.List())
}
