// Package define implements functions separate of project.
package define

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/constraints"
)

// Itoa convet int to string.
var Itoa = strconv.Itoa

// Atoi convet string to int.
var Atoi = func(x string) int { xInt, _ := strconv.Atoi(x); return xInt }

// itoa convet int to []byte.
var Itob = func(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// Min return minimal value in args.
func Min[T constraints.Ordered](args ...T) T {
	if len(args) == 0 {
		var zero T
		return zero
	}
	min := args[0]
	for _, arg := range args {
		if arg < min {
			min = arg
		}
	}
	return min
}

// Max return maximal value in args.
func Max[T constraints.Ordered](args ...T) T {
	if len(args) == 0 {
		var zero T
		return zero
	}
	max := args[0]
	for _, arg := range args {
		if arg > max {
			max = arg
		}
	}
	return max
}

// GetToday returns string today date.
func GetToday() string {
	ctime := time.Now()
	cday, cmonth := ctime.Day(), int(ctime.Month())
	day, month := Itoa(cday), Itoa(cmonth)
	if len(day) < 2 {
		day = "0" + day
	}
	if len(month) < 2 {
		month = "0" + month
	}
	return fmt.Sprintf("%v-%v-%v", day, month, ctime.Year())
}

// Dict convet interface{} to fiber.Map.
func Dict(dict interface{}) fiber.Map {
	return dict.(map[string]interface{})
}

// ErrorToStr convert []error to string.
func ErrorsToStr(errs []error) string {
	errors := ""
	for _, err := range errs {
		errors += err.Error() + ", "
	}
	return errors[:len(errors)-2]
}

// Contains returns
func Contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func Pow[T constraints.Integer](a T, b int) float64 {
	res := float64(1)
	var aa float64
	if b < 0 {
		aa = float64(1) / float64(a)
	} else {
		aa = float64(a)
	}
	for i := 0; i < Abs(b); i++ {
		res *= aa
	}
	return res
}

func Insert[T constraints.Ordered](a []T, index int, value T) []T {
	if index < 0 {
		index = 0
	}
	if len(a) <= index {
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...)
	a[index] = value
	return a
}

// Hash returns hash data
func Hash(data []byte) string {
	hasher := sha1.New()
	hasher.Write(data)
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

type Set[T constraints.Ordered] struct {
	items []T
}

func (this *Set[T]) GetItems() []T {
	var tItems []T
	for _, x := range this.items {
		tItems = append(tItems, x)
	}
	return tItems
}

func (this *Set[T]) Add(x T) {
	if !Contains(this.items, x) {
		this.items = append(this.items, x)
	}
}

func (this *Set[T]) Adds(lst ...T) {
	for _, x := range lst {
		this.Add(x)
	}
}

func (this *Set[T]) Get(index int) T {
	var t T
	if index < 0 {
		index += len(this.items)
	}
	if index > len(this.items) || index < 0 {
		return t
	}
	return this.items[index]
}
