package demo

import (
	"encoding/json"
	"fmt"
	"math/rand"

	bucket "github.com/PoulIgorson/sub_engine_fiber/database"
	db "github.com/PoulIgorson/sub_engine_fiber/database"
)

type Car struct {
	ID    uint   `json:"id"`
	Model string `json:"model"`
	Color string `json:"color"`
	City  string `json:"city"`
}

func (car Car) Id() uint {
	return car.ID
}

func Create(db_ *db.DB, carStr string) *Car {
	car := &Car{}
	json.Unmarshal([]byte(carStr), car)
	return car
}

func (car Car) Create(db_ *db.DB, carStr string) db.Model {
	return Create(db_, carStr)
}

func (car *Car) Save(bct *bucket.Bucket) error {
	return bucket.SaveModel(bct, car)
}

func CreateModels(db_ *db.DB) {
	models := []string{"BMW", "Volvo", "Porch", "WW", "Tesla", "Bug"}
	colors := []string{"red", "green", "blue", "white", "black", "pink"}
	cities := []string{"Moscow", "SP", "Vladimir", "Paris", "Rostov"}

	carBct, _ := db_.Bucket("car", Car{})
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

func show(cars []db.Model) {
	fmt.Printf("%5v | %10v | %10v | %10v\n", "ID", "model", "color", "city")
	fmt.Printf("----- | ---------- | ---------- | ----------\n")
	for _, carM := range cars {
		car := carM.(*Car)
		fmt.Printf("%5v | %10v | %10v | %10v\n", car.ID, car.Model, car.Color, car.City)
	}
}

func Run() {
	db_, err := db.Open("sub_engine_fiber_db.db")
	if err != nil {
		panic(err)
	}
	defer db_.Close()
	fmt.Println("App start")

	// CreateModels(db_)

	carBct, _ := db_.Bucket("car", Car{})
	car := &Car{
		Model: "BMW",
		Color: "pink2",
		City:  "Moscow",
	}
	car.Save(carBct)

	// carBct.Delete(carBct.Objects.Count() - 3)
	// cars := carBct.Objects.Filter(db.Params{"Model": "Bug"}, db.Params{"Color": "black", "City": "Moscow"})
	cars := carBct.Objects.Filter(db.Params{"Color": "pink2"})
	show(cars.All())
}
