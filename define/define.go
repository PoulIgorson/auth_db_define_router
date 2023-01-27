// Package define implements functions separate of project.
package define

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/constraints"
)

// Itoa convet int to string.
var Itoa = strconv.Itoa

var ParseUint = func(x string) uint { xUint, _ := strconv.ParseUint(x, 10, 0); return uint(xUint) }

// Atoi convet string to int.
var Atoi = func(x string) int { xInt, _ := strconv.Atoi(x); return xInt }

// Itob convet int to []byte.
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

func (this *Set[T]) Count() int {
	return len(this.items)
}

func GetEncodeFunc(format string) func(io.Writer, image.Image) error {
	switch format {
	case "image/png":
		return png.Encode
	case "png":
		return png.Encode
	case "image/jpeg":
		return func(w io.Writer, i image.Image) error {
			return jpeg.Encode(w, i, nil)
		}
	case "jpeg":
		return func(w io.Writer, i image.Image) error {
			return jpeg.Encode(w, i, nil)
		}
	case "image/jpg":
		return func(w io.Writer, i image.Image) error {
			return jpeg.Encode(w, i, nil)
		}
	case "jpg":
		return func(w io.Writer, i image.Image) error {
			return jpeg.Encode(w, i, nil)
		}
	}
	return nil
}

func GetDecodeFunc(format string) func(io.Reader) (image.Image, error) {
	switch format {
	case "image/png":
		return png.Decode
	case "png":
		return png.Decode
	case "image/jpeg":
		return jpeg.Decode
	case "jpeg":
		return jpeg.Decode
	case "image/jpg":
		return jpeg.Decode
	case "jpg":
		return jpeg.Decode
	}
	return nil
}

func GetImagesFromRequestBody(body []byte) ([]image.Image, []string) {
	var data map[string]interface{}
	json.Unmarshal(body, &data)
	imagesData := data["images"].([]interface{})
	var images []image.Image
	var formats []string
	for i := 0; i < len(imagesData); i++ {
		imgData := imagesData[i].(string)
		coI := strings.Index(imgData, ",")
		rawImage := string(imgData)[coI+1:]
		unbased, _ := base64.StdEncoding.DecodeString(string(rawImage))
		res := bytes.NewReader(unbased)
		format := imgData[strings.Index(imgData, ":")+1 : strings.Index(imgData, ";")]
		img, err := GetDecodeFunc(format)(res)
		if err != nil {
			fmt.Println("GetImagesFromRequestBody: decode:", err.Error())
			continue
		}
		images = append(images, img)
		formats = append(formats, format)
	}
	return images, formats
}

func ImagesToBytes(images []image.Image, formats []string) [][]byte {
	var res [][]byte
	for i := 0; i < len(images); i++ {
		img := images[i]
		buf := new(bytes.Buffer)
		err := GetEncodeFunc("png")(buf, img)
		if err == nil {
			res = append(res, buf.Bytes())
		}
	}
	return res
}

func IndexOf[T comparable](lst []T, x T) int {
	for i := 0; i < len(lst); i++ {
		if lst[i] == x {
			return i
		}
	}
	return -1
}

func Check(imodel interface{}, field_name string) (*reflect.Value, error) {
	vPointerTomodel := reflect.ValueOf(imodel)
	if vPointerTomodel.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("Check: getting value is not pointer")
	}
	vModel := vPointerTomodel.Elem()
	if vModel.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Check: getting value is not struct")
	}
	withins := strings.Split(field_name, ".")
	vfield := vModel
	i := 0
	for i, field_name = range withins {
		if vfield.Kind() != reflect.Struct {
			return nil, fmt.Errorf("Check: %v is not struct, is %v", withins[i-1], vfield.Kind())
		}
		cvfield := vfield.FieldByName(field_name)
		if cvfield.Kind() == reflect.Invalid {
			vfield = reflect.Value{}
			break
		}
		vfield = cvfield
	}
	if vfield.Kind() == reflect.Invalid {
		return nil, fmt.Errorf("Check: Field `%v` does not exists", field_name)
	}
	return &vfield, nil
}

func ChangeFieldOfName(imodel interface{}, field_name string, value interface{}) error {
	vfield, err := Check(imodel, field_name)
	if err != nil {
		return err
	}
	if vfield.Kind() == reflect.Invalid {
		return fmt.Errorf("ChangeFieldOfName: Field `%v` does not exists", field_name)
	}
	vvalue := reflect.ValueOf(value)
	if !vvalue.CanConvert(vfield.Type()) {
		return fmt.Errorf("ChangeFieldOfName: `%v` (type %T) not be converted to type %v", value, value, vfield.Type())
	}
	vvalue = vvalue.Convert(vfield.Type())
	vfield.Set(vvalue)
	return nil
}

func CopyMap[T1, T2 comparable](Map map[T1]T2) map[T1]T2 {
	newMap := map[T1]T2{}
	for k, v := range Map {
		newMap[k] = v
	}
	return newMap
}

func CopyMapAny[T1, T2 comparable](Map map[T1]T2) map[string]any {
	newMap := map[string]any{}
	for k, v := range Map {
		newMap[fmt.Sprint(k)] = v
	}
	return newMap
}
