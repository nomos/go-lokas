package dockerclient

import (
	"fmt"
	"testing"
)

func TestIntegrationServices(t *testing.T) {
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

	images, err := client.ListImages()
	if err != nil {
		panic(err)
	}
	for _, image := range images {
		fmt.Println("ID: ", image.ID)
		fmt.Println("Created: ", image.Created)
	}
}
