package pocketbase

/*
Пакет pocketbase содержит функции и методы для работы с pocketbase
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	pocketbase "github.com/pocketbase/pocketbase"

	. "github.com/PoulIgorson/sub_engine_fiber/define"
	. "github.com/PoulIgorson/sub_engine_fiber/log"
)

// Pocketbase структура с данными авторизации для pb
type PocketBase struct {
	local             bool
	address           string
	identity          string
	password          string
	isAdmin           bool
	updateCollections bool
}

func (app *PocketBase) IsAdmin() bool {
	return app.isAdmin
}

func (app *PocketBase) Local() bool {
	return app.local
}

func (app *PocketBase) Address() string {
	return app.address
}

// New возвращает экземпляр *Pocketbase с адресом `address`, индификатором `identity` и паролем `password`
//
// address in format - http(s)://127.0.0.1(:8090)
func New(address, identity, password string, isAdmin bool, updateCollections ...bool) *PocketBase {
	update := false
	if len(updateCollections) > 0 {
		update = updateCollections[0]
	}
	return &PocketBase{false, address, identity, password, isAdmin, update}
}

func NewLocal(identity, password string, isAdmin bool, port ...string) *PocketBase {
	go pocketbase.New().Start()
	address := "http://127.0.0.1"
	if len(port) > 0 {
		if len(port[0]) > 0 && port[0] != ":" {
			address += ":"
		}
		address += port[0]
	} else {
		address += ":8090"
	}
	app := &PocketBase{true, address, identity, password, isAdmin, true}
	time.Sleep(5 * time.Second)
	return app
}

// Record структура записи в pb
type Record struct {
	collectionNameOrId string
	app                *PocketBase
	data               map[string]any
}

// Record.CollectionNameOrId возвращает имяИлиId колекции
func (record *Record) CollectionNameOrId() string {
	return record.collectionNameOrId
}

// Record.Get возвращает значение по ключу `key`
func (record *Record) Get(key string) any {
	return record.data[key]
}

// Record.Set устанавливает значение `value` по ключу `key`
func (record *Record) Set(key string, value any) {
	record.data[key] = value
}

// NewRecord возвращает экземпляр *Record
func NewRecord(collectionNameOrId string, app *PocketBase) *Record {
	return &Record{collectionNameOrId, app, nil}
}

// Form структура формы для создания или обновления записи
type Form struct {
	record *Record
	app    *PocketBase

	files map[string][]string
	data  map[string]any
}

// NewForm возвращает экземпляр *Form
func NewForm(app *PocketBase, record *Record) *Form {
	return &Form{record, app, nil, nil}
}

// Form.LoadData загружает в форму данные
func (form *Form) LoadData(data map[string]any) {
	if form.data == nil {
		form.data = map[string]any{}
	}
	for field, value := range data {
		if inner, ok := value.(map[string]any); ok {
			valueB, _ := json.Marshal(inner)
			value = string(valueB)
		}
		form.data[field] = value
	}
}

// Form.AddFiles добавляет пути файлов к форме
func (form *Form) AddFiles(field string, path ...string) {
	if form.files == nil {
		form.files = map[string][]string{}
	}
	form.files[field] = append(form.files[field], path...)
}

// Form.Submit записывает изменения в pb
func (form *Form) Submit() (string, error) {
	token, err := form.app.getToken()
	if err != nil {
		LogError.Printf("pocketbase.Submit.token.error: %v\n", err)
		return "", fmt.Errorf("pb.Form.Submit.token: %v", err.Error())
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for field, value := range form.data {
		if field == "id" {
			continue
		}
		part, _ := writer.CreateFormField(field)
		io.Copy(part, bytes.NewReader([]byte(fmt.Sprint(value))))
	}

	for field, paths := range form.files {
		for _, path := range paths {
			file, err := os.Open(path)
			if err != nil {
				LogError.Printf("pocketbase.Submit.path: %v: error: %v\n", path, err)
				continue
			}
			defer file.Close()

			part, _ := writer.CreateFormFile(field, path)
			io.Copy(part, file)
		}
	}
	writer.Close()

	var req *http.Request
	curl := fmt.Sprintf("%v/api/collections/%v/records", form.app.address, form.record.collectionNameOrId)
	if id, ok := form.data["id"]; !ok || id == "" || id == nil {
		req, _ = http.NewRequest("POST", curl, bytes.NewReader(body.Bytes()))
	} else {
		req, _ = http.NewRequest("PATCH", curl+"/"+fmt.Sprint(form.data["id"]), bytes.NewReader(body.Bytes()))
	}
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		LogError.Println("pocketbase.Submit.getResponse.error:", err)
		LogError.Println("pocketbase.Submit.getResponse.body:", response.Body)
		return "", err
	}
	responseBody, _ := io.ReadAll(response.Body)
	var resp struct {
		Id string `json:"id"`
	}
	json.Unmarshal(responseBody, &resp)
	if response.StatusCode != 200 && response.StatusCode != 204 {
		LogError.Println("pocketbase.Submit.status-resp:", response.StatusCode, resp)
		return "", fmt.Errorf("%v, %v", response.StatusCode, resp)
	}
	return resp.Id, nil
}

// PocketBase.getToken возвращает токен для работы с защищенным api
func (pb *PocketBase) getToken() (string, error) {
	if pb.identity == "" {
		return "", nil
	}
	data := map[string]any{
		"identity": pb.identity,
		"password": pb.password,
	}
	headers := map[string]string{
		"Content-Type": "application/json; charset=utf8",
	}
	collection := "collections/users"
	if pb.isAdmin {
		collection = "admins"
	}
	status, resp, err := GetJSONResponse(
		"POST", fmt.Sprintf("%v/api/%v/auth-with-password", pb.address, collection),
		Headers(headers), Data(data),
	)
	if err != nil {
		return "", err
	}
	if status != 200 && status != 204 {
		return "", fmt.Errorf("%v, %v, %v", status, resp, err)
	}
	return resp.(map[string]any)["token"].(string), nil
}

// PocketBase.Filter возвращает список записей из pb удовлетворяющим фильтру `data`
func (pb *PocketBase) Filter(collectionNameOrId string, data map[string]any, page ...uint) ([]*Record, error) {
	token, err := pb.getToken()
	if err != nil {
		LogError.Println("pocketbase.Filter.token.error:", err)
		return nil, fmt.Errorf("pb.Filter.token: %v", err.Error())
	}
	headers := map[string]string{
		"Content-Type":    "application/x-www-form-urlencoded",
		"Accept-Encoding": "identity",
		"Authorization":   token,
	}
	filter := ""
	for key, value := range data {
		if strings.IndexByte("=<>~", key[len(key)-1]) == -1 {
			key += "="
		}
		if len(filter) != 0 {
			filter += "&&"
		}
		if str, ok := value.(string); ok {
			value = `"` + str + `"`
		}
		filter += key + fmt.Sprint(value)
	}
	curl := fmt.Sprintf(`%v/api/collections/%v/records?perPage=500&filter=%v`, pb.address, collectionNameOrId, filter)
	if len(page) > 0 {
		curl += "&page=" + fmt.Sprint(page[0])
	}
	status, respI, err := GetJSONResponse(
		"GET", curl,
		Headers(headers), nil,
	)
	if err != nil {
		LogError.Println("pocketbase.Filter.getResponse.error:", err)
		return nil, err
	}
	if status == 204 {
		return []*Record{}, nil
	}
	if status != 200 {
		LogError.Println("pocketbase.Filter.getResponse:", status, respI)
		return nil, fmt.Errorf("%v, %v", status, respI)
	}

	resp := respI.(map[string]any)
	records := []*Record{}
	for _, item := range resp["items"].([]any) {
		records = append(records, &Record{collectionNameOrId, pb, item.(map[string]any)})
	}
	if int(resp["page"].(float64)) < int(resp["totalPages"].(float64)) {
		nextRecords, _ := pb.Filter(collectionNameOrId, data, uint(int(resp["page"].(float64))+1))
		records = append(records, nextRecords...)
	}
	return records, nil
}

func (pb *PocketBase) Delete(collectionNameOrId, id string) error {
	token, err := pb.getToken()
	if err != nil {
		LogError.Println("pocketbase.Filter.token.error:", err)
		return fmt.Errorf("pb.Delete.token: %v", err.Error())
	}
	headers := Headers{
		"Content-Type":    "application/x-www-form-urlencoded",
		"Accept-Encoding": "identity",
		"Authorization":   token,
	}
	curl := fmt.Sprintf(`%v/api/collections/%v/records/%v`, pb.address, collectionNameOrId, id)
	status, body, err := GetResponse(
		"DELETE", curl,
		headers, nil,
	)
	if err != nil {
		LogError.Println("pocketbase.Filter.getResponse.error:", err)
		return err
	}
	if status != 200 && status != 204 {
		LogError.Println("pocketbase.Filter.getResponse:", status, string(body))
		return fmt.Errorf("%v, %v", status, string(body))
	}
	return nil
}

// PocketBase.GetFileAsSliceByte возвращает список байтов файла из pb
func (pb *PocketBase) GetFileAsSliceByte(collentionNameOrId, recordId, fileName string) ([]byte, error) {
	token, err := pb.getToken()
	if err != nil {
		LogError.Println("pocketbase.GetFileAsSliceByte.token.error:", err)
		return nil, fmt.Errorf("pb.GetFileAsSliceByte.token: %v", err.Error())
	}
	curl := fmt.Sprintf("http://%v/api/files/%v/%v/%v", pb.address, collentionNameOrId, recordId, fileName)
	req, _ := http.NewRequest("GET", curl, nil)
	req.Header.Set("Authorization", token)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		LogError.Println("pocketbase.GetFileAsSliceByte.response.error:", err)
		return nil, err
	}
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != 204 {
		return []byte{}, nil
	}
	if response.StatusCode != 200 {
		LogError.Printf("pocketbase.GetFileAsSliceByte.response: status: %v, body: %v\n", response.StatusCode, fmt.Errorf("%v", body))
		return nil, fmt.Errorf("%v", body)
	}
	return body, nil
}

func (pb *PocketBase) doCollection(method, curl string, data map[string]any) error {
	token, err := pb.getToken()
	if err != nil {
		LogError.Println("pocketbase.doCollection.token:", err)
		return fmt.Errorf("pocketbase.doCollection.token: %v", err)
	}
	headers := Headers{
		"Content-Type":    "application/json",
		"Accept-Encoding": "identity",
		"Authorization":   token,
	}

	status, body, err := GetResponse(
		method, curl,
		headers, data,
	)
	if err != nil {
		LogError.Println("pocketbase.doCollection.getResponse.error:", err)
		return fmt.Errorf("getResponse.error: %v", err)
	}
	if status != 200 && status != 204 {
		LogError.Println("pocketbase.doCollection.getResponse:", status, string(body))
		return fmt.Errorf("getResponse: %v, %v", status, string(body))
	}
	return nil
}

func (pb *PocketBase) CreateCollection(data map[string]any) error {
	fmt.Println("create", pb)
	if !pb.updateCollections {
		return nil
	}
	return pb.doCollection("POST", fmt.Sprintf(`%v/api/collections`, pb.address), data)
}

func (pb *PocketBase) UpdateCollection(data map[string]any) error {
	fmt.Println("update", pb)
	if !pb.updateCollections {
		return nil
	}
	return pb.doCollection("PATCH", fmt.Sprintf(`%v/api/collections/%v`, pb.address, data["name"]), data)
}
