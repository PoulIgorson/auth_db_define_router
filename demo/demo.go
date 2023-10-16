package demo

import (
	"fmt"
	"math/rand"
	"time"

	db "github.com/PoulIgorson/sub_engine_fiber/database"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
)

type Car struct {
	ID       string `json:"id"`
	ModelCar string `json:"modelCar"`
	Color    string `json:"color"`
	City     string `json:"city"`
	Year     uint   `json:"year"`
	InSale   PBTime `json:"inSale"`
}

func (car Car) Id() any {
	return car.ID
}

func (Car) Create(db_ DB, carStr string) Model {
	car := &Car{}
	//json.Unmarshal([]byte(carStr), car)
	JSONParse([]byte(carStr), car)
	return car
}

func (car *Car) Save(bct Table) error {
	return bct.Save(car)
}

func (car *Car) Delete(db_ DB) error {
	bct, _ := db_.Table("car", &Car{})
	return bct.Delete(car.ID)
}

func createModels(db_ DB, count int) {
	models := []string{"BMW", "Volvo", "Porch", "WW", "Tesla", "Bug"}
	colors := []string{"red", "green", "blue", "white", "black", "pink"}
	cities := []string{"Moscow", "SP", "Vladimir", "Paris", "Rostov"}

	carBct, _ := db_.Table("car", &Car{})
	for i := 0; i < count; i++ {
		car := &Car{
			ModelCar: models[rand.Int()%len(models)],
			Color:    colors[rand.Int()%len(colors)],
			City:     cities[rand.Int()%len(cities)],
			Year:     uint(2000 + rand.Int()%30),
			InSale:   PBTime(time.Now()),
		}
		if err := car.Save(carBct); err != nil {
			fmt.Printf("create cars[%v]: %v\n", i, err)
		}
		fmt.Printf("saved car: %+v\n", car)
	}
}

func showCars(cars []Model) {
	fmt.Printf(" %15v | %10v | %10v | %10v | %10v\n", "ID", "modelCar", "color", "city", "year")
	fmt.Printf("---------------- | ---------- | ---------- | ---------- | ----------\n")
	for _, carM := range cars {
		car, ok := carM.(*Car)
		if !ok {
			continue
		}
		fmt.Printf(" %15v | %10v | %10v | %10v | %10v\n", car.ID, car.ModelCar, car.Color, car.City, car.Year)
	}
}

func Run() {
	fmt.Println("opening database")
	db_, err := db.OpenPocketBase("http://localhost:8090", "backend@mail.com", "backendbackend", true, true)
	//db_, err := db.OpenPocketBaseLocal("backend@mail.com", "backendbackend", true)
	//db_, err := db.OpenBbolt("sub_engine_fiber_db.db")
	if err != nil {
		panic("db: " + err.Error())
	}
	defer db_.Close()

	_, err = db_.Table("car", &Car{})
	for err != nil {
		_, err = db_.Table("car", &Car{})
	}

	fmt.Println("getting table")
	table, err := db_.Table("car", &Car{})
	if err != nil {
		panic("DB.Table: " + err.Error())
	}

	// filling database
	createModels(db_, 10)

	fmt.Println("all models")
	showCars(table.Manager().All())

	fmt.Println("creating model")
	car := &Car{
		ModelCar: "BMW",
		Color:    "pink",
		City:     "Moscow",
		Year:     3000,
		InSale:   PBTime(time.Now()),
	}
	fmt.Println("saving model")
	if err := car.Save(table); err != nil {
		panic("Car.Save: " + err.Error())
	}

	fmt.Printf("saved car: %+v\n", car)

	fmt.Println("filtering models - ModelCar=\"BMW\"")
	cars := table.Manager().Filter(Params{"ModelCar": "BMW"})
	showCars(cars.All())

	fmt.Println("deleting our model")
	if err := car.Delete(db_); err != nil {
		panic("car.Delete: " + err.Error())
	}

	fmt.Println("filtering - ModelCar=\"BMW\"")
	cars = table.Manager().Filter(Params{"ModelCar": "BMW"})
	showCars(cars.All())

	var c string
	fmt.Print("press Enter to exit: ")
	fmt.Scanf("%s", &c)
	fmt.Println("")

	fmt.Println("deleting all models in table")
	if err := table.DeleteAll(); err != nil {
		panic("table.DeleteAll: " + err.Error())
	}

	fmt.Println("exit")
}
