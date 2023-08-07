package demo

import (
	"encoding/json"
	"fmt"
	"math/rand"

	db "github.com/PoulIgorson/sub_engine_fiber/database"
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
	return bct.Save(car)
}

func (car *Car) Delete(db_ DB) error {
	bct, _ := db_.Table("car", &Car{})
	return bct.Delete(car)
}

func CreateModels(db_ DB) {
	models := []string{"BMW", "Volvo", "Porch", "WW", "Tesla", "Bug"}
	colors := []string{"red", "green", "blue", "white", "black", "pink"}
	cities := []string{"Moscow", "SP", "Vladimir", "Paris", "Rostov"}

	carBct, _ := db_.Table("car", &Car{})
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
	fmt.Printf(" %15v | %10v | %10v | %10v\n", "ID", "model", "color", "city")
	fmt.Printf("---------------- | ---------- | ---------- | ----------\n")
	for _, carM := range cars {
		car := carM.(*Car)
		fmt.Printf(" %15v | %10v | %10v | %10v\n", car.ID, car.Model, car.Color, car.City)
	}
}

func Run() {
	identity := ""
	password := ""
	db_, err := db.OpenPocketBaseLocal(identity, password)
	//db_, err := db.OpenBbolt("sub_engine_fiber_db.db")
	if err != nil {
		panic(err)
	}
	defer db_.Close()
	fmt.Println("App start")
	fmt.Println("Please if you use pocketbase go to admin (login if required)")
	fmt.Println("Create tables and set custom rules for demo of interface to pb:")
	fmt.Println("\tcar with fields (model string, color string, city string)")
	// fmt.Println("If you not want set custom rules, then set `identity` and `password` in demo/demo_db.go")

	// CreateModels(db_)

	carBct, err := db_.Table("car", &Car{})
	if err != nil {
		panic(err)
	}
	car := &Car{
		Model: "BMW",
		Color: "pink2",
		City:  "Moscow",
	}
	if err := car.Save(carBct); err != nil {
		fmt.Println(err)
	}

	// carBct.Delete(carBct.Objects.Count() - 3)
	// cars := carBct.Objects.Filter(db.Params{"Model": "Bug"}, db.Params{"Color": "black", "City": "Moscow"})
	cars := carBct.Manager().Filter(Params{"Color": "pink2"})
	show(cars.All())
	for {
	}
}
