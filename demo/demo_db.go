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
	bct, _ := db_.Table("", &Car{})
	return bct.Delete(car.ID)
}

func CreateModels(db_ DB) {
	models := []string{"BMW", "Volvo", "Porch", "WW", "Tesla", "Bug"}
	colors := []string{"red", "green", "blue", "white", "black", "pink"}
	cities := []string{"Moscow", "SP", "Vladimir", "Paris", "Rostov"}

	carBct, _ := db_.Table("", &Car{})
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

func showCar(cars []Model) {
	fmt.Printf(" %15v | %10v | %10v | %10v\n", "ID", "model", "color", "city")
	fmt.Printf("---------------- | ---------- | ---------- | ----------\n")
	for _, carM := range cars {
		car, ok := carM.(*Car)
		if !ok {
			continue
		}
		fmt.Printf(" %15v | %10v | %10v | %10v\n", car.ID, car.Model, car.Color, car.City)
	}
}

type RedCar_withHand struct {
	ID    uint
	Hand  bool
	CarId uint

	Car *Car `json:"-"`
}

func (car *RedCar_withHand) Id() any {
	return car.ID
}

func (RedCar_withHand) Create(db_ DB, carStr string) Model {
	redCar := &RedCar_withHand{}
	json.Unmarshal([]byte(carStr), redCar)

	carBct, _ := db_.Table("car", &Car{})
	if carBct != nil {
		carM := carBct.Manager().Get(redCar.CarId)
		if carM != nil {
			redCar.Car = carM.(*Car)
		}
	}
	return redCar
}

func (car *RedCar_withHand) Save(bct Table) error {
	return bct.Save(car)
}

func (car *RedCar_withHand) Delete(db_ DB) error {
	bct, _ := db_.Table("", &RedCar_withHand{})
	return bct.Delete(car.ID)
}

func showRedCar_withHand(cars []Model) {
	fmt.Printf(" %15v | %15v | %10v | %10v | %10v | %10v\n", "ID", "carId", "model", "color", "city", "hand")
	fmt.Printf(" --------------- | --------------- | ---------- | ---------- | ---------- | ----------\n")
	for _, carM := range cars {
		car, ok := carM.(*RedCar_withHand)
		if !ok {
			continue
		}
		if car.Car == nil {
			car.Car = &Car{}
		}
		fmt.Printf(" %15v | %15v | %10v | %10v | %10v | %10v\n", car.ID, car.CarId, car.Car.Model, car.Car.Color, car.Car.City, car.Hand)
	}
}

type Catalog struct {
	ID     uint
	CarsId []uint

	Cars []*Car `json:"-"`
}

func (cat *Catalog) Id() any {
	return cat.ID
}

func (Catalog) Create(db_ DB, str string) Model {
	cat := &Catalog{}
	json.Unmarshal([]byte(str), cat)

	carBct, _ := db_.Table("car", &Car{})
	if carBct != nil {
		for _, carId := range cat.CarsId {
			carM := carBct.Manager().Get(carId)
			if carM != nil {
				cat.Cars = append(cat.Cars, carM.(*Car))
			}
		}
	}
	return cat
}

func (cat *Catalog) Save(bct Table) error {
	return bct.Save(cat)
}

func (cat *Catalog) Delete(db_ DB) error {
	bct, _ := db_.Table("", &Catalog{})
	return bct.Delete(cat.ID)
}

func showCatalog(cats []Model) {
	fmt.Printf(" %15v | %15v | %15v\n", "ID", "carsId", "cars")
	fmt.Printf(" --------------- | --------------- | ---------------\n")
	for _, catM := range cats {
		cat, ok := catM.(*Catalog)
		if !ok {
			continue
		}
		ids := []uint{}
		for _, car := range cat.Cars {
			if car == nil {
				car = &Car{}
			}
			ids = append(ids, car.ID)
		}
		fmt.Printf(" %15v | %15v |\n%34v | %15v\n", cat.ID, fmt.Sprint(cat.CarsId), "", fmt.Sprint(ids))
	}
}

func Run() {
	// identity := ""
	// password := ""
	// db_, err := db.OpenPocketBaseLocal(identity, password)
	db_, err := db.OpenBbolt("sub_engine_fiber_db.db")
	if err != nil {
		panic(err)
	}
	defer db_.Close()
	fmt.Println("App start")
	fmt.Println("Please if you use pocketbase go to admin (login if required)")
	fmt.Println("Create tables and set custom rules for demo of interface to pb:")
	fmt.Println("\t`car` with fields (`model` string, `color` string, `city` string)")
	// fmt.Println("If you not want set custom rules, then set `identity` and `password` in demo/demo_db.go")

	// CreateModels(db_)

	carBct, err := db_.Table("", &Car{})
	if err != nil {
		panic(err)
	}
	/*catBct, err := db_.Table("", &Catalog{})
	if err != nil {
		panic(err)
	}

	car := &Car{
		Model: "BMW",
		Color: "pink2",
		City:  "Moscow",
	}
	if err := car.Save(carBct); err != nil {
		fmt.Println("car.Save:", err)
	}
	fmt.Println("car:", car)*/

	/*redCar := &RedCar_withHand{
		CarId: car.ID,
		Hand:  true,
	}
	if err := redCar.Save(redCarBct); err != nil {
		fmt.Println("redCar.Save:", err)
	}
	fmt.Println("redCar:", redCar)*/

	/*cat := &Catalog{
		CarsId: []uint{180, 181, 182, 183, car.ID},
	}
	if err := cat.Save(catBct); err != nil {
		fmt.Println("cat.Save:", err)
	}
	fmt.Println("cat:", cat)

	fmt.Println("car():", catBct.Manager().Get(cat.ID))*/

	// carBct.Delete(carBct.Objects.Count() - 3)
	// cars := carBct.Objects.Filter(db.Params{"Model": "Bug"}, db.Params{"Color": "black", "City": "Moscow"})
	cars := carBct.Manager().Filter(Params{"Color": "black"})
	showCar(cars.All())

	if err := carBct.Delete(uint(6)); err != nil {
		fmt.Println("car.Delete:", err)
	}
	cars = carBct.Manager().Filter(Params{"Color": "black"})
	showCar(cars.All())
	/*for {
	}*/
}
