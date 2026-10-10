// package main

// import (
// 	"errors"
// )

// var ErrTruckNotFound = errors.New("truck not found")

// type FleetManager interface {
// 	AddTruck(id string, cargo int) error
// 	GetTruck(id string) (*Truck, error)
// 	RemoveTruck(id string) error
// 	UpdateTruckCargo(id string, cargo int) error
// }

// type Truck struct {
// 	ID    string
// 	Cargo int
// }

// type truckManager struct {
// 	trucks map[string]*Truck
// }

// func NewTruckManager() truckManager {
// 	return truckManager{
// 		trucks: make(map[string]*Truck),
// 	}
// }

// func (tm truckManager) AddTruck(id string, cargo int) error {
// 	tm.trucks[id] = &Truck{
// 		ID:    id,
// 		Cargo: cargo,
// 	}
// 	return nil
// }

// func (tm truckManager) GetTruck(id string) (*Truck, error) {
// 	truck, ok := tm.trucks[id]

// 	if !ok {
// 		return nil, ErrTruckNotFound
// 	}
// 	return truck, nil
// }

// func (tm truckManager) UpdateTruckCargo(id string, cargo int) error {
// 	truck, ok := tm.trucks[id]

// 	if !ok {
// 		return ErrTruckNotFound
// 	}

// 	truck.Cargo = cargo

// 	return nil
// }

// func (tm truckManager) RemoveTruck(id string) error {

// 	_, ok := tm.trucks[id]

// 	if !ok {
// 		return ErrTruckNotFound
// 	}

// 	delete(tm.trucks, id)
// 	return nil
// }

package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, jobs chan int, results chan int, wg *sync.WaitGroup) {

	defer wg.Done()

	for job := range jobs {
		fmt.Println("worker", id, "got job", job)
		result := job * job
		time.Sleep(time.Second)
		results <- result
	}
}

func main() {

	jobs := make(chan int)
	results := make(chan int, 5)

	var wg sync.WaitGroup

	wg.Add(3)

	go worker(1, jobs, results, &wg)
	go worker(2, jobs, results, &wg)
	go worker(3, jobs, results, &wg)

	jobs <- 100
	jobs <- 200
	jobs <- 300
	jobs <- 400
	jobs <- 500

	close(jobs) // used not to free memory, but we will not send data into the jobs channel from now onwards

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("results from jobs are:")

	for result := range results {
		fmt.Println(result)
	}

	fmt.Println("all workers finished")
}
