// Copyright (c) 2018, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/globalsign/mgo/bson"
	jsonresp "github.com/sylabs/json-resp"
)

const testToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.TCYt5XsITJX1CxPCT8yAV-TVkIEq_PbChOMqsLfRoPsnsgw5WEuts01mq-pQy7UJiN5mgRxD-WUcX16dUEMGlv50aqzpqh4Qktb3rk-BuQy72IFLOqV0G_zS245-kronKb78cPN25DGlcTwLtjPAYuNzVBAh4vGHSrQyHUdBBPM"

var (
	testEntity = Entity{
		ID:          bson.NewObjectId(),
		Name:        "test-user",
		Description: "A test user",
	}

	testCollection = Collection{
		ID:          bson.NewObjectId(),
		Name:        "test-collection",
		Description: "A test collection",
		Entity:      testEntity.ID,
		EntityName:  testEntity.Name,
	}

	testContainer = Container{
		ID:             bson.NewObjectId(),
		Name:           "test-container",
		Description:    "A test container",
		Entity:         testEntity.ID,
		EntityName:     testEntity.Name,
		Collection:     testEntity.ID,
		CollectionName: testCollection.Name,
		ImageTags: map[string]bson.ObjectId{
			"test-tag": bson.NewObjectId(),
			"latest":   bson.NewObjectId()},
	}

	testImage = Image{
		ID:             bson.NewObjectId(),
		Hash:           "sha256.e50a30881ace3d5944f5661d222db7bee5296be9e4dc7c1fcb7604bcae926e88",
		Entity:         testEntity.ID,
		EntityName:     testEntity.Name,
		Collection:     testEntity.ID,
		CollectionName: testCollection.Name,
		Container:      testContainer.ID,
		ContainerName:  testContainer.Name,
	}

	testSearch = SearchResults{
		Entities:    []Entity{testEntity},
		Collections: []Collection{testCollection},
		Containers:  []Container{testContainer},
	}
)

type mockService struct {
	t           *testing.T
	code        int
	body        interface{}
	reqCallback func(*http.Request, *testing.T)
	httpAddr    string
	httpPath    string
	httpServer  *httptest.Server
	baseURI     string
}

func (m *mockService) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc(m.httpPath, m.ServeHTTP)
	m.httpServer = httptest.NewServer(mux)
	m.httpAddr = m.httpServer.Listener.Addr().String()
	m.baseURI = "http://" + m.httpAddr
}

func (m *mockService) Stop() {
	m.httpServer.Close()
}

func (m *mockService) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if m.reqCallback != nil {
		m.reqCallback(r, m.t)
	}
	w.WriteHeader(m.code)
	err := json.NewEncoder(w).Encode(&m.body)
	if err != nil {
		m.t.Errorf("Error encoding mock response: %v", err)
	}
}

func Test_GetEntity(t *testing.T) {

	tests := []struct {
		description  string
		code         int
		body         interface{}
		reqCallback  func(*http.Request, *testing.T)
		entityRef    string
		expectEntity *Entity
		expectFound  bool
		expectError  bool
	}{
		{
			description:  "NotFound",
			code:         http.StatusNotFound,
			body:         jsonresp.Response{Error: &jsonresp.Error{Code: http.StatusNotFound}},
			reqCallback:  nil,
			entityRef:    "notthere",
			expectEntity: nil,
			expectFound:  false,
			expectError:  false,
		},
		{
			description:  "Unauthorized",
			code:         http.StatusUnauthorized,
			body:         jsonresp.Response{Error: &jsonresp.Error{Code: http.StatusUnauthorized}},
			reqCallback:  nil,
			entityRef:    "notmine",
			expectEntity: nil,
			expectFound:  false,
			expectError:  true,
		},
		{
			description:  "ValidResponse",
			code:         http.StatusOK,
			body:         EntityResponse{Data: testEntity},
			reqCallback:  nil,
			entityRef:    "test-user",
			expectEntity: &testEntity,
			expectFound:  true,
			expectError:  false,
		},
	}

	// Loop over test cases
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {

			m := mockService{
				t:           t,
				code:        tt.code,
				body:        tt.body,
				reqCallback: tt.reqCallback,
				httpPath:    "/v1/entities/" + tt.entityRef,
			}

			m.Run()
			defer m.httpServer.Close()

			c, err := NewClient(&Config{AuthToken: testToken, BaseURL: m.baseURI})
			if err != nil {
				t.Errorf("Error initializing client: %v", err)
			}

			entity, found, err := GetEntity(c, tt.entityRef)

			if err != nil && !tt.expectError {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && tt.expectError {
				t.Errorf("Unexpected success. Expected error.")
			}
			if found != tt.expectFound {
				t.Errorf("Got found %v - expected %v", found, tt.expectFound)
			}
			if !reflect.DeepEqual(entity, tt.expectEntity) {
				t.Errorf("Got entity %v - expected %v", entity, tt.expectEntity)
			}

			m.Stop()

		})

	}
}

func Test_GetCollection(t *testing.T) {

	tests := []struct {
		description      string
		code             int
		body             interface{}
		reqCallback      func(*http.Request, *testing.T)
		collectionRef    string
		expectCollection *Collection
		expectFound      bool
		expectError      bool
	}{
		{
			description:      "NotFound",
			code:             http.StatusNotFound,
			body:             jsonresp.Response{Error: &jsonresp.Error{Code: http.StatusNotFound}},
			reqCallback:      nil,
			collectionRef:    "notthere",
			expectCollection: nil,
			expectFound:      false,
			expectError:      false,
		},
		{
			description:      "Unauthorized",
			code:             http.StatusUnauthorized,
			body:             jsonresp.Response{Error: &jsonresp.Error{Code: http.StatusUnauthorized}},
			reqCallback:      nil,
			collectionRef:    "notmine",
			expectCollection: nil,
			expectFound:      false,
			expectError:      true,
		},
		{
			description:      "ValidResponse",
			code:             http.StatusOK,
			body:             CollectionResponse{Data: testCollection},
			reqCallback:      nil,
			collectionRef:    "test-entity/test-collection",
			expectCollection: &testCollection,
			expectFound:      true,
			expectError:      false,
		},
	}

	// Loop over test cases
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {

			m := mockService{
				t:           t,
				code:        tt.code,
				body:        tt.body,
				reqCallback: tt.reqCallback,
				httpPath:    "/v1/collections/" + tt.collectionRef,
			}

			m.Run()
			defer m.httpServer.Close()

			c, err := NewClient(&Config{AuthToken: testToken, BaseURL: m.baseURI})
			if err != nil {
				t.Errorf("Error initializing client: %v", err)
			}

			collection, found, err := GetCollection(c, tt.collectionRef)

			if err != nil && !tt.expectError {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && tt.expectError {
				t.Errorf("Unexpected success. Expected error.")
			}
			if found != tt.expectFound {
				t.Errorf("Got found %v - expected %v", found, tt.expectFound)
			}
			if !reflect.DeepEqual(collection, tt.expectCollection) {
				t.Errorf("Got entity %v - expected %v", collection, tt.expectCollection)
			}

			m.Stop()

		})

	}
}

func Test_GetContainer(t *testing.T) {

	tests := []struct {
		description     string
		code            int
		body            interface{}
		reqCallback     func(*http.Request, *testing.T)
		containerRef    string
		expectContainer *Container
		expectFound     bool
		expectError     bool
	}{
		{
			description:     "NotFound",
			code:            http.StatusNotFound,
			body:            jsonresp.Response{Error: &jsonresp.Error{Code: http.StatusNotFound}},
			reqCallback:     nil,
			containerRef:    "notthere",
			expectContainer: nil,
			expectFound:     false,
			expectError:     false,
		},
		{
			description:     "Unauthorized",
			code:            http.StatusUnauthorized,
			body:            jsonresp.Response{Error: &jsonresp.Error{Code: http.StatusUnauthorized}},
			reqCallback:     nil,
			containerRef:    "notmine",
			expectContainer: nil,
			expectFound:     false,
			expectError:     true,
		},
		{
			description:     "ValidResponse",
			code:            http.StatusOK,
			body:            ContainerResponse{Data: testContainer},
			reqCallback:     nil,
			containerRef:    "test-entity/test-collection/test-container",
			expectContainer: &testContainer,
			expectFound:     true,
			expectError:     false,
		},
	}

	// Loop over test cases
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {

			m := mockService{
				t:           t,
				code:        tt.code,
				body:        tt.body,
				reqCallback: tt.reqCallback,
				httpPath:    "/v1/containers/" + tt.containerRef,
			}

			m.Run()
			defer m.httpServer.Close()

			c, err := NewClient(&Config{AuthToken: testToken, BaseURL: m.baseURI})
			if err != nil {
				t.Errorf("Error initializing client: %v", err)
			}

			container, found, err := GetContainer(c, tt.containerRef)

			if err != nil && !tt.expectError {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && tt.expectError {
				t.Errorf("Unexpected success. Expected error.")
			}
			if found != tt.expectFound {
				t.Errorf("Got found %v - expected %v", found, tt.expectFound)
			}
			if !reflect.DeepEqual(container, tt.expectContainer) {
				t.Errorf("Got container %v - expected %v", container, tt.expectContainer)
			}

			m.Stop()

		})

	}
}

func Test_GetImage(t *testing.T) {

	tests := []struct {
		description string
		code        int
		body        interface{}
		reqCallback func(*http.Request, *testing.T)
		imageRef    string
		expectImage *Image
		expectFound bool
		expectError bool
	}{
		{
			description: "NotFound",
			code:        http.StatusNotFound,
			body:        jsonresp.Response{Error: &jsonresp.Error{Code: http.StatusNotFound}},
			reqCallback: nil,
			imageRef:    "notthere",
			expectImage: nil,
			expectFound: false,
			expectError: false,
		},
		{
			description: "Unauthorized",
			code:        http.StatusUnauthorized,
			body:        jsonresp.Response{Error: &jsonresp.Error{Code: http.StatusUnauthorized}},
			reqCallback: nil,
			imageRef:    "notmine",
			expectImage: nil,
			expectFound: false,
			expectError: true,
		},
		{
			description: "ValidResponse",
			code:        http.StatusOK,
			body:        ImageResponse{Data: testImage},
			reqCallback: nil,
			imageRef:    "test",
			expectImage: &testImage,
			expectFound: true,
			expectError: false,
		},
	}

	// Loop over test cases
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {

			m := mockService{
				t:           t,
				code:        tt.code,
				body:        tt.body,
				reqCallback: tt.reqCallback,
				httpPath:    "/v1/images/" + tt.imageRef,
			}

			m.Run()
			defer m.Stop()

			c, err := NewClient(&Config{AuthToken: testToken, BaseURL: m.baseURI})
			if err != nil {
				t.Errorf("Error initializing client: %v", err)
			}

			image, found, err := GetImage(c, tt.imageRef)

			if err != nil && !tt.expectError {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && tt.expectError {
				t.Errorf("Unexpected success. Expected error.")
			}
			if found != tt.expectFound {
				t.Errorf("Got found %v - expected %v", found, tt.expectFound)
			}
			if !reflect.DeepEqual(image, tt.expectImage) {
				t.Errorf("Got image %v - expected %v", image, tt.expectImage)
			}
		})
	}
}

func Test_CreateEntity(t *testing.T) {

	tests := []struct {
		description  string
		code         int
		body         interface{}
		reqCallback  func(*http.Request, *testing.T)
		entityRef    string
		expectEntity *Entity
		expectError  bool
	}{
		{
			description:  "Valid Request",
			code:         http.StatusOK,
			body:         EntityResponse{Data: testEntity},
			entityRef:    "test",
			expectEntity: &testEntity,
			expectError:  false,
		},
		{
			description:  "Error response",
			code:         http.StatusInternalServerError,
			body:         Entity{},
			entityRef:    "test",
			expectEntity: nil,
			expectError:  true,
		},
	}

	// Loop over test cases
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {

			m := mockService{
				t:           t,
				code:        tt.code,
				body:        tt.body,
				reqCallback: tt.reqCallback,
				httpPath:    "/v1/entities/",
			}

			m.Run()
			defer m.httpServer.Close()

			c, err := NewClient(&Config{AuthToken: testToken, BaseURL: m.baseURI})
			if err != nil {
				t.Errorf("Error initializing client: %v", err)
			}

			entity, err := CreateEntity(c, tt.entityRef)

			if err != nil && !tt.expectError {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && tt.expectError {
				t.Errorf("Unexpected success. Expected error.")
			}
			if !reflect.DeepEqual(entity, tt.expectEntity) {
				t.Errorf("Got created entity %v - expected %v", entity, tt.expectEntity)
			}

			m.Stop()

		})

	}
}

func Test_CreateCollection(t *testing.T) {

	tests := []struct {
		description      string
		code             int
		body             interface{}
		reqCallback      func(*http.Request, *testing.T)
		collectionRef    string
		expectCollection *Collection
		expectError      bool
	}{
		{
			description:      "Valid Request",
			code:             http.StatusOK,
			body:             CollectionResponse{Data: Collection{Name: "test"}},
			collectionRef:    "test",
			expectCollection: &Collection{Name: "test"},
			expectError:      false,
		},
	}

	// Loop over test cases
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {

			m := mockService{
				t:           t,
				code:        tt.code,
				body:        tt.body,
				reqCallback: tt.reqCallback,
				httpPath:    "/v1/collections/",
			}

			m.Run()

			c, err := NewClient(&Config{AuthToken: testToken, BaseURL: m.baseURI})
			if err != nil {
				t.Errorf("Error initializing client: %v", err)
			}

			collection, err := CreateCollection(c, tt.collectionRef, bson.NewObjectId().Hex())

			if err != nil && !tt.expectError {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && tt.expectError {
				t.Errorf("Unexpected success. Expected error.")
			}
			if !reflect.DeepEqual(collection, tt.expectCollection) {
				t.Errorf("Got created collection %v - expected %v", collection, tt.expectCollection)
			}

			m.Stop()

		})

	}
}

func Test_CreateContainer(t *testing.T) {

	tests := []struct {
		description     string
		code            int
		body            interface{}
		reqCallback     func(*http.Request, *testing.T)
		containerRef    string
		expectContainer *Container
		expectError     bool
	}{
		{
			description:     "Valid Request",
			code:            http.StatusOK,
			body:            ContainerResponse{Data: Container{Name: "test"}},
			containerRef:    "test",
			expectContainer: &Container{Name: "test"},
			expectError:     false,
		},
	}

	// Loop over test cases
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {

			m := mockService{
				t:           t,
				code:        tt.code,
				body:        tt.body,
				reqCallback: tt.reqCallback,
				httpPath:    "/v1/containers/",
			}

			m.Run()
			defer m.httpServer.Close()

			c, err := NewClient(&Config{AuthToken: testToken, BaseURL: m.baseURI})
			if err != nil {
				t.Errorf("Error initializing client: %v", err)
			}

			container, err := CreateContainer(c, tt.containerRef, bson.NewObjectId().Hex())

			if err != nil && !tt.expectError {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && tt.expectError {
				t.Errorf("Unexpected success. Expected error.")
			}
			if !reflect.DeepEqual(container, tt.expectContainer) {
				t.Errorf("Got created collection %v - expected %v", container, tt.expectContainer)
			}

			m.Stop()

		})

	}
}

func Test_CreateImage(t *testing.T) {

	tests := []struct {
		description string
		code        int
		body        interface{}
		reqCallback func(*http.Request, *testing.T)
		imageRef    string
		expectImage *Image
		expectError bool
	}{
		{
			description: "Valid Request",
			code:        http.StatusOK,
			body:        ImageResponse{Data: Image{Hash: "sha256.e50a30881ace3d5944f5661d222db7bee5296be9e4dc7c1fcb7604bcae926e88"}},
			imageRef:    "test",
			expectImage: &Image{Hash: "sha256.e50a30881ace3d5944f5661d222db7bee5296be9e4dc7c1fcb7604bcae926e88"},
			expectError: false,
		},
		{
			description: "Error response",
			code:        http.StatusInternalServerError,
			body:        Image{},
			imageRef:    "test",
			expectImage: nil,
			expectError: true,
		},
	}

	// Loop over test cases
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {

			m := mockService{
				t:           t,
				code:        tt.code,
				body:        tt.body,
				reqCallback: tt.reqCallback,
				httpPath:    "/v1/images/",
			}

			m.Run()
			defer m.httpServer.Close()

			c, err := NewClient(&Config{AuthToken: testToken, BaseURL: m.baseURI})
			if err != nil {
				t.Errorf("Error initializing client: %v", err)
			}

			image, err := CreateImage(c, tt.imageRef, bson.NewObjectId().Hex(), "No Description")

			if err != nil && !tt.expectError {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && tt.expectError {
				t.Errorf("Unexpected success. Expected error.")
			}
			if !reflect.DeepEqual(image, tt.expectImage) {
				t.Errorf("Got created collection %v - expected %v", image, tt.expectImage)
			}

			m.Stop()

		})

	}
}

func Test_SetTags(t *testing.T) {

	tests := []struct {
		description  string
		code         int
		reqCallback  func(*http.Request, *testing.T)
		containerRef string
		imageRef     string
		tags         []string
		expectError  bool
	}{
		{
			description:  "Valid Request",
			code:         http.StatusOK,
			containerRef: "test",
			imageRef:     bson.NewObjectId().Hex(),
			tags:         []string{"tag1", "tag2", "tag3"},
			expectError:  false,
		},
		{
			description:  "Error response",
			code:         http.StatusInternalServerError,
			containerRef: "test",
			imageRef:     bson.NewObjectId().Hex(),
			tags:         []string{"tag1", "tag2", "tag3"},
			expectError:  true,
		},
	}

	// Loop over test cases
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {

			m := mockService{
				t:           t,
				code:        tt.code,
				reqCallback: tt.reqCallback,
				httpPath:    "/v1/tags/" + tt.containerRef,
			}

			m.Run()
			defer m.Stop()

			c, err := NewClient(&Config{AuthToken: testToken, BaseURL: m.baseURI})
			if err != nil {
				t.Errorf("Error initializing client: %v", err)
			}

			err = SetTags(c, tt.containerRef, tt.imageRef, tt.tags)

			if err != nil && !tt.expectError {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && tt.expectError {
				t.Errorf("Unexpected success. Expected error.")
			}
		})
	}
}

func Test_Search(t *testing.T) {
	tests := []struct {
		description   string
		code          int
		body          interface{}
		reqCallback   func(*http.Request, *testing.T)
		value         string
		expectResults *SearchResults
		expectError   bool
	}{
		{
			description:   "ValidRequest",
			value:         "test",
			code:          http.StatusOK,
			body:          jsonresp.Response{Data: testSearch},
			expectResults: &testSearch,
			expectError:   false,
		},
		{
			description: "InternalServerError",
			value:       "test",
			code:        http.StatusInternalServerError,
			expectError: true,
		},
		{
			description: "BadRequest",
			value:       "test",
			code:        http.StatusBadRequest,
			expectError: true,
		},
	}

	// Loop over test cases
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {

			m := mockService{
				t:           t,
				code:        tt.code,
				body:        tt.body,
				reqCallback: tt.reqCallback,
				httpPath:    "/v1/search",
			}

			m.Run()
			defer m.httpServer.Close()

			c, err := NewClient(&Config{AuthToken: testToken, BaseURL: m.baseURI})
			if err != nil {
				t.Errorf("Error initializing client: %v", err)
			}

			results, err := Search(c, tt.value)

			if err != nil && !tt.expectError {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && tt.expectError {
				t.Errorf("Unexpected success. Expected error.")
			}
			if !reflect.DeepEqual(results, tt.expectResults) {
				t.Errorf("Got created collection %v - expected %v", results, tt.expectResults)
			}

			m.Stop()

		})

	}
}