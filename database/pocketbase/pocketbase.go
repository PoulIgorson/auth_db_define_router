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

	pocketbase "github.com/pocketbase/pocketbase"

	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
	. "github.com/PoulIgorson/sub_engine_fiber/define"
)

// Pocketbase структура с данными авторизации для pb
type PocketBase struct {
	local    bool
	address  string
	identity string
	password string
}

// New возвращает экземпляр *Pocketbase с адресом `address`, индификатором `identity` и паролем `password`
//
// address in format - http(s)://127.0.0.1(:8090)
func New(address, identity, password string) *PocketBase {
	return &PocketBase{false, address, identity, password}
}

func NewLocal(identity, password string, port ...string) *PocketBase {
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
	return &PocketBase{true, address, identity, password}
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
		if len(field) > 0 && !strings.Contains("=<>~", string(field[len(field)-1])) {
			field += "="
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
func (form *Form) Submit() error {
	token, err := form.app.getToken()
	if err != nil {
		fmt.Printf("pocketbase.Submit.token.error: %v\n", err)
		return err
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
				fmt.Printf("pocketbase.Submit.path: %v: error: %v\n", path, err)
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
	if id, ok := form.data["id"]; !ok || id == "" {
		req, _ = http.NewRequest("POST", curl, bytes.NewReader(body.Bytes()))
	} else {
		req, _ = http.NewRequest("PATCH", curl+"/"+fmt.Sprint(form.data["id"]), bytes.NewReader(body.Bytes()))
	}
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("pocketbase.Submit.getResponse.error:", err)
		fmt.Println("pocketbase.Submit.getResponse.body:", response.Body)
		return err
	}
	responseBody, _ := io.ReadAll(response.Body)
	var resp any
	json.Unmarshal(responseBody, &resp)
	if response.StatusCode != 200 {
		fmt.Println("pocketbase.Submit.status-resp:", response.StatusCode, resp)
		return fmt.Errorf("%v, %v", response.StatusCode, resp)
	}
	return nil
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
	status, resp, err := GetJSONResponse(
		"POST", fmt.Sprintf("%v/api/collections/users/auth-with-password", pb.address),
		Headers(headers), Data(data),
	)
	if status != 200 {
		return "", fmt.Errorf("%v, %v, %v", status, resp, err)
	}
	if err != nil {
		return "", err
	}
	return resp.(map[string]any)["token"].(string), nil
}

// PocketBase.Filter возвращает список записей из pb удовлетворяющим фильтру `data`
func (pb *PocketBase) Filter(collectionNameOrId string, data map[string]any, page ...uint) ([]*Record, error) {
	token, err := pb.getToken()
	if err != nil {
		fmt.Println("pocketbase.Filter.token.error:", err)
		return nil, err
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
		fmt.Println("pocketbase.Filter.getResponse.error:", err)
		return nil, err
	}
	if status != 200 {
		fmt.Println("pocketbase.Filter.getResponse:", status, respI)
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
		fmt.Println("pocketbase.Filter.token.error:", err)
		return err
	}
	headers := Headers{
		"Content-Type":    "application/x-www-form-urlencoded",
		"Accept-Encoding": "identity",
		"Authorization":   token,
	}
	curl := fmt.Sprintf(`%v/api/collections/%v/records/%v`, pb.address, collectionNameOrId, id)
	status, respI, err := GetResponse(
		"DELETE", curl,
		headers, nil,
	)
	if err != nil {
		fmt.Println("pocketbase.Filter.getResponse.error:", err)
		return err
	}
	if status != 200 {
		fmt.Println("pocketbase.Filter.getResponse:", status, respI)
		return fmt.Errorf("%v, %v", status, respI)
	}
	return nil
}

// PocketBase.GetFileAsSliceByte возвращает список байтов файла из pb
//
// По id записи `recordId` в колекции `collentionNameOrId` и имени файла `fileName`
func (pb *PocketBase) GetFileAsSliceByte(collentionNameOrId, recordId, fileName string) ([]byte, error) {
	token, err := pb.getToken()
	if err != nil {
		fmt.Println("pocketbase.GetFileAsSliceByte.token.error:", err)
		return nil, err
	}
	curl := fmt.Sprintf("http://%v/api/files/%v/%v/%v", pb.address, collentionNameOrId, recordId, fileName)
	req, _ := http.NewRequest("GET", curl, nil)
	req.Header.Set("Authorization", token)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("pocketbase.GetFileAsSliceByte.response.error:", err)
		return nil, err
	}
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != 200 {
		fmt.Printf("pocketbase.GetFileAsSliceByte.response: status: %v, body: %v\n", response.StatusCode, fmt.Errorf("%v", body))
		return nil, fmt.Errorf("%v", body)
	}
	return body, nil
}

// -------------------

// DB
type DataBase struct {
	pb      *PocketBase
	buckets map[string]*Bucket
}

func Open(address, identity, password string) *DataBase {
	return &DataBase{
		pb:      New(address, identity, password),
		buckets: map[string]*Bucket{},
	}
}

func OpenWith(pb *PocketBase, buckets map[string]*Bucket) *DataBase {
	return &DataBase{
		pb:      pb,
		buckets: buckets,
	}
}

func (db *DataBase) Close() Error {
	return nil
}

func (db *DataBase) Table(name string, model Model) (Table, Error) {
	if db.buckets[name] != nil {
		return db.buckets[name], nil
	}
	if !db.ExistsTable(name) {
		return nil, NewErrorf("pb: table not exists")
	}
	if _, ok := model.Id().(string); !ok && name != "user" {
		return nil, NewErrorf("pb: id must be string")
	}
	bucket := &Bucket{
		db:    db,
		name:  name,
		model: model,
	}
	bucket.Objects = &Manager{
		bucket:  bucket,
		objects: map[string]Model{},
	}
	db.buckets[name] = bucket
	return bucket, nil
}

func (db *DataBase) ExistsTable(name string) bool {
	_, err := db.pb.Filter(name, map[string]any{})
	return err == nil
}

// Bucket
type Bucket struct {
	db   *DataBase
	name string

	model   Model
	Objects *Manager
}

func (bucket *Bucket) Manager() ManagerI {
	return bucket.Objects
}

func (bucket *Bucket) DB() DB {
	return bucket.db
}

func (bucket *Bucket) Name() string {
	return bucket.name
}

func (bucket *Bucket) Model() Model {
	return bucket.model
}

func (bucket *Bucket) Count() uint {
	records, err := bucket.db.pb.Filter(bucket.name, map[string]any{})
	if err != nil {
		return 0
	}
	return uint(len(records))
}

func (bucket *Bucket) Get(id any) (Model, Error) {
	records, err := bucket.db.pb.Filter(bucket.name, map[string]any{"id": id})
	if err != nil {
		return nil, ToError(err)
	}
	if len(records) == 0 {
		return nil, NewErrorf("pb: record not found")
	}
	dataByte, _ := json.Marshal(records[0].data)
	model := bucket.model.Create(bucket.db, string(dataByte))
	for bucket.Objects.rwObjects {
	}
	bucket.Objects.rwObjects = true
	bucket.Objects.objects[model.Id().(string)] = model
	bucket.Objects.rwObjects = false
	bucket.Objects.count++
	return model, nil
}

func (bucket *Bucket) Save(model Model) Error {
	dataByte, _ := json.Marshal(model)
	data := map[string]any{}
	json.Unmarshal(dataByte, &data)
	form := NewForm(bucket.db.pb, NewRecord(bucket.name, bucket.db.pb))
	form.LoadData(data)
	return ToError(form.Submit())
}

func (bucket *Bucket) Delete(idI any) Error {
	id, ok := idI.(string)
	if !ok {
		return NewErrorf("pb: id must be string")
	}
	return ToError(bucket.db.pb.Delete(bucket.name, id))
}

func (bucket *Bucket) DeleteAll() Error {
	return NewErrorf("pocketbase does not support DeleteAll")
}

// Manager
type Manager struct {
	isInstance bool
	bucket     *Bucket

	count     uint
	objects   map[string]Model
	rwObjects bool
	maxId     string
	minId     string
}

func (manager *Manager) IsInstance() bool {
	return manager.isInstance
}

func (manager *Manager) Bucket() *Bucket {
	return manager.bucket
}

func (manager *Manager) Table() Table {
	return manager.bucket
}

func (manager *Manager) Copy() ManagerI {
	return &Manager{
		isInstance: true,
		bucket:     manager.bucket,
		objects:    map[string]Model(manager.objects),
		count:      manager.count,
		maxId:      manager.maxId,
		minId:      manager.minId,
	}
}

func (manager *Manager) Get(idI any) Model {
	id, ok := idI.(string)
	if !ok {
		return nil
	}
	for manager.rwObjects {
	}
	manager.rwObjects = true
	if m := manager.objects[id]; m != nil {
		manager.rwObjects = false
		return m
	}
	manager.rwObjects = false
	model, _ := manager.bucket.Get(id)
	if model != nil {
		for manager.rwObjects {
		}
		manager.rwObjects = true
		manager.objects[id] = model
		manager.rwObjects = false
		manager.count++
		if manager.maxId < id {
			manager.maxId = id
		}
		if manager.minId > id || manager.minId == "" {
			manager.minId = id
		}
	}
	return model
}

func (manager *Manager) Delete(id any) {
	manager.bucket.Delete(id)
}

func recordToModel(record *Record, db DB, model Model) Model {
	dataByte, _ := json.Marshal(record.data)
	return model.Create(db, string(dataByte))
}

func nameFieldsToJSONTags(model Model, params Params) Params {
	tagParams := Params{}
	for nameField, value := range params {
		tag := GetTagField(model, nameField)
		if len(tag) > 5 {
			tag = tag[6 : len(tag)-1]
		}
		fmt.Println(nameField, tag)
		tagParams[tag] = value
	}
	return tagParams
}

func (manager *Manager) Filter(include Params, _ ...Params) ManagerI {
	objects := map[string]Model{}
	records, _ := manager.bucket.db.pb.Filter(manager.bucket.name, nameFieldsToJSONTags(manager.bucket.model, include))
	for _, record := range records {
		model := recordToModel(record, manager.bucket.db, manager.bucket.model)
		for manager.rwObjects {
		}
		manager.rwObjects = true
		if manager.objects[model.Id().(string)] == nil {
			manager.count++
		}
		manager.objects[model.Id().(string)] = model
		manager.rwObjects = false
		if manager.maxId < model.Id().(string) {
			manager.maxId = model.Id().(string)
		}
		if manager.minId > model.Id().(string) || manager.minId == "" {
			manager.minId = model.Id().(string)
		}
		objects[model.Id().(string)] = model
	}
	return &Manager{
		isInstance: true,
		bucket:     manager.bucket,
		objects:    objects,
		maxId:      manager.maxId,
		minId:      manager.minId,
	}
}

func (manager *Manager) All() []Model {
	objects := []Model{}
	be := false
	for id := range manager.objects {
		for manager.rwObjects {
		}
		manager.rwObjects = true
		model := manager.objects[id]
		manager.rwObjects = false
		be = true
		objects = append(objects, model)
	}
	if !be && !manager.isInstance {
		records, _ := manager.bucket.db.pb.Filter(manager.bucket.name, map[string]any{})
		for _, record := range records {
			model := recordToModel(record, manager.bucket.db, manager.bucket.model)
			for manager.rwObjects {
			}
			manager.rwObjects = true
			if manager.objects[model.Id().(string)] == nil {
				manager.count++
			}
			manager.objects[model.Id().(string)] = model
			manager.rwObjects = false
			if manager.maxId < model.Id().(string) {
				manager.maxId = model.Id().(string)
			}
			if manager.minId > model.Id().(string) || manager.minId == "" {
				manager.minId = model.Id().(string)
			}
			objects = append(objects, model)
		}
	}
	return objects
}

func (manager *Manager) First() Model {
	for manager.rwObjects {
	}
	manager.rwObjects = true
	model := manager.objects[manager.minId]
	manager.rwObjects = false
	return model
}

func (manager *Manager) Last() Model {
	for manager.rwObjects {
	}
	manager.rwObjects = true
	model := manager.objects[manager.maxId]
	manager.rwObjects = false
	return model
}

func (manager *Manager) Count() uint {
	if manager.count == 0 {
		return uint(len(manager.All()))
	}
	return manager.count
}
