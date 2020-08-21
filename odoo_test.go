package odoo

import "testing"

var config = &ClientConfig{
	Url:      "http://localhost:8069",
	Db:       "test",
	Username: "admin",
	Password: "admin",
}

func TestNewClient(t *testing.T) {
	client, err := NewClient(config)
	// Authenticate is implicitly tested via NewClient
	if err != nil {
		t.Fatal(err)
	}

	err = client.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Read(t *testing.T) {
	client, _ := NewClient(config)

	readResult, err := client.Read("res.users", []int64{client.uid}, []string{"login", "name"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(readResult)

	client.Close()
}

func TestClient_Search(t *testing.T) {
	client, _ := NewClient(config)

	searchResult, err := client.Search("res.users", NewDomain(Clause("active", "=", 1)))
	if err != nil {
		t.Fatal(err)
	}

	var uidFound = false
	for _, id := range searchResult {
		if id == client.uid {
			uidFound = true
			break
		}
	}
	if !uidFound {
		t.Fatalf("UID %d not found in results %d", client.uid, searchResult)
	}

	client.Close()
}

func TestClient_Create_And_Unlink(t *testing.T) {
	client, _ := NewClient(config)

	createResult, err := client.Create("ir.logging", map[string]interface{}{"type": "server", "name": "go-test", "path": "/", "line": "/", "func": "/", "message": "Test"})
	if err != nil {
		t.Fatal(err)
	}
	if createResult == 0 {
		t.Fatal("Create not successfully done")
	} else {
		t.Log(createResult)
	}

	unlinkResult, err := client.Unlink("ir.logging", []int64{createResult})
	if err != nil {
		t.Fatal(err)
	}
	if !unlinkResult {
		t.Fatal("unlinkResult not successfully done")
	}

	client.Close()
}

func TestClient_Write(t *testing.T) {
	client, _ := NewClient(config)

	writeResult, err := client.Write("res.users", []int64{client.uid}, map[string]interface{}{"signature": "Test"})
	if err != nil {
		t.Fatal(err)
	}
	if !writeResult {
		t.Fatal("Write not successfully done")
	}

	client.Close()
}
