package dockerclient

import (
	"fmt"
	"testing"
)

func TestIntergrationServices(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	services, err := client.ListServices()
	if err != nil {
		panic(err)
	}
	for _, service := range services {
		fmt.Println("ID: ", service.ID)
		fmt.Println("Name: ", service.Spec.Name)
		fmt.Println("CreatedAt: ", service.CreatedAt)
	}
}
