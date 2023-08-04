package demo

import (
	"encoding/json"
	"fmt"
	"math/rand"

	db "github.com/PoulIgorson/sub_engine_fiber/database"
	bbolt "github.com/PoulIgorson/sub_engine_fiber/database/bbolt"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
)

type Car struct {
	ID    uint   `json:"id"`
	Model string `json:"model"`
	Color string `json:"color"`
	City  string `json:"city"`
}

func (car Car) Id() any {
	return car.ID
}

func Create(db_ DB, carStr string) *Car {
	car := &Car{}
	json.Unmarshal([]byte(carStr), car)
	return car
}

func (car Car) Create(db_ DB, carStr string) Model {
	return Create(db_, carStr)
}

func (car *Car) Save(bct Table) error {
	return bbolt.SaveModel(bct.(*bbolt.Bucket), car)
}

func (car Car) Delete(db_ DB) error {
	bct, _ := db_.Table("car", Car{})
	return bct.Delete(car)
}

func CreateModels(db_ DB) {
	models := []string{"BMW", "Volvo", "Porch", "WW", "Tesla", "Bug"}
	colors := []string{"red", "green", "blue", "white", "black", "pink"}
	cities := []string{"Moscow", "SP", "Vladimir", "Paris", "Rostov"}

	carBct, _ := db_.Table("car", Car{})
	for i := 0; i < 10; i++ {
		car := &Car{
			Model: models[rand.Int()%len(models)],
			Color: colors[rand.Int()%len(colors)],
			City:  cities[rand.Int()%len(cities)],
		}
		if err := car.Save(carBct); err != nil {
			fmt.Println(i, err)
		}
	}
}

func show(cars []Model) {
	fmt.Printf("%5v | %10v | %10v | %10v\n", "ID", "model", "color", "city")
	fmt.Printf("----- | ---------- | ---------- | ----------\n")
	for _, carM := range cars {
		car := carM.(*Car)
		fmt.Printf("%5v | %10v | %10v | %10v\n", car.ID, car.Model, car.Color, car.City)
	}
}

func Run() {
	db_, err := db.OpenBbolt("sub_engine_fiber_db.db")
	if err != nil {
		panic(err)
	}
	defer db_.Close()
	fmt.Println("App start")

	// CreateModels(db_)

	carBct, _ := db_.Table("car", Car{})
	car := &Car{
		Model: "BMW",
		Color: "pink2",
		City:  "Moscow",
	}
	car.Save(carBct)

	// carBct.Delete(carBct.Objects.Count() - 3)
	// cars := carBct.Objects.Filter(db.Params{"Model": "Bug"}, db.Params{"Color": "black", "City": "Moscow"})
	cars := carBct.Manager().Filter(Params{"Color": "pink2"})
	show(cars.All())
}
